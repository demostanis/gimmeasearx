package instances

import (
	"io"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"net/http"
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
	Error string `json:"error,omit_empty"`
	Version string `json:"version"`
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

