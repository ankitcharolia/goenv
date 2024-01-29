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
	"os/exec"
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
			return
		} else {
			fmt.Println("Installed Golang versions:")
			for _, version := range versions {
				fmt.Println(version)
			}
		}
		return
	}

	if *uninstallVersion != "" {
		uninstallGoVersion(*uninstallVersion)
		return
	}

	if *remoteList {
		listRemoteVersions()
	} else if *installVersion != "" {
		installGoVersion(*installVersion)
	} else if *useVersion != "" {
		useGoVersion(*useVersion)
		return
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

	// Print the versions in reverse order
	for i := len(versions) - 1; i >= 0; i-- {
		version := strings.TrimPrefix(versions[i].Version, "go")
		fmt.Println(version)
	}
}

func installGoVersion(version string) {
	activeVersion := getCurrentGoVersion()
	if isInstalled(version) || version == activeVersion {
		greenColor := "\033[32m"
		resetColor := "\033[0m"
		fmt.Printf("%sGo version %s is already installed.%s\n", greenColor, version, resetColor)
		os.Exit(0)
	}

	// Install the specific Go version
	greenColor := "\033[32m"
	resetColor := "\033[0m"
	fmt.Printf("%sInstalling Go version %s...%s\n", greenColor, version, resetColor)

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
	if err := extractAndCopy(reader, installPath); err != nil {
		log.Fatalf("Failed to extract and copy: %v", err)
	}

	bar.Finish()
	fmt.Printf("%sGo version %s is installed at %s.\nTo make this your default version, run 'goenv --use %s'%s\n", greenColor, version, installPath, version, resetColor)
}

// extractAndCopy extracts the contents of a tar.gz archive to the specified directory.
func extractAndCopy(reader io.Reader, destination string) error {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	baseDir := ""

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Skip if the header is nil or empty
		if header == nil || header.Name == "" {
			continue
		}

		// Handle the case where the archive contains a single top-level directory
		if baseDir == "" {
			baseDir = filepath.Dir(header.Name)
		}

		// Construct the destination path
		destPath := filepath.Join(destination, strings.TrimPrefix(header.Name, baseDir))

		// Create directories if needed
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(destPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
			continue
		}

		// Create the parent directory if needed
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Create and copy the file
		file, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
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

	// Check if the .go directory exists
	_, err := os.Stat(installPath)
	if err != nil {
		// If the .go directory doesn't exist, create it and return an empty list
		if os.IsNotExist(err) {
			err = os.Mkdir(installPath, os.ModePerm)
			if err != nil {
				log.Fatalf("Failed to create .go directory: %v", err)
			}
			return []string{} // Return an empty list to indicate no Golang versions are installed
		}
		log.Fatalf("Failed to read directory: %v", err)
	}

	activeVersion := getCurrentGoVersion()
	fileInfos, err := os.ReadDir(installPath)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	versions := []string{}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			version := fileInfo.Name()
			if version == activeVersion {
				version = "* " + version + "  (Currently active GOLANG version)" // Mark the active version with an asterisk
			} else {
				version = "  " + version
			}
			versions = append(versions, version)
		}
	}

	return versions
}

// getCurrentGoVersion returns the actively used Go version by running "go env GOROOT" command.
func getCurrentGoVersion() string {
	cmd := exec.Command("go", "env", "GOROOT")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	goRoot := strings.TrimSpace(string(output))
	version := strings.TrimPrefix(goRoot, filepath.Join(os.Getenv("HOME"), ".go")+"/")
	return version
}

