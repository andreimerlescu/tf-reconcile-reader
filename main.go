package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var err error
	check(configure(application()))

	if *figs.Bool(argVersion) {
		fmt.Println(Version())
		os.Exit(0)
	}

	if reportData, err = loadReportData(*figs.String(argInputFile)); err != nil {
		log.Fatalf("Failed to load report data: %v", err)
	}

	if *figs.Bool(argNonInteractive) {
		fmt.Println("NON INTERACTIVE MODE ENABLED")
		check(run())
	} else {
		errorCounts := aggregateErrors(reportData)
		p := tea.NewProgram(
			initialModel(reportData, errorCounts, Version()),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
		defer clearTUI()
		var finalModel tea.Model
		if finalModel, err = p.Run(); err != nil {
			log.Fatalf("Error running program: %v", err)
		}

		if m, ok := finalModel.(model); ok && m.quitMessage != "" {
			fmt.Println(m.quitMessage)
		}
	}
}
