package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
)

type RepoType int

const (
	Personal RepoType = iota
	Work
	External
)

// TODO: Read from config
var Domains = map[RepoType][]string{
	Personal: {"poppet.io"},
	Work:     {},
}

// TODO: Read from config
var Users = map[RepoType][]string{
	Personal: {"gazwald"},
	Work:     {},
}

type Details struct {
	Url      string
	Hostname string
	Path     string
	Repopath string
	Type     RepoType
}

func (d Details) String() string {
	return fmt.Sprintf("Details{\n\tURL: \t\t%s,\n\tHostname: \t%s,\n\tPath: \t\t%s,\n\tFile: \t\t%s,\n\tType: \t\t%v\n}",
		d.Url,
		d.Hostname,
		d.Path,
		d.Repopath,
		d.Type)
}

func getUrl() string {
	if len(os.Args) == 2 {
		return strings.ToLower(os.Args[1])
	}
	return ""
}

// TODO: Clean up
func parseGitURL(u string) (*url.URL, error) {
	if strings.HasPrefix(u, "git@") {
		u = strings.Replace(u, ":", "/", 1)
		u = strings.Replace(u, "git@", "https://", -1)
	}
	u = strings.TrimSuffix(u, ".git")
	parsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	return parsed, nil
}

func processUrl(url string) (*Details, error) {
	parsed, err := parseGitURL(url)
	if err != nil {
		return nil, fmt.Errorf("Url did not match expected format: %s", url)
	}

	repoType := getRepoType(parsed.Host, parsed.Path)
	Repopath, err := processUrlPath(repoType, parsed.Path)
	if err != nil {
		return nil, err
	}

	return &Details{
		Hostname: parsed.Host,
		Url:      url,
		Path:     parsed.Path,
		Repopath: Repopath,
		Type:     repoType,
	}, nil
}

func checkDomain(repoType RepoType, hostname string) bool {
	for _, domain := range Domains[repoType] {
		if hostname == domain {
			return true
		}
	}
	return false
}

func checkUser(repoType RepoType, urlPath string) bool {
	for _, user := range Users[repoType] {
		if strings.Contains(urlPath, user) {
			return true
		}
	}
	return false
}

func getRepoType(hostname string, urlPath string) RepoType {
	for _, repoType := range []RepoType{Personal, Work} {
		if checkDomain(repoType, hostname) || checkUser(repoType, urlPath) {
			return repoType
		}
	}
	return External
}

func processUrlPath(repoType RepoType, urlpath string) (string, error) {
	splitUrl := strings.Split(urlpath, "/")
	if len(splitUrl) < 2 {
		return "", fmt.Errorf("Path did not contain at least one '/' separator: %s", urlpath)
	}

	// TODO: Use HomDir
	// home, _ := os.UserHomeDir()
	home := "/tmp"
	repoOrg, repoName := splitUrl[1], splitUrl[2]

	switch repoType {
	case Personal:
		return path.Join(home, "local/checkouts/personal", repoName), nil
	case Work:
		return path.Join(home, "local/checkouts/work", repoOrg, repoName), nil
	default:
		return path.Join(home, "local/checkouts/external", repoOrg, repoName), nil
	}
}

func isDir(repopath string) bool {
	info, err := os.Stat(repopath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func createDir(details Details) bool {
	if !isDir(details.Repopath) {
		os.MkdirAll(details.Repopath, 0750)
		return true
	} else {
		if !isDir(path.Join(details.Repopath, ".git")) {
			return true
		}
	}
	return false
}

func main() {
	details, err := processUrl(getUrl())
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(details)
	createDir(*details)
	// TODO: Skip clone if it already exists
	command := exec.Command("git", "clone", details.Url, details.Repopath)
	command.Run()

	// TODO: git status
	// 			 print Repopath
}
