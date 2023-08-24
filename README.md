[![Release Version](https://img.shields.io/github/v/release/ankitcharolia/goenv?label=goenv)](https://github.com/ankitcharolia/goenv/releases/latest)
![Build CI](https://github.com/ankitcharolia/goenv/actions/workflows/build-publish.yaml/badge.svg)
![CodeQL](https://github.com/ankitcharolia/goenv/actions/workflows/codeql-analysis.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ankitcharolia/goenv)](https://goreportcard.com/report/github.com/ankitcharolia/goenv)
[![License](https://img.shields.io/badge/License-MIT%20-blue.svg)](https://github.com/ankitcharolia/goenv/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/downloads/ankitcharolia/goenv/total.svg)]()

# goenv (Golang Version Manager)
goenv is a command-line tool to manage multiple versions of Golang on your system.

## Installation
**Releases**: https://github.com/ankitcharolia/goenv/releases

**LINUX**
```shell
wget -O - https://github.com/ankitcharolia/goenv/releases/latest/download/goenv-linux-amd64.tar.gz | tar -xz -C ~/.go
export PATH=$HOME/.go:$PATH >> ~/.bashrc
source ~/.bashrc
```

**MAC**
```shell
wget -O - https://github.com/ankitcharolia/goenv/releases/latest/download/goenv-darwin-amd64.tar.gz | tar -xz -C ~/.go
export PATH=$HOME/.go:$PATH >> ~/.zshrc
source ~/.zshrc
```

## Usage
goenv provides several commands to manage Golang versions on your system.
```shell
$ goenv --help
Usage: goenv [flags] [<options>]
Flags:
  --help          goenv help command
  --install       Install a specific version of GOLANG
  --list          List all installed GOLANG versions
  --list-remote   List all remote versions of GOLANG
  --uninstall     Uninstall a specific version of GOLANG
  --use           Use a specific version of GOLANG
```

### List all remote versions of GOLANG
```shell
$ goenv --list-remote
1.21rc3
1.21rc2
1.20.6
1.20.5
1.20.4
1.20.3
1.20.2
1.20.1
...
```

### Install a specific version of GOLANG
```shell
$ goenv --install 1.20.6
Installing Go version 1.20.6...
https://dl.google.com/go/go1.20.6.linux-amd64.tar.gz
95.57 MiB / 95.57 MiB [-----------------------------------------------------------------------------------------------] 100.00% 13.92 MiB p/s 7.1s
Go version 1.20.6 is installed at $HOME_DIRECTORY/.go/1.20.6.
```

### Use a specific version of GOLANG
```shell
$ goenv --use 1.20.6
Using Go version 1.20.6.
Please make sure to execute: source ~/.bashrc
```

### List all installed GOLANG versions
```shell
$ goenv --list
Installed Golang versions:
* 1.20.6  (Currently active GOLANG version)
  1.20.5
  1.20.4
```

## Supported Shell
* Bash
* Zsh

## Support
Feel free to create an Issue/Pull Request if you find any bug with `goenv`