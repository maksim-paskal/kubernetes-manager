package main

import (
	"regexp"
	"strings"
	"time"
)

func isSystemBranch(gitBranch string) bool {
	for _, gitBranchRegexp := range strings.Split(*appConfig.systemGitTags, ",") {
		re := regexp.MustCompile(gitBranchRegexp)
		if re.MatchString(strings.ToLower(gitBranch)) {
			return true
		}
	}
	return false
}

func isSystemNamespace(namespace string) bool {
	for _, namespaceRegexp := range strings.Split(*appConfig.systemNamespaces, ",") {
		re := regexp.MustCompile(namespaceRegexp)
		if re.MatchString(strings.ToLower(namespace)) {
			return true
		}
	}
	return false
}

func diffToNow(t time.Time) int {
	t1 := time.Now()
	return int(t1.Sub(t).Hours() / 24)
}
