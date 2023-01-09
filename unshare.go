package main

import (
	"errors"
	"fmt"
)

type unshare struct {
	list

	ID string `kong:"help='id of share to remove'"`
}

func (u *unshare) Help() string {
	return `
Removes all the shares of a given file.
Can be narrowed down to single share with the --id flag.
`
}

func (u *unshare) Run() error {
	if u.ID == "" && u.remoteFile == "" {
		return errors.New("not removing all shares")
	}
	if u.remoteFile == "" {
		if len(u.accounts) > 1 {
			return errors.New("ambiguous empty path (matches more than one account)")
		}
		u.account = u.accounts[0]
	}
	shares, err := u.loadShares(u.account)
	if err != nil {
		return err
	}
	removed := false
	for _, s := range shares {
		if u.ID == "" || s.ID == u.ID {
			_, err := request[struct{}](u.account, "DELETE", nil, s.ID)
			if err != nil {
				return fmt.Errorf("removing share: %w", err)
			}
			removed = true
		}
	}
	if !removed {
		return fmt.Errorf("no shares matched to remove")
	}
	return nil
}
