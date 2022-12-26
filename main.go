package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

func main() {
	var cli cli
	err := kong.Parse(&cli).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

type cli struct {
	Share   share   `cmd:"" aliases:"s" help:"share a local file"`
	Unshare unshare `cmd:"" aliases:"us" help:"unshare a local file"`
	List    list    `cmd:"" aliases:"ls" help:"list shares of local files"`
}
