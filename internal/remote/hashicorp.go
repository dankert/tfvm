package remote

import (
	"fmt"
	"github.com/cbuschka/tfvm/internal/util"
	"github.com/cbuschka/tfvm/internal/version"
	goversion "github.com/hashicorp/go-version"
	"io/ioutil"
	"net/http"
	"regexp"
	"runtime"
	"strings"
)

func ListTerraformReleases() ([]*version.TerraformVersion, error) {

	releasesPage, err := downloadReleasesPage()
	if err != nil {
		return nil, err
	}

	return extractReleases(releasesPage)
}

func downloadReleasesPage() (string, error) {
	url := fmt.Sprintf("%s/index.html", getReleasesBaseUrl())
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(html), nil
}

func extractReleases(releasePage string) ([]*version.TerraformVersion, error) {
	re, err := regexp.Compile(">terraform_([^<]+)</a>")
	if err != nil {
		return nil, err
	}

	matchSets := re.FindAllStringSubmatch(releasePage, -1)

	releases := make([]*version.TerraformVersion, len(matchSets))
	for index, matchSet := range matchSets {
		semVersionStr := strings.TrimSpace(matchSet[1])
		semVersion, err := goversion.NewVersion(semVersionStr)
		if err != nil {
			return nil, err
		}
		releases[index] = &version.TerraformVersion{Version: semVersion}
	}

	return releases, nil
}

func GetUrl(release *version.TerraformVersion) string {
	tfArch := util.GetFirstEnv("TFVM_TERRAFORM_ARCH", "TERRAFORM_ARCH")
	if tfArch == "" {
		tfArch = runtime.GOARCH
	}

	tfOs := util.GetFirstEnv("TFVM_TERRAFORM_OS", "TERRAFORM_OS")
	if tfOs == "" {
		tfOs = runtime.GOOS
	}

	return fmt.Sprintf("%s/%s/terraform_%s_%s_%s.zip", getReleasesBaseUrl(),
		release.Version.String(), release.Version, tfOs, tfArch)
}

func getReleasesBaseUrl() string {
	baseUrl := util.GetFirstEnv("TFVM_TERRAFORM_RELEASES_BASE_URL")
	if baseUrl != "" {
		return baseUrl
	}

	return "https://releases.hashicorp.com/terraform"
}
