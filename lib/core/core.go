package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"encoding/json"

	"github.com/CompeyDev/wsl-archlinux-manager/lib/checkers"
	logger "github.com/CompeyDev/wsl-archlinux-manager/util"
	"github.com/briandowns/spinner"
	"github.com/cavaliergopher/grab/v3"
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

	bar.Suffix = " Untarring archive..."
	bar.Start()

	stdoute, untarErr := exec.Command("wsl.exe", "bash", "-c", fmt.Sprintf(`tar -xzvf archlinux-bootstrap-%s-x86_64.tar.gz`, version)).CombinedOutput()

	if untarErr != nil {
		println(untarErr)
		logger.Error("Failed to untar archive; this is a non-recoverable error. Quitting.")
	}
	println(stdoute)
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
