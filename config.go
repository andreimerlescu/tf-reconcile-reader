package main

import (
	"github.com/andreimerlescu/figtree/v2"
	"os"
	"path/filepath"
	"strconv"
)

func setup() {
	figs = figtree.With(figtree.Options{
		Germinate:  true,
		ConfigFile: configFile(),
		Tracking:   false,
	})
}

func configFile() string {
	v, ok := os.LookupEnv(envConfigFile)
	if !ok {
		return filepath.Join(".", "config.yaml")
	}
	return v
}

func configure(_ error) error {
	if err := figs.Load(); err != nil {
		return err
	}
	// Ensure the save directory exists
	savePath := *figs.String(argSaveDir)
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return err
	}
	return nil
}

func application() error {
	if figs == nil {
		setup()
	}
	// -input
	figs = figs.NewString(argInputFile, filepath.Join(".", "report.dev.json"), "Path to input file (report.<env>.json)")
	figs = figs.WithAlias(argInputFile, argAliasInputFile)
	figs = figs.WithValidator(argInputFile, figtree.AssureStringNoPrefix("~"))
	figs = figs.WithValidator(argInputFile, figtree.AssureStringHasSuffix(".json"))
	figs = figs.WithValidator(argInputFile, figtree.AssureStringLengthGreaterThan(7))

	// -contains
	figs = figs.NewString(argCommandContains, "", "Substring search of executed command to select when using -non-interactive")
	figs = figs.WithAlias(argCommandContains, argAliasCommandContains)

	// -json
	figs = figs.NewBool(argJsonOutput, false, "format output as JSON")
	figs = figs.WithAlias(argJsonOutput, argAliasJsonOutput)

	// -non-interactive
	figs = figs.NewBool(argNonInteractive, false, "enable non-interactive mode (auto responds with no)")

	// -save
	figs = figs.NewString(argSaveDir, filepath.Join(".", "tf-state-man-workspace"), "Path to save downloaded S3 files")

	// -vim
	vimEnabled := false
	v, ok := os.LookupEnv(envVimMode)
	if ok {
		ok, _ := strconv.ParseBool(v)
		if ok {
			vimEnabled = true
		}
	}
	figs = figs.NewBool(argVimEnabled, vimEnabled, "enable vim-style keybindings (h to go back, j, k to scroll)")

	// -github
	githubEnabled, _ := strconv.ParseBool(os.Getenv(envGitHub))
	figs = figs.NewBool(argGitHub, githubEnabled, "Indicate if running in a GitHub Actions environment")

	// -v (version)
	figs = figs.NewBool(argVersion, false, "print version")

	return nil
}
