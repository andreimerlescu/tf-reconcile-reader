package main

import (
	"github.com/andreimerlescu/figtree/v2"
	"github.com/charmbracelet/lipgloss"
)

var (
	figs figtree.Plant

	reportData *JSONOutput

	// resultCategories provides a consistent order for iterating through result types.
	resultCategories = []string{
		"INFO", "OK", "POTENTIAL_IMPORT", "REGION_MISMATCH", "WARNING", "ERROR", "DANGEROUS",
	}

	titleStyle         = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("205"))
	itemStyle          = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle  = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("170"))
	notificationStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("228")).Background(lipgloss.Color("63")).Padding(0, 1)
	helpStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	keyStyle           = helpStyle.Copy().Foreground(lipgloss.Color("208"))
	valueStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
	confirmPromptStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	codeStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Background(lipgloss.Color("236")).Padding(0, 1)
	boldStyle          = lipgloss.NewStyle().Bold(true)

	// Common terraform error strings to look for.
	commonTerraformErrors = []string{
		"Error:",
		"failed:",
		"timed out",
		"no such host",
		"permission denied",
		"not found",
		"InvalidClientTokenId",
		"unauthorized",
	}
)
