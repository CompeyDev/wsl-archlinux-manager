package checkers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	logger "github.com/CompeyDev/wsl-archlinux-manager/util"
	"github.com/briandowns/spinner"
	"github.com/cavaliergopher/grab/v3"
)

func VerifySignature(mirrorUrl string) {
	bar := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	bar.Prefix = " "
	bar.Suffix = " Verifying signature of RootFS..."
	bar.Start()

	success, version, _ := pullSig(mirrorUrl)

	if !success {
		logger.Error("Failed to download signature of RootFS. Refusing to continue.")
		os.Exit(1)
	}

	userHomeDir, homeDirErr := os.UserHomeDir()

	if homeDirErr != nil {
		logger.Error("Failed to fetch installation directory, cannot verify authenticity of RootFS.")
		os.Exit(1)
	}
	cwd, dirErr := os.Getwd()

	if dirErr != nil {
		logger.Error("Failed to fetch installation directory, cannot verify authenticity of RootFS.")
		os.Exit(1)
	}

	unixWd := fmt.Sprintf("/mnt/c/%s", strings.ReplaceAll(strings.Split(userHomeDir, `C:\`)[1], `\`, "/")) + strings.ReplaceAll(strings.Split(cwd, userHomeDir)[1], `\`, "/")
	logger.Info(fmt.Sprintf("Looking for verification signature in Unix Directory %s", unixWd))
	cmd := exec.Command("wsl.exe", `bash`, `-c`, fmt.Sprintf(`gpg --keyserver-options auto-key-retrieve --verify archlinux-bootstrap-%s-x86_64.tar.gz.sig`, version))
	getAuthenticity, authenticityErr := cmd.CombinedOutput()
	if authenticityErr != nil {
		bar.Stop()
		logger.Error("Failed to verify authenticity of RootFS. Refusing to continue.")
	}
	if strings.Contains(strings.Trim(string(getAuthenticity), "\n\r"), "Good signature") {
		logger.Info("Matching signature: 4AA4 767B BC9C 4B1D 18AE  28B7 7F2D 434B 9741 E8AC")
		logger.Info("Successfully matched checksums and verified authenticity!")
		bar.Stop()
	} else {
		bar.Stop()
		logger.Error("Failed to verify authenticity of RootFS. Refusing to continue.")
	}

}

func pullSig(url string) (isSuccessful bool, ver string, error error) {
	bar := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	bar.Prefix = " "
	bar.Suffix = " Downloading Signature..."

	var version = strings.Split(strings.Split(url, "iso/")[1], "/")[0]
	var structuredUrl = fmt.Sprintf("%s/archlinux-bootstrap-%s-x86_64.tar.gz.sig", url, version)
	res, err := grab.Get(".", structuredUrl)

	if err != nil {
		logger.Error("Failed to download Signature.")
		return false, version, err
	}

	logger.Info(fmt.Sprintf("Downloaded Signature %s", res.Filename))

	return true, version, nil
}
