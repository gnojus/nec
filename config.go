package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
	"github.com/gnojus/keyring"
	"gopkg.in/ini.v1"
)

func openConfig() (*os.File, error) {
	p, err := xdg.SearchConfigFile(filepath.Join("Nextcloud", "nextcloud.cfg"))
	if err != nil {
		return nil, fmt.Errorf("locating config file: %w", err)
	}
	return os.Open(p)
}

type pathDetails struct {
	remoteFile string
	account
	localPath  string
	targetPath string
}

type pathConfig struct {
	Path string `arg:"" help:"file on local filesystem"`
	pathDetails
}

// optPathConfig is like pathConfig, but the Path is optional.
// If Path is empty, it looks for a single account to match.
type optPathConfig struct {
	Path string `arg:"" optional:"" help:"file on local filesystem. Empty matches all files"`
	pathDetails
}

func (c *optPathConfig) AfterApply() error {
	return (*pathConfig)(c).loadHook(true)
}

func (c *pathConfig) AfterApply() (err error) {
	return c.loadHook(false)
}

func (c *pathConfig) loadHook(opt bool) (err error) {
	err = c.readConfig(opt)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}
	if c.Path != "" {
		rel, err := filepath.Rel(c.localPath, c.Path)
		if err != nil {
			return fmt.Errorf("making path: %w", err)
		}
		c.remoteFile = path.Join(c.targetPath, filepath.ToSlash(rel))
	}

	return nil
}

func (c *pathConfig) readConfig(opt bool) error {
	if c.Path != "" {
		p, err := filepath.EvalSymlinks(c.Path)
		if err != nil {
			return err
		}
		p, err = filepath.Abs(p)
		if err != nil {
			return err
		}
		c.Path = p
	} else if !opt {
		return errors.New("path is empty")
	}
	return c.read(opt)
}

func (c *pathConfig) read(opt bool) error {
	f, err := openConfig()
	if err != nil {
		return err
	}
	defer f.Close()
	cfg, err := ini.Load(f)
	if err != nil {
		return err
	}

	acc := cfg.Section("Accounts")
	if acc.Name() == "" {
		return fmt.Errorf("no Accounts section in config")
	}

	ids := findFolders(acc.KeyStrings())
	for id, folders := range ids {
		c.url, c.user = acc.Key(id+`\url`).String(), acc.Key(id+`\dav_user`).String()

		// return just user data if path is empty and optional
		if c.Path == "" && opt {
			// TODO: somehow make this better
			if len(ids) > 1 {
				return fmt.Errorf("ambiguous empty path (matches more than one account)")
			}

			return c.fetchPassword(id)
		}

		for _, f := range folders {
			folder := id + `\Folders\` + f + `\`
			lpath, err := acc.GetKey(folder + "localPath")
			if err != nil {
				return err
			}
			tpath, err := acc.GetKey(folder + "targetPath")
			if err != nil {
				return err
			}

			c.localPath = filepath.Clean(lpath.String())
			if c.localPath != "" && strings.HasPrefix(c.Path, c.localPath) {
				c.targetPath = tpath.String()
				return c.fetchPassword(id)
			}
		}
	}

	return fmt.Errorf("no matching account found")
}

func (c *pathConfig) fetchPassword(id string) error {
	var err error
	key := fmt.Sprintf("%s:%s/:%s", c.user, c.url, id)
	c.pass, err = keyring.ReadPassword("Nextcloud", "nec", key)
	return err
}

var rFolder = regexp.MustCompile(`([0-9]+)\\Folders\\([0-9]+)\\localPath`)

func findFolders(keys []string) map[string][]string {
	f := make(map[string][]string)
	for _, key := range keys {
		m := rFolder.FindStringSubmatch(key)
		if m != nil {
			f[m[1]] = append(f[m[1]], m[2])
		}
	}

	return f
}
