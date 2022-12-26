package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
)

type account struct {
	user, pass, url string
}

func request[T any](s *account, method string, v url.Values, sub ...string) (T, error) {
	var r response[T]
	sub = append([]string{"ocs/v2.php/apps/files_sharing/api/v1/shares"}, sub...)
	URL, err := url.JoinPath(s.url, sub...)
	if err != nil {
		return r.Data, fmt.Errorf("creating share url: %w", err)
	}

	req, err := http.NewRequest(method, URL+"?"+v.Encode(), nil)
	if err != nil {
		return r.Data, fmt.Errorf("creating request: %w", err)
	}

	req.SetBasicAuth(s.user, s.pass)
	req.Header.Add("OCS-APIRequest", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return r.Data, fmt.Errorf("calling api: %w", err)
	}
	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return r.Data, fmt.Errorf("decoding response: %w", err)
	}
	if r.Meta.Statuscode != 200 {
		return r.Data, fmt.Errorf("api request: %s (%d): %s", r.Meta.Status, r.Meta.Statuscode, r.Meta.Message)
	}

	return r.Data, nil
}

type response[T any] struct {
	XMLName xml.Name `xml:"ocs"`
	Text    string   `xml:",chardata"`
	Meta    struct {
		Text       string `xml:",chardata"`
		Status     string `xml:"status"`
		Statuscode int    `xml:"statuscode"`
		Message    string `xml:"message"`
	} `xml:"meta"`
	Data T `xml:"data"`
}
