package updater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Change struct {
	ShortSHA string
	Message  string
}

type Updates struct {
	From       string
	To         string
	Message    string
	Updateable bool
	Changes    []Change
}

type releases []release
type release struct {
	HTMLURL string `json:"html_url"`
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Draft   bool   `json:"draft"`
	Body    string `json:"body"`
}

func getJSON(url string, target interface{}) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %v", err)
	}

	err = json.Unmarshal(b, &target)
	if err != nil {
		return 0, fmt.Errorf("could not parse json: %v", err)
	}

	return resp.StatusCode, nil
}

func getInfo() (*releases, error) {
	r := &releases{}
	_, err := getJSON("https://api.github.com/repos/geek1011/BookBrowser/releases", r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func CheckForUpdates(version string) (*Updates, error) {
	u := &Updates{}

	if strings.HasPrefix(version, "dev") {
		u.From = "dev"
		u.To = "dev"
		u.Message = "You are using a development version of BookBrowser, therefore BookBrowser cannot check for updates."
		u.Updateable = false
		u.Changes = nil
		return u, nil
	}

	r, err := getInfo()
	if err != nil {
		return nil, err
	}

	rv := *r

	if len(rv) == 0 {
		return nil, fmt.Errorf("unknown error checking for updates")
	}

	u.From = version
	u.To = rv[0].TagName
	if u.From == u.To {
		u.Message = "You are using the latest version of BookBrowser."
		u.Updateable = false
		return u, nil
	}

	rvs := []release{}
	in := false
	for _, rvr := range rv {
		if in {
			rvs = append(rvs, rvr)
		}

		if rvr.TagName == u.From {
			in = true
		}
	}

	if !in {
		return nil, fmt.Errorf("unknown error checking for updates")
	}

	u.Changes = []Change{}
	for _, rvsr := range rvs {
		for _, l := range strings.Split(rvsr.Body, "\n") {
			if len(l) < 9 {
				continue
			}

			// detect a short sha
			if !strings.Contains(l[7:8], " ") || strings.Contains(l[0:7], " ") {
				continue
			}

			sl := strings.SplitN(l, " ", 2)

			u.Changes = append(u.Changes, Change{
				ShortSHA: sl[0],
				Message:  sl[1],
			})
		}
	}

	u.Message = fmt.Sprintf("An update is available from %s to %s.", u.From, u.To)
	u.Updateable = true

	return u, nil
}
