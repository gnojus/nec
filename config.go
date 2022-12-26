package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/gnojus/keyring"
	"gopkg.in/ini.v1"
)

func openConfig() (*os.File, error) {
	return os.Open(filepath.Join(xdg.ConfigHome, "Nextcloud", "nextcloud.cfg"))
}

type pathDetails struct {
	remoteFile string
	account
	localPath  string
	targetPath string
}

type pathConfig struct {
	Path string `arg:"" help:"file on local filesystem to be shared"`
	pathDetails
}

type optPathConfig struct {
	Path string `arg:"" optional:"" help:"file on local filesystem to be shared"`
	pathDetails
}

func (c *optPathConfig) AfterApply() error {
	err := (*pathConfig)(c).readCheck(true)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}
	return nil
}

func (c *pathConfig) AfterApply() (err error) {
	err = c.readCheck(false)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}
	rel, err := filepath.Rel(c.localPath, c.Path)
	if err != nil {
		return fmt.Errorf("making path: %w", err)
	}
	c.remoteFile = path.Join(c.targetPath, filepath.ToSlash(rel))

	return nil
}

func (c *pathConfig) readCheck(opt bool) error {
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

	f, err := ini.Load(filepath.Join(xdg.ConfigHome, "Nextcloud", "nextcloud.cfg"))
	if err != nil {
		return err
	}
	acc := f.Section("Accounts")
	if acc.Name() == "" {
		return fmt.Errorf("no Accounts section in config")
	}

	for i := 0; ; i++ {
		id := strconv.Itoa(i) + `\`
		url, user := acc.Key(id+"url").String(), acc.Key(id+"dav_user").String()
		if url == "" || user == "" {
			if i == 1 && opt {
				return c.fetchPassword(0)
			}
			return fmt.Errorf("no matching account found")
		}
		c.url = url
		c.user = user
		for j := 1; ; j++ {
			folder := id + `Folders\` + strconv.Itoa(j) + `\`
			if !acc.HasKey(folder + "localPath") {
				break
			}

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
				return c.fetchPassword(i)
			}
		}
	}

}

func (c *pathConfig) fetchPassword(id int) error {
	var err error
	key := fmt.Sprintf("%s:%s/:%d", c.user, c.url, id)
	c.pass, err = keyring.ReadPassword("Nextcloud", "nec", key)
	return err
}
