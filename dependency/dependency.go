package dependency

import (
	"encoding/json"
	"github.com/Masterminds/semver/v3"
	"os"
	"regexp"
	"sort"
)

type Version struct {
}

type Attachment struct {
	Shasum string `json:"shasum"`
}

type DistFile struct {
	Url      string `json:"url"`
	Sha      string `json:"sha"`
	Registry string `json:"registry"`
}

type Package struct {
	Name        string                 `json:"name"`
	Versions    map[string]interface{} `json:"versions"`
	Time        map[string]string      `json:"time"`
	Users       interface{}            `json:"users"`
	DistTags    map[string]string      `json:"dist-tags"`
	Uplinks     interface{}            `json:"_uplinks"`
	DistFiles   map[string]DistFile    `json:"_distfiles"`
	Attachments map[string]Attachment  `json:"_attachments"`
	Rev         string                 `json:"_rev"`
	Id          string                 `json:"_id"`
	Readme      string                 `json:"readme"`
}

func GetPackage(path string) (*Package, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	pkg := Package{}
	err = json.Unmarshal(content, &pkg)
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

// GetLocalDistFiles 获取依赖包目录下的所有发布版
func GetLocalDistFiles(path string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for _, file := range files {
		if file.Name() != "package.json" {
			names = append(names, file.Name())
		}
	}
	return names, nil
}

func GetVersions(dists []string) []string {
	reg := regexp.MustCompile(`(.+)-(\d+\.\d+\.\d+.*).tgz`)
	var versions = make([]string, 0)
	for _, dist := range dists {
		match := reg.FindStringSubmatch(dist)
		if len(match) > 1 {
			versions = append(versions, match[2])
		}
	}
	return versions
}

func GetSortedDistFiles(dists []string) []string {
	sort.SliceStable(dists, func(i, j int) bool {
		preVersion := GetVersionFromDistFile(dists[i])
		currVersion := GetVersionFromDistFile(dists[j])
		return semver.MustParse(currVersion).LessThan(semver.MustParse(preVersion))
	})
	return dists
}

func GetSortedVersions(versions []string) []string {
	sort.SliceStable(versions, func(i, j int) bool {
		return semver.MustParse(versions[j]).LessThan(semver.MustParse(versions[i]))
	})
	return versions
}

func GetVersionFromDistFile(dist string) string {
	reg := regexp.MustCompile(`(.+)-(\d+\.\d+\.\d+.*).tgz`)
	match := reg.FindStringSubmatch(dist)
	if len(match) == 3 {
		return match[2]
	}
	return ""
}

func GetLatestDist(dists []string) string {
	sortedDists := GetSortedDistFiles(dists)
	if len(sortedDists) > 0 {
		return sortedDists[0]
	} else {
		return ""
	}
}
