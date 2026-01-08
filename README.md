# discotheque

Golang implementation of xklb/library

## Install

    go install github.com/chapmanjacobd/discotheque/cmd/disco@latest

## Usage

    $ disco /src/folder/ /dest/folder/
    ^C
    Interrupt received. Finishing source directory tree scan...
    Press Ctrl+C again in >2s to cancel and delete incomplete progress file

    Remaining paths saved to: folder.remainingfiles
    $ disco /src/folder/ /dest/folder/ --resume=folder.remainingfiles
    (repeat as many times as desired or wait to hit ENOSPC error)

## Help

    $ disco -h
    Usage: disco <source> <destination> [flags]

    Arguments:
    <source>         Source directory.
    <destination>    Destination directory.

    Flags:
    -h, --help           Show context-sensitive help.
    -r, --resume=FILE    Text file containing relative paths to copy.
