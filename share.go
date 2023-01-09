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

	Clipboard bool   `kong:"help='copy share url into systems clipboard',short=c"`
	Expire    string `kong:"help='expire date of this share, truncated to days'"`
	Note      string `kong:"help='note displayed together with the shared file'"`
	With      string `kong:"help='user to share with',xor=to"`
	WithGroup string `kong:"help='group to share with',xor=to"`
	Email     string `kong:"help='send email with shared file to this email',xor=to"`
	Edit      bool   `kong:"help='edit permission on the shared file'"`
	Upload    bool   `kong:"help='allow uploads to folder using public link'"`
}

func fmtExpiry(expiry string) string {
	return strings.TrimSuffix(expiry, " 00:00:00")
}

func (s *share) Help() string {
	return `
Shares a file/folder identified by local path.
By default it's a public share and url with the shared file is printed to stdout.
`
}

func (s *share) Run() error {
	v := url.Values{}
	v.Set("shareType", "3") // public link
	v.Set("path", s.remoteFile)
	v.Set("note", s.Note)

	if s.Email != "" {
		v.Set("shareType", "4") // email
		v.Set("shareWith", s.Email)
	}

	if s.With != "" {
		v.Set("shareType", "0") // user
		v.Set("shareWith", s.With)
	}

	if s.WithGroup != "" {
		v.Set("shareType", "1") // user
		v.Set("shareWith", s.WithGroup)
	}

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

	if s.Edit {
		v.Set("permissions", "15")
	}
	if s.Upload {
		v.Set("publicUpload", "true")
	}

	data, err := request[sharedFile](s.account, "POST", v)
	if err != nil {
		return err
	}
	if data.Expiration != "" {
		fmt.Fprintln(os.Stderr, "share expires on:", fmtExpiry(data.Expiration))
	}

	if s.With == "" && s.WithGroup == "" {
		fmt.Println(data.fmtShareLink(s.url))
	}
	if s.Clipboard {
		return clipboard.WriteAll(data.URL)
	}
	return nil
}

type sharedFile struct {
	Text                 string `xml:",chardata"`
	ID                   string `xml:"id"`
	ShareType            int    `xml:"share_type"`
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

func (s *sharedFile) fmtShareLink(serverUrl string) string {
	if s.URL != "" || s.Token == "" {
		return s.URL
	}
	// TODO: this may not work on servers without rewrite rules
	p, _ := url.JoinPath(serverUrl, "s", s.Token)
	return p
}

func (s *sharedFile) fmtNote() string {
	note := s.Note
	suffix := ""
	switch s.ShareType {
	case 0:
		suffix = "user"
	case 1:
		suffix = "group"
	case 4:
		suffix = "email"
	}

	if suffix != "" {
		if note != "" {
			note += ", "
		}
		note = note + suffix + ": " + s.ShareWith
	}
	return note
}