// uninstallGoVersion uninstalls a specific Golang version.
func uninstallGoVersion(version string) {
	if isInstalled(version) {
		if version == getCurrentGoVersion() {
			// If the version to be uninstalled is the currently active version,
			// switch to another installed version before uninstalling it.
			versions := listInstalledVersions()
			for _, v := range versions {
				v = strings.TrimSpace(strings.TrimPrefix(v, "* "))
				if v != version {
					useGoVersion(v)
					break
				}
			}
			fmt.Printf("Switched to Go version %s (currently active) before uninstalling.\n", version)
		}

		installPath := filepath.Join(os.Getenv("HOME"), ".go", version)
		err := os.RemoveAll(installPath)
		if err != nil {
			log.Fatalf("Failed to uninstall Go version: %v", err)
		}

		fmt.Printf("Go version %s has been uninstalled.\n", version)
	} else {
		if version == getCurrentGoVersion() {
			fmt.Printf("Cannot uninstall Go version %s because it is currently active.\n", version)
			fmt.Printf("To uninstall, please switch to another installed version first.\n")
		} else {
			fmt.Printf("Go version %s is not installed. Please install it first.\n", version)
		}
	}
}

// isInstalled checks if a specific Golang version is already installed.
func isInstalled(version string) bool {
	versions := listInstalledVersions()
	for _, v := range versions {
		v = strings.TrimSpace(strings.TrimPrefix(v, "* "))
		if v == version || v == "* "+version {
			return true
		}
	}
	return false
}

// useGoVersion sets the specified Go version as the active version to use.
func useGoVersion(version string) {

	// Get the currently active Go version
	activeVersion := getCurrentGoVersion()
	// Check if the specified version is the same as the currently active version
	if version == activeVersion {
		fmt.Printf("Go version %s is already the active version.\n", version)
		return
	} else {
		// Check if the specified Go version is installed
		if !isInstalled(version) {
			fmt.Printf("Go version %s is not installed. Please install it first.\n", version)
			return
		}
	}

	// Get the installation path for the specified Go version
	goPath := filepath.Join(os.Getenv("HOME"), ".go", version)

	// Update the environment variables to point to the specified Go version
	os.Setenv("GOPATH", goPath)

	// Update the PATH variable to include the Go binaries for the specified version
	binPath := filepath.Join(goPath, "bin")
	newPath := fmt.Sprintf("%s%c$PATH", binPath, os.PathListSeparator)
	os.Setenv("PATH", newPath)

	// ANSI escape code for red color
	redColor := "\033[31m"
	// ANSI escape code to reset color to default
	resetColor := "\033[0m"

	// Update the Go version in the appropriate shell configuration file
	shell := getShell()
	if shell == "bash" {
		updateGoVersionInShellRC(version, ".bashrc")
		message := fmt.Sprintf("Using Go version %s.%s\nPlease make sure to execute: source ~/.bashrc\n%s", version, redColor, resetColor)
		fmt.Print(message)
	} else if shell == "zsh" {
		updateGoVersionInShellRC(version, ".zshrc")
		message := fmt.Sprintf("Using Go version %s.%s\nPlease make sure to execute: source ~/.zshrc\n%s", version, redColor, resetColor)
		fmt.Print(message)
	}
}

// getShell returns the user's default shell
func getShell() string {
	shell := os.Getenv("SHELL")
	shellParts := strings.Split(shell, "/")
	return strings.TrimSpace(shellParts[len(shellParts)-1])
}

// updateGoVersionInShellRC updates the Go version in the specified shell RC file.
func updateGoVersionInShellRC(version string, rcFile string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user's home directory: %v", err)
	}

	rcFilePath := filepath.Join(homeDir, rcFile)
	rcData, err := os.ReadFile(rcFilePath)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", rcFilePath, err)
	}

	// Remove the old "export PATH=$HOME/.go" line
	lines := strings.Split(string(rcData), "\n")
	var newLines []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "export PATH=$HOME/.go") {
			newLines = append(newLines, line)
		}
	}

	// Add the new "export PATH=$HOME/.go/<version>/go/bin:$PATH" line
	placeholder := fmt.Sprintf("export PATH=$HOME/.go/%s/bin:$PATH", version)
	newLines = append(newLines, placeholder)

	// Join the lines and remove leading/trailing empty lines
	newRCData := strings.Join(newLines, "\n")
	newRCData = strings.Trim(newRCData, "\n") // Remove trailing empty line, if any

	err = os.WriteFile(rcFilePath, []byte(newRCData), 0644)
	if err != nil {
		log.Fatalf("Failed to write to %s: %v", rcFilePath, err)
	}
}
