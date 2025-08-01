package main

import (
	"embed"
	"fmt"
	"os"
	"strings"
)

//go:embed VERSION
var versionBytes embed.FS

var currentVersion string

func About() string {
	return Version() + " by " + appAuthor
}

func Version() string {
	if len(currentVersion) == 0 {
		versionBytes, err := versionBytes.ReadFile("VERSION")
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return "v0.0.0"
		}
		currentVersion = strings.TrimSpace(string(versionBytes))
	}
	return currentVersion
}
