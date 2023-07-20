# goenv (Golang Version Manager)
goenv is a command-line tool to manage multiple versions of Golang on your system.

## Installation
**Download**: https://github.com/ankitcharolia/goenv/releases

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

## License
This project is licensed under the MIT License - see the LICENSE file for details.