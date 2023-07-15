package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
)

type GoVersion struct {
	Version string `json:"version"`
}

func main() {
	remoteList := flag.Bool("list-remote", false, "List all remote versions of GOLANG")
	installVersion := flag.String("install", "", "Install a specific version of GOLANG")
	listInstalled := flag.Bool("list", false, "List all installed GOLANG versions")
	uninstallVersion := flag.String("uninstall", "", "Uninstall a specific version of GOLANG")
	useVersion := flag.String("use", "", "Use a specific version of GOLANG")
	help := flag.Bool("help", false, "goenv help command")
	flag.Parse()

	flag.Usage = func() {
		fmt.Printf("Usage: %s [flags] [<options>]\n", "goenv")
		fmt.Println("Flags:")
		flag.VisitAll(func(f *flag.Flag) {
			padding := strings.Repeat(" ", 14-len(f.Name))
			fmt.Printf("  --%s%s%s\n", f.Name, padding, f.Usage)
		})
	}

	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.Usage = flag.Usage

	if len(flag.Args()) > 0 {
		fmt.Println("Invalid command provided.")
		flag.Usage()
		return
	}

	if *listInstalled {
		versions := listInstalledVersions()
		if len(versions) == 0 {
			fmt.Println("No installed Golang versions found.")
		} else {
			fmt.Println("Installed Golang versions:")
			for _, version := range versions {
				fmt.Println(version)
			}
		}
		return
	}

	if *uninstallVersion != "" {
		if !isInstalled(*uninstallVersion) {
			fmt.Printf("Go version %s is not installed.\n", *uninstallVersion)
			return
		}
		uninstallGoVersion(*uninstallVersion)
		return
	}

	if *remoteList {
		listRemoteVersions()
	} else if *installVersion != "" {
		installGoVersion(*installVersion)
	} else if *useVersion != "" {
		useGoVersion(*useVersion)
	} else if *help {
		flag.Usage()
	} else {
		flag.Usage()
	}
}

func listRemoteVersions() {
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
}

func installGoVersion(version string) {
	if isInstalled(version) {
		fmt.Printf("Go version %s is already installed.\n", version)
		return
	}

	// Install the specific Go version
	fmt.Printf("Installing Go version %s...\n", version)

	// Download the Go distribution archive
	url := fmt.Sprintf("https://dl.google.com/go/go%s.%s-%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to download Go version: %v", err)
	}
	defer resp.Body.Close()

	// Get the content length to set progress bar
	contentLength := resp.ContentLength

	// Create the progress bar
	bar := pb.Full.Start64(contentLength)
	bar.Set(pb.Bytes, true)
	bar.SetRefreshRate(time.Millisecond * 10)

	// Create a proxy reader to track the progress
	reader := bar.NewProxyReader(resp.Body)

	// Extract the archive to the desired installation location
	installPath := filepath.Join(os.Getenv("HOME"), ".go", version)
	err = extractTarGz(reader, installPath)
	if err != nil {
		log.Fatalf("Failed to extract Go version: %v", err)
	}

	bar.Finish()
	fmt.Printf("Go version %s is installed at %s.\n", version, installPath)
}

// extractTarGz extracts the contents of a tar.gz archive to the specified directory.
func extractTarGz(src io.Reader, dest string) error {
	gzr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(dest, header.Name)

		if header.Typeflag == tar.TypeDir {
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, tr)
		if err != nil {
			return err
		}
	}

	return nil
}

// listInstalledVersions lists all installed Golang versions.
func listInstalledVersions() []string {
	installPath := filepath.Join(os.Getenv("HOME"), ".go")
	fileInfos, err := os.ReadDir(installPath)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	versions := []string{}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			versions = append(versions, fileInfo.Name())
		}
	}

	return versions
}

// uninstallGoVersion uninstalls a specific Golang version.
func uninstallGoVersion(version string) {
	installPath := filepath.Join(os.Getenv("HOME"), ".go", version)
	err := os.RemoveAll(installPath)
	if err != nil {
		log.Fatalf("Failed to uninstall Go version: %v", err)
	}

	fmt.Printf("Go version %s has been uninstalled.\n", version)
}

// isInstalled checks if a specific Golang version is already installed.
func isInstalled(version string) bool {
	versions := listInstalledVersions()
	for _, v := range versions {
		if v == version {
			return true
		}
	}
	return false
}

// useGoVersion sets the specified Go version as the active version to use.
func useGoVersion(version string) {

	// Get the installation path for the specified Go version
	goPath := filepath.Join(os.Getenv("HOME"), ".go", version, "go")

	// Update the environment variables to point to the specified Go version
	os.Setenv("GOPATH", goPath)

	// Update the PATH variable to include the Go binaries for the specified version
	binPath := filepath.Join(goPath, "bin")
	newPath := fmt.Sprintf("%s%c$PATH", binPath, os.PathListSeparator)
	os.Setenv("PATH", newPath)

	// Update the Go version in the ~/.bashrc file
	updateGoVersionInBashrc(version)

	fmt.Printf("Using Go version %s.\n", version)
}

// updateGoVersionInBashrc updates the Go version in the ~/.bashrc file.
func updateGoVersionInBashrc(version string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user's home directory: %v", err)
	}

	bashrcPath := filepath.Join(homeDir, ".bashrc")
	bashrcData, err := os.ReadFile(bashrcPath)
	if err != nil {
		log.Fatalf("Failed to read ~/.bashrc: %v", err)
	}

	// Remove the old "export PATH=$HOME/.go" line
	lines := strings.Split(string(bashrcData), "\n")
	var newLines []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "export PATH=$HOME/.go") {
			newLines = append(newLines, line)
		}
	}

	// Add the new "export PATH=$HOME/.go/<version>/go/bin:$PATH" line
	placeholder := fmt.Sprintf("export PATH=$HOME/.go/%s/go/bin:$PATH", version)
	newLines = append(newLines, placeholder)

	newBashrcData := strings.Join(newLines, "\n")

	err = os.WriteFile(bashrcPath, []byte(newBashrcData), 0644)
	if err != nil {
		log.Fatalf("Failed to write to ~/.bashrc: %v", err)
	}
}
