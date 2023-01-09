nec - Nextcloud command line client
===================================
`nec` is a command line tool for [Nextcloud](https://nextcloud.com/), primary
for sharing files. It's made to be cross-platform, tested on Mac OS, Windows
and Linux, on Linux working at least on KDE with kwallet and other libsecret
backends.

Installation
------------
Download prebuilt binaries on [releases](https://github.com/gnojus/nec/releases)
or build with `go install github.com/gnojus/nec@latest`. Note that CGO
(`CGO_ENABLED=1`) is required to build with keychain support for Mac OS.

The Github releases also feature self-updating capabilities via `update` command
(since `v0.0.11`).
When building from source updates can be enabled using linker flags:
`-ldflags "-X main.version=<version> -X main.repo=gnojus/nec"`.

Usage
-----
nec is intended to be zero-configuration tool. This means that it works alongside
official [Nextcloud desktop client](https://github.com/nextcloud/desktop/). It parses
the existing configuration and operates on the folders of local filesystem, synced
with server. 

This tool does not upload files, it only shares ones that are synced with the server.
Therefore if you want to share a file that is not synced, you may want to move/copy
it to a folder that is synchronized and share it from there. The upload will be
performed by the desktop client.

### Commands
Most nec commands take local path as an argument, but the actually affected file
(shared or unshared) is the one on the server, synced to the local one by desktop
client.

    $ nec --help
    Usage: nec <command>

    nec is a command line tool for file sharing on Nextcloud. It parses the existing
    configuration of the official desktop client and operates on the folders of
    local filesystem, while actually sharing the files that are synced with the
    server.

    Commands:
      share (s)       share a local file
      unshare (us)    unshare a local file
      list (ls)       list shares of local files
      update          update nec using github releases if new version available

    Run "nec <command> --help" for more information on a command.

### Example
    $ ls
    api.go  config.go  go.mod  go.sum  LICENSE  list.go  main.go  README.md  share.go  unshare.go
    $ nec s README.md
    https://cloud.example.com/s/NB8LiLSgqpSmPxW
    $ nec share README.md --expire 'in 1 week' --note 'nec readme'
    share expires on: 2023-01-13
    https://cloud.example.com/s/DHHLgYxjNmJDsCr
    $ nec ls README.md
    65  https://cloud.example.com/s/NB8LiLSgqpSmPxW
    66  https://cloud.example.com/s/DHHLgYxjNmJDsCr  2023-01-13  nec readme
    $ nec ls -r .
    README.md  65  https://cloud.example.com/s/NB8LiLSgqpSmPxW
    README.md  66  https://cloud.example.com/s/DHHLgYxjNmJDsCr  2023-01-13  nec readme
    $ nec us --id 65
    $ nec ls -r .
    README.md  66  https://cloud.example.com/s/DHHLgYxjNmJDsCr  2023-01-13  nec readme

Issues
------
The tool is in the early stage of developent, but should get the job done.
Please report any bugs or suggestions on the [Github issue tracker](https://github.com/gnojus/nec/issues).

