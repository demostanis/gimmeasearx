package instances

import (
	"io"
	"fmt"
	"github.com/demostanis/gimmeasearx/pkg/grade"
	"io/ioutil"
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"regexp"
)

type NotOKError struct {
	status string
}

func (err *NotOKError) Error() string {
	return err.status
}

type FetchError struct {
	Reason error
	Url string
}

func (err *FetchError) Error() string {
	return fmt.Sprintf("Failed to fetch resource at URL %s: %s",
		err.Url, err.Reason)
}

type InstancesData struct {
	Instances map[string]Instance `json:"instances"`
}
type Instance struct {
	Comments []string `json:"comments"`
	NetworkType string `json:"network_type"`
	Error *string `json:"error,omit_empty"`
	Version *string `json:"version"`
	Html struct {
		Resources struct {} `json:"ressources"`
		Grade string `json:"grade"`
	} `json:"html,omit_empty"`
}

func InstancesNew(data io.ReadCloser) (*InstancesData, error) {
	var instances InstancesData
	resp, err := ioutil.ReadAll(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp, &instances); err != nil {
		return nil, err
	}
	return &instances, nil
}

func VerifyInstance(url string, instance Instance) bool {
	result := false
	// We need other tests
	tests := map[string][]string{
		"south+park": []string{"Trey Parker", "Matt Stone"},
	}
	for search, matches := range tests {
		resp, err := http.Get(url + "/search?q=" + search)
		if err == nil && resp != nil {
			page, _ := ioutil.ReadAll(resp.Body)
			for _, regex := range matches {
				r := regexp.MustCompile(regex)
				result = r.MatchString(string(page))
			}
			resp.Body.Close()
		}
	}
	return result
}

var InstancesApiUrl = "https://searx.space/data/instances.json"

func Fetch() (*InstancesData, error) {
	resp, err := http.Get(InstancesApiUrl)
	if err != nil {
		return nil, &FetchError{
			err,
			InstancesApiUrl,
		}
	}
	if resp.StatusCode != 200 {
		return nil, &FetchError{
			&NotOKError{resp.Status},
			InstancesApiUrl,
		}
	}
	defer resp.Body.Close()
	instances, err := InstancesNew(resp.Body)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func containsGrade(arr []string, elem string) bool {
	for _, a := range arr {
		if grade.Symbol(a) == elem {
			return true
		}
	}
	return false
}

func FindRandomInstance(fetchedInstances *map[string]Instance, gradesEnabled []string, blacklist []string, torEnabled bool, torOnlyEnabled bool) *string {
	keys := *new([]string)
	LOOP: for key, instance := range *fetchedInstances {
		if instance.Error == nil && instance.Version != nil {
			if !containsGrade(gradesEnabled, instance.Html.Grade) {
				continue LOOP
			}

			for _, blacklisted := range blacklist {
				if len(strings.TrimSpace(blacklisted)) < 1 {
					continue
				}
				if r, err := regexp.Compile(blacklisted); err == nil && r.MatchString(key) {
					continue LOOP
				}
			}

			if torEnabled && instance.NetworkType == "tor" {
				keys = append(keys, key)
			} else if !torOnlyEnabled && instance.NetworkType != "tor" {
				keys = append(keys, key)
			}
		}
	}

	if len(keys) < 1 {
		return nil
	}
	randInt := rand.Intn(len(keys))
	randUrl := keys[randInt]

	return &randUrl
}
