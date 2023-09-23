# Nohns CLI

A CLI with commands for automating simple but repetitive time consuming tasks.

## Getting started

First and foremost you must have Go installed on your system. Thereafter, you can install the binary by running `go install github.com/nohns/nohns-cli/cmd/nohns@latest`

Now the `nohns` command should be available in your terminal.

## Compatibility

All commands are to be used on Unix like system such as macOS or Linux.
Some commands are meant to only be run on Linux while some are cross platform. You will be warned when running the command.

## Secrets

A .nohns directory will be created at the users directory which will be used to store secrets. These are in plain text
