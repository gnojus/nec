package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

type list struct {
	optPathConfig

	Recursive bool `short:"r" help:"print the shares recursively on folder"`
}

func (l *list) Run() error {
	shares, err := l.loadShares()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)
	for _, e := range shares {
		if l.Recursive || l.remoteFile == "" {
			if !strings.HasPrefix(e.Path, l.remoteFile) {
				continue
			}

			relRemote := strings.TrimPrefix(e.Path, l.remoteFile)
			fmt.Fprintf(w, "%s\t", filepath.Join(l.Path, filepath.FromSlash(relRemote)))
		}

		fmt.Fprintf(w, "%s\t%s\t", e.ID, e.URL)
		if e.Expiration != "" {
			fmt.Fprintf(w, "\t%s", e.Expiration)
		}
		fmt.Fprintln(w)
	}
	return w.Flush()
}

func (l *list) loadShares() ([]sharedFile, error) {
	v := url.Values{}
	if !l.Recursive && l.remoteFile != "" {
		v.Set("path", l.remoteFile)
	}
	data, err := request[shares](&l.account, "GET", v)
	if err != nil {
		return nil, fmt.Errorf("loading shares: %w", err)
	}

	pos := 0
	for i, e := range data.Elements {
		if l.Recursive && !strings.HasPrefix(e.Path, l.remoteFile) {
			continue
		}
		data.Elements[pos] = data.Elements[i]
		pos++
	}
	return data.Elements[:pos], nil
}

type shares struct {
	Elements []sharedFile `xml:"element"`
}
