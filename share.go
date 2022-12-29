package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/olebedev/when"
)

type share struct {
	pathConfig

	Clipboard bool   `short:"c" help:"copy share url into system's clipboard"`
	Expire    string `optional:"" help:"expire date of this share"`
}

func fmtExpiry(expiry string) string {
	return strings.TrimSuffix(expiry, " 00:00:00")
}

func (s *share) Run() error {
	v := url.Values{}
	v.Set("shareType", "3") // public link
	v.Set("path", s.remoteFile)

	if s.Expire != "" {
		res, err := when.EN.Parse(s.Expire, time.Now())
		if err != nil {
			return fmt.Errorf("parsing time: %w", err)
		}
		if res == nil {
			return errors.New("failed to parse expire time")
		}
		v.Set("expireDate", res.Time.Format(time.RFC3339))
	}

	data, err := request[sharedFile](&s.account, "POST", v)
	if err != nil {
		return err
	}
	if data.Expiration != "" {
		fmt.Fprintln(os.Stderr, "share expires on:", fmtExpiry(data.Expiration))
	}

	fmt.Println(data.URL)
	if s.Clipboard {
		return clipboard.WriteAll(data.URL)
	}
	return nil
}

type sharedFile struct {
	Text                 string `xml:",chardata"`
	ID                   string `xml:"id"`
	ShareType            string `xml:"share_type"`
	UidOwner             string `xml:"uid_owner"`
	DisplaynameOwner     string `xml:"displayname_owner"`
	Permissions          string `xml:"permissions"`
	CanEdit              string `xml:"can_edit"`
	CanDelete            string `xml:"can_delete"`
	Stime                string `xml:"stime"`
	Parent               string `xml:"parent"`
	Expiration           string `xml:"expiration"`
	Token                string `xml:"token"`
	UidFileOwner         string `xml:"uid_file_owner"`
	Note                 string `xml:"note"`
	Label                string `xml:"label"`
	DisplaynameFileOwner string `xml:"displayname_file_owner"`
	Path                 string `xml:"path"`
	ItemType             string `xml:"item_type"`
	Mimetype             string `xml:"mimetype"`
	HasPreview           string `xml:"has_preview"`
	StorageID            string `xml:"storage_id"`
	Storage              string `xml:"storage"`
	ItemSource           string `xml:"item_source"`
	FileSource           string `xml:"file_source"`
	FileParent           string `xml:"file_parent"`
	FileTarget           string `xml:"file_target"`
	ShareWith            string `xml:"share_with"`
	ShareWithDisplayname string `xml:"share_with_displayname"`
	Password             string `xml:"password"`
	SendPasswordByTalk   string `xml:"send_password_by_talk"`
	URL                  string `xml:"url"`
	MailSend             string `xml:"mail_send"`
	HideDownload         string `xml:"hide_download"`
	Attributes           string `xml:"attributes"`
}
