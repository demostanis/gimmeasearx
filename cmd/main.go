package main

import (
	"fmt"
	"net/http"
	"net/url"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/demostanis/gimmeasearx/pkg/grade"
	"github.com/demostanis/gimmeasearx/pkg/instances"
	"github.com/demostanis/gimmeasearx/pkg/version"
	"github.com/hashicorp/go-version"
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
		go func() {
			for key, instance := range *fetchedInstances {
				if instances.VerifyInstance(key, instance) {
					delete(*fetchedInstances, key)
				}
			}
		}()
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
	minVersion := params.minVersion

	randUrl := instances.FindRandomInstance(fetchedInstances, gradesEnabled, blacklist, torEnabled, torOnlyEnabled, minVersion)
	if randUrl == nil {
		return c.Render(http.StatusExpectationFailed, "index.html", map[string]bool{
			"Error": true,
		})
	}

	if fetchedInstances != nil {
		return c.Redirect(http.StatusFound, *randUrl + "?preferences=" + url.QueryEscape(*preferences) + "&q=" + url.QueryEscape(c.QueryParam("q")))
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
	minVersion := params.minVersion
	latestVersion := params.latestVersion

	data := map[string]interface{}{
		"CurrentUrl": c.Request().URL.RequestURI(),
		"OptionsSelected": map[string]interface{}{
			"Tor": torEnabled,
			"TorOnly": torOnlyEnabled,
			"Blacklist": blacklist,
			"Latest": latestVersion,
		},
		"Grades": grade.Grades(),
		"GradesSelected": gradesEnabled,
		"Preferences": preferences,
		"MinVersion": minVersion.Original(),
	}

	if fetchedInstances != nil {
		randUrl := instances.FindRandomInstance(fetchedInstances, gradesEnabled, blacklist, torEnabled, torOnlyEnabled, minVersion)
		if randUrl == nil {
			data["Error"] = true
			return c.Render(http.StatusExpectationFailed, "index.html", data)
		}
		randInstance := (*fetchedInstances)[*randUrl]

		data["Instance"] = randInstance
		data["InstanceUrl"] = randUrl
		data["GradeComment"] = grade.Comment(randInstance.Html.Grade)

		return c.Render(http.StatusOK, "index.html", data)
	} else {
		data["Error"] = true
		return c.Render(http.StatusTooEarly, "index.html", data)
	}
}

type Params struct {
	torOnlyEnabled bool
	torEnabled bool
	gradesEnabled []string
	blacklist []string
	preferences *string
	minVersion version.Version
	latestVersion bool
}

func parseParams(c echo.Context) Params {
	torOnlyEnabled := c.QueryParam("toronly") == "on"
	torEnabled := torOnlyEnabled || c.QueryParam("tor") == "on"
	latestVersion := c.QueryParam("latestversion") == "on"
	minVersion, _ := version.NewVersion("0.0.0")
	if !latestVersion {
		r, err := version.NewVersion(c.QueryParam("minversion"))
		if err != nil {
			minVersion, _ = version.NewVersion("0.0.0")
		} else {
			minVersion = r
		}
	} else {
		minVersion, _ = version.NewVersion(findlatestversion.Searx())
	}
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
		*minVersion,
		latestVersion,
	}
}

