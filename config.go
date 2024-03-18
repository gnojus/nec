package main

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
	"github.com/gnojus/keyring"
	"gopkg.in/ini.v1"
)

func getCfgPath() (string, error) {
	return xdg.SearchConfigFile(filepath.Join("Nextcloud", "nextcloud.cfg"))
}

type folder struct {
	localPath  string
	targetPath string
}

func (f folder) String() string {
	return fmt.Sprintf("{%s -> %s}", f.localPath, f.targetPath)
}

type pathConfig struct {
	Path string `arg:"" help:"file on local filesystem"`
	account
	remoteFile string
	accounts   []account
}

// optPathConfig is like pathConfig, but the Path is optional.
// If Path is empty, it looks for a single account to match.
type optPathConfig struct {
	Path string `arg:"" optional:"" help:"file on local filesystem. Empty matches all files"`
	account
	remoteFile string
	accounts   []account
}

func (c *optPathConfig) AfterApply() error {
	return (*pathConfig)(c).loadHook(true)
}

func (c *pathConfig) AfterApply() (err error) {
	return c.loadHook(false)
}

func (c *pathConfig) loadHook(optionalPath bool) (err error) {
	accounts, folders, err := loadAccounts()
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}
	c.accounts = accounts
	if c.Path == "" {
		if !optionalPath {
			return errors.New("empty path")
		}
		return nil
	}
	c.Path, err = resolveFile(c.Path)
	if err != nil {
		return fmt.Errorf("resolving file: %w", err)
	}
	debugf("resolved file: %s", c.Path)

	for i := range accounts {
		for _, folder := range folders[i] {
			rel, _ := filepath.Rel(folder.localPath, c.Path)
			if !strings.HasPrefix(rel, "..") {
				c.account = accounts[i]

				c.remoteFile = path.Join(folder.targetPath, filepath.ToSlash(rel))
				debugf("folder %s matches: %s", folder, c.remoteFile)
				return nil
			}
			debugf("folder %s does not match", folder)
		}
	}

	return fmt.Errorf("file %s is not synced by any account", c.Path)
}

func resolveFile(fpath string) (string, error) {
	p, err := filepath.EvalSymlinks(fpath)
	if err != nil {
		return "", err
	}
	return filepath.Abs(p)
}

func loadAccounts() ([]account, [][]folder, error) {
	as, fs := []account{}, [][]folder{}
	p, err := getCfgPath()
	if err != nil {
		return nil, nil, fmt.Errorf("locating config file: %w", err)
	}
	cfg, err := ini.Load(p)
	if err != nil {
		return nil, nil, fmt.Errorf("reading config file: %w", err)
	}
	acc := cfg.Section("Accounts")
	for id, folderIDs := range findFolderIDs(acc.KeyStrings()) {
		account := account{
			url:  acc.Key(id + `\url`).String(),
			user: acc.Key(id + `\webflow_user`).String(),
		}
		if account.user == "" {
			account.user = acc.Key(id + `\dav_user`).String()
		}
		debugf("account %q for %q", account.user, account.url)
		if account.url == "" || account.user == "" {
			return nil, nil, fmt.Errorf("incomplete account information: %+v", account)
		}
		folders := []folder{}
		err := account.fetchPassword(id)
		if err != nil {
			return nil, nil, err
		}

		for _, fID := range folderIDs {
			fKey := id + `\` + fID + `\`
			f := folder{
				localPath:  filepath.Clean(acc.Key(fKey + "localPath").String()),
				targetPath: acc.Key(fKey + "targetPath").String(),
			}
			debugf("account folder %s", f)
			if f.localPath == "" || f.targetPath == "" {
				return nil, nil, fmt.Errorf("incomplete folder information: %+v", f)
			}
			folders = append(folders, f)
		}
		as = append(as, account)
		fs = append(fs, folders)
	}
	if as == nil {
		return nil, nil, errors.New("no account found with folder sync configured")
	}
	return as, fs, nil
}

func (a *account) fetchPassword(id string) error {
	var err error
	key := a.user + ":" + a.url
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}
	if id != "" {
		key += ":" + id
	}
	debugf(`reading password for "%s"`, key)
	a.pass, err = keyring.ReadPassword("Nextcloud", "nec", key)
	if err != nil {
		err = fmt.Errorf(`fetch password: %w`, err)
	}
	return err
}

var rFolder = regexp.MustCompile(`([0-9]+)\\((Multifolders|Folders|FoldersWithPlaceholders)\\[0-9]+)\\localPath`)

func findFolderIDs(keys []string) map[string][]string {
	debugf("config keys: %s", keys)
	f := make(map[string][]string)
	for _, key := range keys {
		m := rFolder.FindStringSubmatch(key)
		if m != nil {
			f[m[1]] = append(f[m[1]], m[2])
		}
	}

	return f
}
