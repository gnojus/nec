package main

import (
	"errors"
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

func isSubfile(dir, f string) bool {
	return dir == "/" || dir == f || strings.HasPrefix(f, dir+"/")
}

func (l *list) AfterApply() error {
	if l.Path == "" && l.Recursive {
		return errors.New("omitted path matches all shares, --recursive (-r) is unnecessary")
	}
	return nil
}

func (l *list) Run() error {
	accounts := l.accounts
	if l.remoteFile != "" {
		accounts = []account{l.account}
	}
	for _, acc := range accounts {
		if len(accounts) > 1 {
			fmt.Printf("%s on %s:\n", acc.user, acc.url)
		}
		shares, err := l.loadShares(acc)
		if err != nil {
			return err
		}
		l.writeShares(shares)
	}
	return nil
}

func (l *list) writeShares(shares []sharedFile) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)
	for _, e := range shares {
		if l.Recursive {
			if !isSubfile(l.remoteFile, e.Path) {
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
	w.Flush()
}

func (l *list) Help() string {
	return `
Lists all shares with their data ([path], id, url, expiry date) on local file or folder.
If recursive, it prints the paths, relative to the one passed as argument.
When no file is supplied, all shares of single account are printed with full server paths.`
}

// loadShares loads and returns all shared files that matches path from command
// line. Empty path matches shared files.
func (l *list) loadShares(acc account) ([]sharedFile, error) {
	v := url.Values{}

	// recursive query needs to fetch the whole list
	if !l.Recursive && l.remoteFile != "" {
		v.Set("path", l.remoteFile)
	}
	data, err := request[shares](acc, "GET", v)
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
