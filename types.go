package main

import (
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

type (
	// ErrorCount stores a specific error message and its occurrence count.
	ErrorCount struct {
		Error string `json:"error"`
		Count int    `json:"count"`
	}

	commandEditedMsg struct {
		newCmd string
		err    error
	}

	viewState int

	// CommandExecutionLog stores the result of a single executed command.
	CommandExecutionLog struct {
		TerraformAddress string `json:"terraform_address,omitempty"`
		Command          string `json:"command"`
		Stdout           string `json:"stdout"`
		Stderr           string `json:"stderr"`
		Error            string `json:"error,omitempty"`
		ExitCode         int    `json:"exit_code"`
		Source           string `json:"source,omitempty"`
	}

	// JSONBackupPaths contains paths and checksums for created backup artifacts.
	JSONBackupPaths struct {
		OriginalPath       string `json:"original_path"`
		OriginalChecksum   string `json:"original_checksum"`
		NewPath            string `json:"new_path"`
		NewChecksum        string `json:"new_checksum"`
		ReportPath         string `json:"report_path"`
		ReportChecksum     string `json:"report_checksum"`
		JsonReportPath     string `json:"json_report_path"`
		JsonReportChecksum string `json:"json_report_checksum"`
	}

	// JSONResultItem is a simplified version of ResourceStatus for the final JSON report.
	JSONResultItem struct {
		Resource string `json:"resource"`
		Command  string `json:"command,omitempty"`
		Kind     string `json:"kind"`
		TFID     string `json:"tf_id"`
		AWSID    string `json:"aws_id"`
		Message  string `json:"message"`
	}

	// JSONResults organizes the final results by category for the JSON report.
	JSONResults struct {
		InfoResults            []JSONResultItem `json:"INFO"`
		OkResults              []JSONResultItem `json:"OK"`
		PotentialImportResults []JSONResultItem `json:"POTENTIAL_IMPORT"`
		RegionMismatchResults  []JSONResultItem `json:"REGION_MISMATCH"`
		WarningResults         []JSONResultItem `json:"WARNING"`
		ErrorResults           []JSONResultItem `json:"ERROR"`
		DangerousResults       []JSONResultItem `json:"DANGEROUS"`
	}

	// JSONOutput is the root object for the final JSON report.
	JSONOutput struct {
		State            string                `json:"state"`
		StateChecksum    string                `json:"state_checksum"`
		Region           string                `json:"region"`
		LocalStateFile   string                `json:"local_statefile"`
		TFVersion        string                `json:"tf_version"`
		StateVersion     uint64                `json:"state_version"`
		Concurrency      int                   `json:"concurrency"`
		Backup           JSONBackupPaths       `json:"backup"`
		ExecutionLogs    []CommandExecutionLog `json:"execution_logs"`
		Results          JSONResults           `json:"results"`
		Arguments        []string              `json:"arguments"`
		Version          string                `json:"version"` // bfsm version
		ApplicationError string                `json:"application_error,omitempty"`
	}

	// mainItem represents an item in the primary execution log list.
	mainItem struct {
		log CommandExecutionLog
	}

	// backupItem represents a key-value pair from the backup configuration.
	backupItem struct {
		key, val string
	}

	// resultCategoryItem represents a category in the results view.
	resultCategoryItem struct {
		name  string
		count int
	}

	// configItem represents an environment variable for configuration.
	configItem struct {
		key, val string
	}

	// --- TUI list delegate ---

	itemDelegate struct{}

	copyToClipboardMsg   struct{ content string }
	copyStatusMsg        struct{ err error }
	clearNotificationMsg struct{}
	deleteResourceMsg    struct{ path string }
	deleteResultMsg      struct{ err error }
	timeoutMsg           struct{}
	runCommandMsg        struct{ content string }
	commandOutputMsg     struct {
		stdout, stderr string
		err            error
	}

	versions struct {
		tf, bfsm, bsmr string
	}

	// --- MODEL ---
	model struct {
		// common
		report      *JSONOutput
		errorCounts *sync.Map
		versions    versions
		state       viewState
		termWidth   int
		termHeight  int
		ready       bool
		err         error

		// components
		mainList            list.Model
		backupList          list.Model
		resultsCategoryList list.Model
		resultsResourceList list.Model
		configList          list.Model
		viewPort            viewport.Model
		spinNer             spinner.Model
		textInput           textinput.Model

		// ui state
		notification string
		quitMessage  string // for final message on exit
		viewStack    []viewState

		// view-specific data
		activeResultCategory string
		activeBackupItem     backupItem
		activeExecutionLog   CommandExecutionLog
		activeConfigItem     configItem
		deleteConfirmPath    string
		deleteTimer          *time.Timer
		commandRunner        struct {
			cmd, stdout, stderr string
			exitError           error
		}
	}
)
