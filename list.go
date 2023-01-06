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

	Recursive bool `kong:"short=r,help='recursively descend into folders'"`
}

func (l *list) Run() error {
	shares, err := l.loadShares()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)
	for _, e := range shares {
		if l.Recursive {
			if !strings.HasPrefix(e.Path, l.remoteFile) {
				continue
			}

			relRemote := filepath.FromSlash(strings.TrimPrefix(e.Path, l.remoteFile))
			sep := ""
			if e.ItemType == "folder" {
				sep = string(filepath.Separator)
			}
			fmt.Fprint(w, filepath.Join(l.Path, relRemote), sep, "\t")
		} else if l.remoteFile == "" {
			sep := ""
			if e.ItemType == "folder" {
				sep = "/"
			}
			fmt.Fprint(w, e.Path, sep, "\t")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.ID, e.fmtShareLink(l.url), fmtExpiry(e.Expiration), e.fmtNote())
	}
	return w.Flush()
}

func (l *list) Help() string {
	return `
Lists all shares with their data ([path], id, url, expiry date) on local file or folder.
If recursive, it prints the paths, relative to the one passed as argument.
When no file is supplied, all shares of single account are printed with full server paths.`
}

// loadShares loads and returns all shared files that matches path from command
// line. Empty path matches shared files.
func (l *list) loadShares() ([]sharedFile, error) {
	v := url.Values{}

	// recursive query needs to fetch the whole list
	if !l.Recursive && l.remoteFile != "" {
		v.Set("path", l.remoteFile)
	}
	data, err := request[shares](&l.account, "GET", v)
	if err != nil {
		return nil, fmt.Errorf("loading shares: %w", err)
	}

	pos := 0
	for i, e := range data.Elements {
		// filter out non matching recursive elements
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
