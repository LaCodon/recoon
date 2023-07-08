package gitrepo

import (
	"os"
	"path/filepath"
	"strings"
)

func MakeLocalPath(localDir, cloneUrl, branchName string) string {
	destinationPath := filepath.Join(localDir, MakeAPIName(cloneUrl, branchName))
	_ = os.MkdirAll(destinationPath, 0664)
	return destinationPath
}

func MakeAPIName(cloneUrl, branchName string, suffixes ...string) string {
	replacer := strings.NewReplacer(
		"git@", "",
		"/", "#",
		":", "#",
		".git", "")

	name := cloneUrl + "#" + branchName

	if suffixes != nil {
		suffixStr := strings.Join(suffixes, "#")
		suffixStr = strings.ReplaceAll(suffixStr, "/", "+")
		name += "#" + suffixStr
	}

	return strings.Trim(replacer.Replace(name), "#+")
}
