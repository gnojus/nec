package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

const description = `
nec is a command line tool for Nextcloud
`

func main() {
	var cli cli

	if len(os.Args) == 1 {
		os.Args = append(os.Args, "--help")
	}

	k := kong.Parse(&cli,
		kong.ConfigureHelp(kong.HelpOptions{
			FlagsLast:           true,
			Compact:             true,
			NoExpandSubcommands: false,
		}),
		kong.Description(description),
	)
	debug = cli.Debug

	err := k.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

var debug = false

type cli struct {
	Share   share   `kong:"cmd,aliases=s,help='share a local file'"`
	Unshare unshare `cmd:"" aliases:"us" help:"unshare a local file"`
	List    list    `cmd:"" aliases:"ls" help:"list shares of local files"`

	Debug bool `kong:"hidden"`
}
