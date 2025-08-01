package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func run() error {
	// Load data from the input file
	report, err := loadReportData(*figs.String(argInputFile))
	if err != nil {
		return err
	}

	// Filter results
	containsFilter := *figs.String(argCommandContains)
	fmt.Printf("FILTER: %s\n", containsFilter)
	var filteredLogs []CommandExecutionLog

	for _, log := range report.ExecutionLogs {
		log.Stderr = strings.ReplaceAll(log.Stderr, "%0A", "\n")
		log.Stdout = strings.ReplaceAll(log.Stdout, "%0A", "\n")
		a := strings.Contains(log.Command, containsFilter)
		b := strings.Contains(log.Stderr, containsFilter)
		c := strings.Contains(log.Stdout, containsFilter)
		if containsFilter == "" || a || b || c {
			filteredLogs = append(filteredLogs, log)
		}
	}

	// Output results in the desired format
	if *figs.Bool(argJsonOutput) {
		return printJSON(filteredLogs)
	}
	printText(filteredLogs)
	return nil
}

func printText(logs []CommandExecutionLog) {
	for _, log := range logs {
		fmt.Printf("COMMAND:\n%s\n", log.Command)
		if log.ExitCode != 0 {
			fmt.Printf("Exit Code: %d", log.ExitCode)
			if log.Error != "" {
				fmt.Printf(" | Error: %s", log.Error)
			}
			fmt.Println()
		}
		fmt.Println("\nSTDOUT:")
		fmt.Println(log.Stdout)
		fmt.Println("STDERR:")
		fmt.Println(log.Stderr)
		fmt.Println("---")
	}
}

func printJSON(logs []CommandExecutionLog) error {
	// Wrapper struct for the final JSON output
	type jsonResult struct {
		Results []CommandExecutionLog `json:"results"`
	}

	output := jsonResult{
		Results: logs,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results to JSON: %w", err)
	}

	fmt.Fprintln(os.Stdout, string(jsonData))
	return nil
}
