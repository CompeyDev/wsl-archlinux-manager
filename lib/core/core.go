package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/CompeyDev/wsl-archlinux-manager/lib/checkers"
	logger "github.com/CompeyDev/wsl-archlinux-manager/util"
	"github.com/briandowns/spinner"
	"github.com/cavaliergopher/grab/v3"
	"github.com/schollz/progressbar/v3"
)

// TODO:

func Build() {
	fmt.Println("🔃 Setting up RootFS...")
	bar := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	bar.Prefix = " "
	bar.Suffix = " Determining the fastest mirror..."
	bar.Start()
	userLocation, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		print(err)
	}
	defer userLocation.Body.Close()
	body, reqErr := io.ReadAll(userLocation.Body)

	if reqErr != nil {
		logger.Error("An internal error occurred when attempting to pull the RootFS. This is probably a bug; you might want to report this.")
		bar.Stop()
		os.Exit(1)
	}

	var resStruct struct {
		Status      string
		Country     string
		CountryCode string
		Region      string
		RegionName  string
		City        string
		Zip         string
		Lat         float64
		Lon         float64
		Timezone    string
		Isp         string
		Org         string
		As          string
		Query       string
	}

	parseErr := json.Unmarshal([]byte(body), &resStruct)

	if parseErr != nil {
		logger.Error("Failed to parse response body! This is a probably a bug; you might want to report this.")
		bar.Stop()
		os.Exit(1)
	}
	mirror := getMirror(resStruct.Country)
	bar.Suffix = fmt.Sprintf(" Using mirror %s...", mirror)
	time.Sleep(time.Second * 5)
	bar.Stop()
	isSuccessful_1, version, _ := pullArchive(mirror)

	if !isSuccessful_1 {
		logger.Warn("Attempt #1 to pull RootFS failed, trying again with Worldwide mirror...")

		bar.Suffix = " Attempt #1 to pull RootFS failed, trying again with Worldwide mirror..."

		bar.Start()

		globalMirror := getMirror("Worldwide")
		isSuccessful_2, _, _ := pullArchive(globalMirror)

		if !isSuccessful_2 {
			logger.Error("Attempt #2 to pull RootFS failed. Please try again.")
			bar.Stop()
			os.Exit(1)
		} else {
			checkers.VerifySignature(globalMirror)
		}

	} else {
		checkers.VerifySignature(mirror)
	}

	archiveName := fmt.Sprintf(`archlinux-bootstrap-%s-x86_64.tar.gz`, version)

	untarArchive(archiveName, bar)

	bar.Suffix = "  Building bootstrap package..."

	bar.Start()

	_d, buildErr := exec.Command("wsl.exe", "bash", "-c", "cd root.x86_64 && tar -zcvf arch_bootstrap_package.tar.gz .").CombinedOutput()

	if buildErr != nil {
		bar.Stop()
		println(buildErr.Error())
		println(string(_d))
		logger.Error("Failed to build bootstrap package; fatal. Aborting installation...")
	}
	bar.Stop()
	logger.Progress("Successfully built bootstrap package, proceeeding!")
}

func untarArchive(archiveName string, bar *spinner.Spinner) {
	bar.Suffix = " Untarring archive..."
	bar.Start()

	untarSizeCmd, sizeCmdErr := exec.Command("wsl.exe", "bash", "-c", fmt.Sprintf("tar -tzf %s | wc -l", archiveName)).CombinedOutput()

	if sizeCmdErr != nil {
		bar.Stop()
		logger.Error("Failed to retrieve archive length; fatal. Refusing to continue.")
	}

	untarSizeArr := strings.Split(string(untarSizeCmd), "\n")
	untarSizeStr := strings.Trim(untarSizeArr[len(untarSizeArr)-2], "\n\r ")
	untarSize, convErr := strconv.ParseInt(untarSizeStr, 10, 64)

	if convErr != nil {
		bar.Stop()
		logger.Error("An internal error occurred while attempting to untar the archive. Aborting.")
	}

	bar.Stop()

	pb := progressbar.NewOptions(1000,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetDescription("[yellow][1/3][reset] Untarring archive..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	progressbar.DefaultBytes(untarSize)

	untarCmd := exec.Command("wsl.exe", "bash", "-c", fmt.Sprintf("tar -xzvf %s", archiveName))

	stdout, stdoutErr := untarCmd.StdoutPipe()
	stderr, _ := untarCmd.StderrPipe()
	_ = untarCmd.Start()

	scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		pb.Add64(1)
	}
	_ = untarCmd.Wait()

	if stdoutErr != nil {
		logger.Error("Failed to untar archive; this is a non-recoverable error. Quitting.")
	}
	bar.Stop()
	logger.Info("Successfully untarred archive!")
}

func getMirror(country string) string {
	resp, err := http.Get("https://archlinux.org/download/")
	if err != nil {
		logger.Error("Failed to download RootFS.")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("An internal error occurred when attempting to pull the RootFS. This is probably a bug; you might want to report this.")
	}

	mirrorLink := strings.Split(strings.Split(strings.Split(strings.Split(strings.Split(strings.Split(string(body), fmt.Sprintf(`title="%s"`, country))[1], `title="Download from`)[0], fmt.Sprintf(`></span> %s</h5>`, country))[1], `<ul>`)[1], `<li><a href="`)[1], `"`)[0]
	return mirrorLink
}

func pullArchive(url string) (isSuccessful bool, ver string, error error) {
	bar := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	bar.Prefix = " "
	bar.Suffix = " Downloading RootFS..."

	var version = strings.Split(strings.Split(url, "iso/")[1], "/")[0]
	var structuredUrl = fmt.Sprintf("%s/archlinux-bootstrap-%s-x86_64.tar.gz", url, version)
	res, err := grab.Get(".", structuredUrl)

	if err != nil {
		logger.Error("Failed to download RootFS.")
		return false, "UNKNOWN", err
	}

	logger.Info(fmt.Sprintf("Downloaded RootFS %s", res.Filename))
	return true, version, nil
}
