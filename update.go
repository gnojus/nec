package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"

	selfupdate "github.com/staktrace/go-update"
)

var version = ""
var repo = ""

type update struct{}

func (update) Run() error {
	if version == "" || repo == "" {
		return fmt.Errorf("update not configured")
	}
	fmt.Fprintf(os.Stderr, "checking for new version in github.com/%s/releases\n", repo)
	fmt.Fprintf(os.Stderr, "current version: %s\n", version)

	file := "nec-" + runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		file += ".exe"
	}
	reg := regexp.MustCompile(repo + "/releases/download/([^/]+)/" + file)
	latest := ""
	c := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			m := reg.FindStringSubmatch(r.URL.Path)
			if m != nil {
				latest = m[1]
			}
			return nil
		},
	}

	req, err := c.Get("https://github.com/" + repo + "/releases/latest/download/" + file)
	if err != nil {
		return fmt.Errorf("fetching latest release: %w", err)
	}
	defer req.Body.Close()
	fmt.Fprintf(os.Stdout, "latest version: %s\n", latest)

	if version == latest {
		return nil
	}
	err = selfupdate.Apply(req.Body, selfupdate.Options{})
	if err != nil {
		return fmt.Errorf("updating binary: %w", err)
	}

	fmt.Fprintln(os.Stderr, "successfully updated to", latest)
	return nil
}
