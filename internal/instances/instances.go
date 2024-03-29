package instances

import (
	"io"
	"net"
	"errors"
	"strconv"
	"github.com/demostanis/gimmeasearx/internal/grade"
	"github.com/hashicorp/go-version"
	"io/ioutil"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"regexp"
)

const USER_AGENT =  "Mozilla/5.0 (gimmeasearx; https://github.com/demostanis/gimmeasearx; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.63 Safari/537.36 I am nice, I promise."

// The data fetched from searx.space.
type InstancesData struct {
	Instances map[string]Instance `json:"instances"`
}
// Struct representing an instance
// in the data fetched.
type Instance struct {
	Comments []string `json:"comments"`
	NetworkType *string `json:"network_type"`
	Error *string `json:"error,omit_empty"`
	Version *string `json:"version"`
	Html *struct {
		Resources struct {} `json:"ressources"`
		Grade string `json:"grade"`
	} `json:"html,omit_empty"`
}

// Creates an InstancesData from the fetched JSON data.
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

// Checks whether tor is listening on :9050 or :9150.
func TorListening() (int, error) {
	_, err := net.Dial("tcp", ":9050")
	if err != nil {
		_, err2 := net.Dial("tcp", ":9150")
		if err2 != nil {
			return 0, errors.New("Tor is not listening")
		} else {
			return 9150, nil
		}
	} else {
		return 9050, nil
	}
}

// Verifies an instance, using tor or not, by checking if it
// returns expected results from searches. It also removes
// instances using Cloudflare.
func Verify(instanceUrl string, instance Instance) bool {
	result := false
	useTor := false
	// We need other tests
	tests := map[string][]string{
		"south+park": []string{"Trey Parker", "Matt Stone"},
		"gimmeasearx": []string{"Find a random searx instance"},
	}
	port, err := TorListening()
	if strings.HasSuffix(instanceUrl, ".onion/") && err == nil {
		useTor = true
	}
	for search, matches := range tests {
		var resp *http.Response
		var err error
		if useTor {
			req, _ := http.NewRequest("GET", instanceUrl + "search?q=" + search, nil)
			req.Header.Set("User-Agent", USER_AGENT)
			req.Header.Set("Accept-Language", "en-US,en;q=0.5")

			tr := &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse("socks5://127.0.0.1:" + strconv.Itoa(port))
				},
			}
			client := &http.Client{Transport: tr}
			resp, err = client.Do(req)
		} else {
			req, _ := http.NewRequest("GET", instanceUrl + "search?q=" + search, nil)
			// These headers are mostly to circumvent filtron.
			// Please don't hate me for bypassing your anti rooboots.
			req.Header.Set("User-Agent", USER_AGENT)
			req.Header.Set("Accept-Language", "en-US,en;q=0.5")

			client := &http.Client{}
			resp, err = client.Do(req)
		}
		if err == nil && resp != nil {
			if resp.Header.Get("server") == "cloudflare" {
				result = false
				continue
			}
			page, _ := ioutil.ReadAll(resp.Body)
			for _, regex := range matches {
				r := regexp.MustCompile(regex)
				result = r.MatchString(string(page))

				// Do not make any other test if one fails
				if !result {
					return result
				}
			}
			resp.Body.Close()
		}
	}
	return result
}

// Fetches data from searx.space.
func Fetch() (*InstancesData, error) {
	resp, err := http.Get("https://searx.space/data/instances.json")
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, err
	}
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

// Finds a random instance between the ones fetched according
// to user's choosen options.
func FindRandomInstance(fetchedInstances *map[string]Instance, gradesEnabled []string, blacklist []string, torEnabled bool, torOnlyEnabled bool, minVersion version.Version, customInstances []string) (*string, bool) {
	keys := *new([]string)
	LOOP: for key, instance := range *fetchedInstances {
		if instance.Error == nil && instance.Version != nil {
			if !containsGrade(gradesEnabled, (*instance.Html).Grade) {
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

			version, err := version.NewVersion(*instance.Version)
			if err == nil && minVersion.GreaterThan(version) {
				continue LOOP
			}

			if torEnabled && *instance.NetworkType == "tor" {
				keys = append(keys, key)
			} else if !torOnlyEnabled && *instance.NetworkType != "tor" {
				keys = append(keys, key)
			}
		}
	}
	for _, customInstance := range customInstances {
		keys = append(keys, customInstance)
	}

	if len(keys) < 1 {
		return nil, false
	}
	randInt := rand.Intn(len(keys))
	randUrl := keys[randInt]

	isCustom := false
	for _, customInstance := range customInstances {
		if randUrl == customInstance {
			isCustom = true
			break
		}
	}

	return &randUrl, isCustom
}

