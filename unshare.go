package main

import (
	"errors"
	"fmt"
)

type unshare struct {
	list

	ID string `kong:"help='id of share to remove'"`
}

func (u *unshare) Run() error {
	if u.ID == "" && u.remoteFile == "" {
		return errors.New("not removing all shares")
	}
	shares, err := u.loadShares()
	if err != nil {
		return err
	}
	for _, s := range shares {
		if u.ID == "" || s.ID == u.ID {
			_, err := request[struct{}](&u.account, "DELETE", nil, s.ID)
			if err != nil {
				return fmt.Errorf("removing share: %w", err)
			}

		}
	}
	return nil
}
