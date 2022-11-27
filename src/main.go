package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/CompeyDev/wsl-archlinux-manager/lib/core"
	logger "github.com/CompeyDev/wsl-archlinux-manager/util"
	"github.com/briandowns/spinner"
	"github.com/gookit/color"
)

func main() {
	if runtime.GOOS == "windows" {
		checks()
		core.Build()
	} else {
		fmt.Println("WSL is reserved for windows users only.")
		os.Exit(1)
	}
}

func checks() {
	color.Blueln("======> Pre-installation checks")
	fmt.Println("🔃 Running checks...")
	_, permsErr := os.Open("\\\\.\\PHYSICALDRIVE0")

	if permsErr != nil {
		fmt.Println("Please run this command with elevated priveleges.")
		os.Exit(1)
	}

	bar := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	bar.Prefix = " "
	bar.Suffix = " Checking for WSL..."
	bar.Start()

	getAvailability, availabilityErr := exec.Command("powershell.exe", "(Get-WindowsOptionalFeature -Online -FeatureName *Subsystem*).State").Output()

	if availabilityErr != nil {
		logger.Error("Failed to check for WSL availability.")
	}

	if strings.Trim(string(getAvailability), "\n\r") == "Enabled" {
		logger.Info("WSL is enabled.")
		bar.Stop()

		bar.Prefix = " "
		bar.Suffix = " Checking for existing WSL distributions..."

		bar.Start()

		preInstalledDistro, installedDistroErr := exec.Command("powershell.exe", "wsl").Output()

		if installedDistroErr != nil {
			logger.Error("Failed to check for preinstalled distributions.")
			bar.Stop()
		}

		if strings.Contains(string(preInstalledDistro), "no installed distributions") {
			logger.Error("Preinstalled distributions do not exist. (Please make sure the default WSL distribution is Debian-based)")
			bar.Stop()
		}

	}

	userHomeDir, homeDirErr := os.UserHomeDir()

	bar.Suffix = " Creating install directory..."
	bar.Start()
	time.Sleep(1 * time.Second)
	if homeDirErr != nil {
		logger.Error("Failed to initialize installation directory.")
	}

	installDir := path.Join(userHomeDir, ".wslm")
	archDir := path.Join(installDir, "arch")

	os.Mkdir(installDir, fs.FileMode(os.O_RDWR))
	os.Mkdir(archDir, fs.FileMode(os.O_RDWR))
	bar.Stop()
	logger.Info("Successfully initialized installation directory.")
	logger.Progress("Initialized WSLm.")
	color.Blueln("===============================")
}
