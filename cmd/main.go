package main

import (
	"fmt"
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/demostanis/gimmeasearx/pkg/instances"
	"html/template"
	"math/rand"
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

	e.Logger.Fatal(e.Start(":8080"))
}

func index(c echo.Context) error {
	torOnlyEnabled := c.QueryParam("toronly") == "on"
	torEnabled := torOnlyEnabled || c.QueryParam("tor") == "on"

	if fetchedInstances != nil {
		keys := *new([]string)
		for key, instance := range *fetchedInstances {
			if torEnabled && instance.NetworkType == "tor" {
				keys = append(keys, key)
			} else if !torOnlyEnabled {
				keys = append(keys, key)
			}
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
			},
		})
	} else {
		return c.Render(http.StatusTooEarly, "index.html", map[string]bool{
			"Error": true,
		})
	}
}

