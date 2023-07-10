package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type GoVersion struct {
	Version string `json:"version"`
}

func main() {
	remoteList := flag.Bool("list-remote", false, "List all remote versions of GOLANG")
	installVersion := flag.String("install", "", "Install a specific version of GOLANG")
	flag.Parse()

	flag.Usage = func() {
		fmt.Printf("Usage: %s [flags] [<options>]\n", "goenv")
		flag.PrintDefaults()
	}

	if *remoteList {
		url := "https://go.dev/dl/?mode=json&include=all"

		response, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to fetch versions: %v", err)
		}
		defer response.Body.Close()

		var versions []GoVersion
		err = json.NewDecoder(response.Body).Decode(&versions)
		if err != nil {
			log.Fatalf("Failed to decode JSON response: %v", err)
		}

		for _, v := range versions {
			version := strings.TrimPrefix(v.Version, "go")
			fmt.Println(version)
		}
	} else if *installVersion != "" {
		// Install the specific Go version
		fmt.Printf("Installing Go version %s...\n", *installVersion)

		// Download the Go distribution archive
		url := fmt.Sprintf("https://dl.google.com/go/go%s.%s-%s.tar.gz", *installVersion, runtime.GOOS, runtime.GOARCH)
		fmt.Println(url)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to download Go version: %v", err)
		}
		defer resp.Body.Close()

		// Extract the archive to the desired installation location
		installPath := filepath.Join(os.Getenv("HOME"), "go"+*installVersion)
		err = extractTarGz(resp.Body, installPath)
		if err != nil {
			log.Fatalf("Failed to extract Go version: %v", err)
		}

		fmt.Printf("Go version %s is installed at %s.\n", *installVersion, installPath)
	} else {
		flag.Usage()
	}
}

// extractTarGz extracts the contents of a tar.gz archive to the specified directory.
func extractTarGz(src io.Reader, dest string) error {
	err := os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return err
	}

	err = exec.Command("tar", "-xzf", "-", "-C", dest).Run()
	if err != nil {
		return err
	}

	return nil
}
