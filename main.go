package main

import (
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/alecthomas/kong"
	"github.com/gnojus/keyring"
)

const description = `
nec is a command line tool for file sharing on Nextcloud.
It parses the existing configuration of the official desktop client
and operates on the folders of local filesystem,
while actually sharing the files that are synced with the server.
`

func main() {

	if len(os.Args) == 1 {
		os.Args = append(os.Args, "--help")
	}

	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	debug = slices.Contains(os.Args, "--debug")
	keyring.Debug = debug
	var cli cli
	k, err := kong.New(&cli,
		kong.ConfigureHelp(kong.HelpOptions{
			FlagsLast:           true,
			Compact:             true,
			NoExpandSubcommands: false,
			WrapUpperBound:      80,
		}),
		kong.Description(description),
	)
	if err != nil {
		panic(err)
	}

	for _, n := range k.Model.Children {
		if n.Name == "update" {
			n.Hidden = version == "" || repo == ""
		}
		for _, f := range n.Flags {
			if f.Name == "help" {
				f.Hidden = true
			}
		}
	}
	k.Model.HelpFlag.Hidden = true

	ctx, err := k.Parse(os.Args[1:])
	if err != nil {
		return err
	}
	return ctx.Run()
}

var debug = false

func debugf(format string, args ...any) {
	if debug {
		log.Printf("[nec] "+format, args...)
	}
}

type cli struct {
	Share   share   `kong:"cmd,aliases=s,help='share a local file'"`
	Unshare unshare `kong:"cmd,aliases=us,help='unshare a local file'"`
	List    list    `kong:"cmd,aliases=ls,help='list shares of local files'"`

	Update update `kong:"cmd,help='update nec using github releases if new version available'"`

	Debug bool `kong:"hidden"`
}
