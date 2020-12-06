package main

import (
	"fmt"
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/demostanis/gimmeasearx/pkg/grade"
	"github.com/demostanis/gimmeasearx/pkg/instances"
	"html/template"
	"strings"
	"time"
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

	var fetch func()
	fetch = func() {
		resp, err := instances.Fetch();
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		fetchedInstances = &resp.Instances
	}

	fetch()
	go func() {
		for range time.Tick(time.Hour * 24) {
			fetch()
		}
	}()

	e.Use(middleware.Gzip())
	e.Use(middleware.Recover())

	e.GET("/", index)
	e.GET("/search", search)

	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = ":8080"
	}
	e.Logger.Fatal(e.Start(port))
}

func search(c echo.Context) error {
	params := parseParams(c)
	torOnlyEnabled := params.torOnlyEnabled
	torEnabled := params.torEnabled
	gradesEnabled := params.gradesEnabled
	blacklist := params.blacklist
	preferences := params.preferences

	randUrl := instances.FindRandomInstance(fetchedInstances, gradesEnabled, blacklist, torEnabled, torOnlyEnabled)
	if randUrl == nil {
		return c.Render(http.StatusExpectationFailed, "index.html", map[string]bool{
			"Error": true,
		})
	}

	if fetchedInstances != nil {
		return c.Redirect(http.StatusMovedPermanently, *randUrl + "?preferences=" + *preferences + "&q=" + c.QueryParam("q"))
	} else {
		return c.String(http.StatusTooEarly, "No instances available. Please try again in a few seconds.")
	}
}

func index(c echo.Context) error {
	params := parseParams(c)
	torOnlyEnabled := params.torOnlyEnabled
	torEnabled := params.torEnabled
	gradesEnabled := params.gradesEnabled
	blacklist := params.blacklist
	preferences := params.preferences

	if fetchedInstances != nil {
		randUrl := instances.FindRandomInstance(fetchedInstances, gradesEnabled, blacklist, torEnabled, torOnlyEnabled)
		if randUrl == nil {
			return c.Render(http.StatusExpectationFailed, "index.html", map[string]bool{
				"Error": true,
			})
		}
		randInstance := (*fetchedInstances)[*randUrl]

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
			"Grades": grade.Grades(),
			"GradesSelected": gradesEnabled,
			"Preferences": preferences,
		})
	} else {
		return c.Render(http.StatusTooEarly, "index.html", map[string]bool{
			"Error": true,
		})
	}
}

type Params struct {
	torOnlyEnabled bool
	torEnabled bool
	gradesEnabled []string
	blacklist []string
	preferences *string
}

func parseParams(c echo.Context) Params {
	torOnlyEnabled := c.QueryParam("toronly") == "on"
	torEnabled := torOnlyEnabled || c.QueryParam("tor") == "on"
	gradesEnabled := *new([]string)
	for _, thisGrade := range grade.Grades() {
		if c.QueryParam(thisGrade["Id"].(string)) == "on" {
			gradesEnabled = append(gradesEnabled, thisGrade["Id"].(string))
		}
	}
	if len(gradesEnabled) < 1 {
		gradesEnabled = grade.Defaults()
	}
	blacklist := *new([]string)

	if b := c.QueryParam("blacklist"); len(b) > 0 {
		for _, s := range strings.Split(b, ";") {
			blacklist = append(blacklist, strings.TrimSpace(s))
		}
	}
	preferences := c.QueryParam("preferences")

	return Params{
		torEnabled,
		torOnlyEnabled,
		gradesEnabled,
		blacklist,
		&preferences,
	}
}

