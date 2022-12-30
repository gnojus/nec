nec - Nextcloud command line client
===================================
`nec` is a command line tool for [Nextcloud](https://nextcloud.com/), primary
for sharing files. At this point in time it's very experimental and not well tested.

Installation
------------
Download prebuilt binaries on [releases](https://github.com/gnojus/nec/releases)
or build with `go install github.com/gnojus/nec@latest`.

Usage
-----
nec is intended to be zero-configuration tool. This means that it works alongside
official [Nextcloud desktop client](https://github.com/nextcloud/desktop/). It parses
the existing configuration and operates on the folders of local filesystem, synced
with server. nec is made to be cross-platform, tested on Mac OS, Windows and linux.
Linux works at least on KDE with kwallet and other libsecret backends.


### Commands
Most nec commands take local path as an argument, but the actually affected file
(shared or unshared) is the one on the server, synced to the local one by desktop
client.

    nec --help
    Usage: nec <command>

    nec is a command line tool for Nextcloud

    Commands:
      share (s)       share a local file
      unshare (us)    unshare a local file
      list (ls)       list shares of local files

    Run "nec <command> --help" for more information on a command.

