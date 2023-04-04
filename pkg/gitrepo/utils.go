package gitrepo

import (
	"github.com/lacodon/recoon/pkg/config"
	"os"
	"path/filepath"
	"strings"
)

func MakeLocalPath(cloneUrl, branchName string) string {
	destinationPath := filepath.Join(config.Cfg.Store.GitDir, MakeAPIName(cloneUrl, branchName))
	_ = os.MkdirAll(destinationPath, 0664)
	return destinationPath
}

func MakeAPIName(cloneUrl, branchName string, suffixes ...string) string {
	replacer := strings.NewReplacer(
		"git@", "",
		"/", ":",
		".git", "")

	name := cloneUrl + ":" + branchName

	if suffixes != nil {
		suffixStr := strings.Join(suffixes, "#")
		suffixStr = strings.ReplaceAll(suffixStr, "/", "+")
		name += ":" + suffixStr
	}

	return strings.Trim(replacer.Replace(name), ":+")
}
