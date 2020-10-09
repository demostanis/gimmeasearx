package main

import (
	"fmt"
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/demostanis/gimmeasearx/pkg/grade"
	"github.com/demostanis/gimmeasearx/pkg/instances"
	"html/template"
	"math/rand"
	"regexp"
	"strings"
	"io"
	"os"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var t = &Template{
	templates: template.Must(template.ParseGlob("templates/*.html")),
}

var fetchedInstances *map[string]instances.Instance = nil

func main() {
	e := echo.New()
	e.Renderer = t

	resp, err := instances.Fetch();
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	fetchedInstances = &resp.Instances

	e.Use(middleware.Gzip())
	e.Use(middleware.Recover())

	e.GET("/", index)

	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = ":8080"
	}
	e.Logger.Fatal(e.Start(port))
}

func index(c echo.Context) error {
	torOnlyEnabled := c.QueryParam("toronly") == "on"
	torEnabled := torOnlyEnabled || c.QueryParam("tor") == "on"
	blacklist := *new([]string)

	if b := c.QueryParam("blacklist"); len(b) > 0 {
		for _, s := range strings.Split(b, ";") {
			blacklist = append(blacklist, strings.TrimSpace(s))
		}
	}

	if fetchedInstances != nil {
		keys := *new([]string)
		for key, instance := range *fetchedInstances {
			if instance.Error == nil && instance.Version != nil {
				stop := false
				for _, blacklisted := range blacklist {
					if len(strings.TrimSpace(blacklisted)) < 1 {
						continue
					}
					if r, err := regexp.Compile(blacklisted); err == nil && r.MatchString(key) {
						stop = true
					}
				}
				if !stop {
					if torEnabled && instance.NetworkType == "tor" {
						keys = append(keys, key)
					} else if !torOnlyEnabled && instance.NetworkType != "tor" {
						keys = append(keys, key)
					}
				}
			}
		}

		if len(keys) < 1 {
			return c.Render(http.StatusExpectationFailed, "index.html", map[string]bool{
				"Error": true,
			})
		}
		randInt := rand.Intn(len(keys))
		randUrl := keys[randInt]
		randInstance := (*fetchedInstances)[randUrl]

		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"CurrentUrl": c.Request().URL.RequestURI(),
			"Instance": randInstance,
			"InstanceUrl": randUrl,
			"OptionsSelected": map[string]interface{}{
				"Tor": torEnabled,
				"TorOnly": torOnlyEnabled,
				"Blacklist": blacklist,
			},
			"GradeComment": grade.Comment(randInstance.Html.Grade),
		})
	} else {
		return c.Render(http.StatusTooEarly, "index.html", map[string]bool{
			"Error": true,
		})
	}
}

