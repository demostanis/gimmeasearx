package findlatestversion

import (
	"net/http"
	"regexp"
	"io/ioutil"
)

// Sucky regex, but it works
var r = regexp.MustCompile("tag/v(.*)\"")

func Searx() string {
	resp, err := http.Get("https://github.com/searx/searx/releases")
	if err != nil {
		// In case the request to Github fails,
		// fallback to current version. Should 
		// it error instead?
		return "0.18.0"
	}
	page, _ := ioutil.ReadAll(resp.Body)
	result := r.FindStringSubmatch(string(page))
	return result[1]
}

