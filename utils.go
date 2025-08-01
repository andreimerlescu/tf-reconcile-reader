package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/list"
)

// loadReportData reads the specified JSON file and unmarshals it into our JSONOutput struct.
func loadReportData(filePath string) (*JSONOutput, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("input file not found: %s", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read input file: %w", err)
	}

	var report JSONOutput
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("could not parse JSON from input file: %w", err)
	}

	return &report, nil
}

func isValidPath(val string) bool {
	_, err := os.Lstat(val)
	return err == nil
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "\n... (truncated)"
	}
	return s
}

var check = func(what interface{}) {
	switch v := what.(type) {
	case error:
		if v != nil {
			log.Fatal(v)
		}
	default:
		log.Println(v)
	}
}

// aggregateErrors scans execution logs for common error patterns and counts them.
func aggregateErrors(report *JSONOutput) *sync.Map {
	errorCounts := &sync.Map{}
	if report.ExecutionLogs == nil {
		return errorCounts
	}

	var wg sync.WaitGroup
	for _, log := range report.ExecutionLogs {
		wg.Add(1)
		go func(log CommandExecutionLog) {
			defer wg.Done()
			// Combine stdout and stderr for searching
			output := log.Stdout + "\n" + log.Stderr
			for _, errStr := range commonTerraformErrors {
				if strings.Contains(output, errStr) {
					// We found an error pattern.
					// Load the existing count, increment it, and store it back.
					val, _ := errorCounts.LoadOrStore(errStr, &ErrorCount{Error: errStr, Count: 0})
					count := val.(*ErrorCount)
					count.Count++
				}
			}
		}(log)
	}
	wg.Wait()
	return errorCounts
}

// toBackupItems converts the Backup struct to a slice of list items.
func toBackupItems(backup JSONBackupPaths) []list.Item {
	var items []list.Item
	v := reflect.ValueOf(backup)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		key := t.Field(i).Tag.Get("json")
		val := v.Field(i).Interface().(string)
		items = append(items, backupItem{key: key, val: val})
	}
	return items
}

func clearTUI() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("sh", "-c", "reset")
	}

	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

func envAsInt(env string) int {
	vs, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s must be an integer\n", env)
		return 3
	}
	return vs
}

func envAsString(env, fallback string) string {
	v, ok := os.LookupEnv(env)
	if !ok {
		return fallback
	}
	return v
}
