package util

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/go-logr/logr"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	ProxyImageName = "proxy"

	CollectorImageName = "wavefront-kubernetes-collector"

	DockerHubImagePrefix = "wavefronthq/"
)

// GetLatestVersion checks for auto upgrade eligibility and returns the latest minor version as applicable.
func GetLatestVersion(crImage string, enableAutoUpgrade bool, reqLogger logr.Logger) (string, error) {
	// Auto upgrade is supported only for docker hub images.
	if !strings.HasPrefix(crImage, DockerHubImagePrefix) {
		reqLogger.Info("Auto Upgrade not supported,", "Cause :: Not a Docker Hub Image.", crImage)
		return "", nil
	}

	imgSlice := strings.Split(crImage, ":")
	crImageName := strings.TrimPrefix(imgSlice[0], DockerHubImagePrefix)
	currentVersion := imgSlice[1]

	// Auto Upgrade support
	if !enableAutoUpgrade {
		reqLogger.Info("Auto Upgrade not supported,", " Cause :: enableAutoUpgrade is set to ", enableAutoUpgrade)
		return currentVersion, nil
	}

	// "latest" effectively renders auto upgrade useless.
	if currentVersion == "latest" {
		reqLogger.Info("Auto Upgrade not supported,", " Cause :: currentVersion is ", currentVersion)
		return currentVersion, nil
	}

	// The last 20 tags should be good. Don't expect customers to be using a really old version of CR.
	url := "https://registry.hub.docker.com/v2/repositories/wavefronthq/" + crImageName + "/tags/?page_size=20"
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected http GET status: %s", resp.Status)
	}
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// The below will get us the versions from json.
	// Ex for Proxy: [latest 5.5 5.1 4.38 4.36 4.35 4.34 4.33 4.32 4.31]
	versions := gjson.Get(string(contents), "results.#.name")

	majorVersion := strings.Split(currentVersion, ".")[0]

	finalSemV, err := semver.NewVersion(currentVersion)
	if err != nil {
		return "", err
	}

	// Filter based on major version and then minor version (Also, should be non-"rc" build).
	foundUpgradeVersion := false
	for _, v := range versions.Array() {
		if strings.HasPrefix(v.String(), majorVersion) && !strings.Contains(v.String(), "rc") && !strings.Contains(v.String(), "beta") {
			if semV, err := semver.NewVersion(v.String()); err == nil {
				if semV.GreaterThan(finalSemV) {
					finalSemV = semV
					foundUpgradeVersion = true
				}
			}
		}
	}

	if foundUpgradeVersion {
		reqLogger.Info("Found newer Minor Upgrade version :: " + finalSemV.Original() + ", " +
			"current version " + currentVersion)
	}

	return finalSemV.Original(), nil
}
