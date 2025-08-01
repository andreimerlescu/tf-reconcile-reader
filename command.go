package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// editCommand opens the user's default editor to modify a command string.
func editCommand(commandStr string) tea.Cmd {
	return func() tea.Msg {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim" // A sensible default
		}

		tmpfile, err := os.CreateTemp("", "command-*.sh")
		if err != nil {
			return commandEditedMsg{err: fmt.Errorf("could not create temp file: %w", err)}
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.WriteString(commandStr); err != nil {
			return commandEditedMsg{err: fmt.Errorf("could not write to temp file: %w", err)}
		}
		if err := tmpfile.Close(); err != nil {
			return commandEditedMsg{err: fmt.Errorf("could not close temp file: %w", err)}
		}

		cmd := exec.Command(editor, tmpfile.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// We need to pause the Bubble Tea program to run the editor
		// This is handled by p.Send(tea.Suspend) in the Update function
		if err := cmd.Run(); err != nil {
			return commandEditedMsg{err: fmt.Errorf("editor command failed: %w", err)}
		}

		editedBytes, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			return commandEditedMsg{err: fmt.Errorf("could not read back temp file: %w", err)}
		}

		return commandEditedMsg{newCmd: strings.TrimSpace(string(editedBytes))}
	}
}

// execCommand prepares and runs an exec.Cmd, returning a message with the output.
func execCommand(commandStr string) tea.Cmd {
	return func() tea.Msg {
		tfDir := os.Getenv(envTfDir)
		tfState := os.Getenv(envTfState)

		// Modify command if necessary
		if !strings.Contains(commandStr, "-state=") && tfState != "" {
			commandStr = strings.Replace(commandStr, "terraform", fmt.Sprintf("terraform -state=%s", tfState), 1)
		}

		cmd := exec.Command("sh", "-c", commandStr)
		if isValidPath(tfDir) {
			cmd.Dir = tfDir
		}

		var out, errOut strings.Builder
		cmd.Stdout = &out
		cmd.Stderr = &errOut

		err := cmd.Run()

		return commandOutputMsg{
			stdout: out.String(),
			stderr: errOut.String(),
			err:    err,
		}
	}
}
