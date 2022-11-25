package checkers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/cavaliergopher/grab/v3"
	"github.com/gookit/color"
)

// TODO:
func VerifySignature(mirrorUrl string) {
	bar := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	bar.Prefix = " "
	bar.Suffix = " Verifying signature of RootFS..."
	bar.Start()

	success, _ := pullSig(mirrorUrl)

	if !success {
		color.Red.Println("\r    ❎ Failed to download signature of RootFS. Refusing to continue.")
		os.Exit(1)
	}

	userHomeDir, homeDirErr := os.UserHomeDir()

	if homeDirErr != nil {
		color.Red.Println("\r    ❎ Failed to fetch installation directory, cannot verify authenticity of RootFS.")
		os.Exit(1)
	}
	cwd, dirErr := os.Getwd()

	if dirErr != nil {
		color.Red.Println("\r    ❎ Failed to fetch installation directory, cannot verify authenticity of RootFS.")
		os.Exit(1)
	}

	unixWd := fmt.Sprintf("/mnt/c/%s", strings.ReplaceAll(strings.Split(userHomeDir, `C:\`)[1], `\`, "/")) + strings.ReplaceAll(strings.Split(cwd, userHomeDir)[1], `\`, "/")

	// does not work?

	getAuthenticity, authenticityErr := exec.Command("powershell.exe", fmt.Sprintf(`wsl bash -c "cd %s && gpg --keyserver-options auto-key-retrieve --verify archlinux-bootstrap-2022.11.01-x86_64.tar.gz.sig"`, unixWd)).Output()

	if authenticityErr != nil {
		color.Red.Println("\r    ❎ Failed to verify authenticity of RootFS. Refusing to continue.")
		os.Exit(1)
	}

	color.Green.Println("\r    ✅ Successfully matched checksums and verified authenticity!")
	bar.Stop()

	fmt.Println(strings.Trim(string(getAuthenticity), "\n\r"))
}

func pullSig(url string) (isSuccessful bool, error error) {
	bar := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	bar.Prefix = " "
	bar.Suffix = " Downloading Signature..."

	var version = strings.Split(strings.Split(url, "iso/")[1], "/")[0]
	var structuredUrl = fmt.Sprintf("%s/archlinux-bootstrap-%s-x86_64.tar.gz.sig", url, version)
	res, err := grab.Get(".", structuredUrl)

	if err != nil {
		color.Red.Println("\r    ❎ Failed to download Signature.")
		return false, err
	}

	color.Green.Println("\r    ✅ Downloaded Signature", res.Filename)

	return true, nil
}
