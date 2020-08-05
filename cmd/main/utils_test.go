package main

import (
	"testing"
)

func TestIsSystemBranch(t *testing.T) {

	systemGitTags := "master,release-.*"
	appConfig.systemGitTags = &systemGitTags

	got := isSystemBranch("master")
	want := true

	if got != want {
		t.Errorf("TestIsSystemBranch, got=%t want=%t", got, want)
	}

	got = isSystemBranch("test")
	want = false

	if got != want {
		t.Errorf("TestIsSystemBranch, got=%t want=%t", got, want)
	}

	got = isSystemBranch("release-123456")
	want = true

	if got != want {
		t.Errorf("TestIsSystemBranch, got=%t want=%t", got, want)
	}
}

func TestIsSystemNamespace(t *testing.T) {

	systemNamespaces := "kube-system,app"
	appConfig.systemNamespaces = &systemNamespaces

	got := isSystemNamespace("master")
	want := false

	if got != want {
		t.Errorf("TestIsSystemNamespace, got=%t want=%t", got, want)
	}

	got = isSystemNamespace("app")
	want = true

	if got != want {
		t.Errorf("TestIsSystemNamespace, got=%t want=%t", got, want)
	}

	got = isSystemNamespace("release-123456")
	want = false

	if got != want {
		t.Errorf("TestIsSystemNamespace, got=%t want=%t", got, want)
	}
}
