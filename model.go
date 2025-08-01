package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/andreimerlescu/prettyboy/prettyboy"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

func initialModel(report *JSONOutput, errors *sync.Map, bsmrVersion string) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	m := model{
		report:      report,
		state:       viewMain,
		viewStack:   []viewState{viewMain},
		spinNer:     s,
		errorCounts: errors,
		textInput:   ti,
		versions: versions{
			tf:   report.TFVersion,
			bfsm: report.Version,
			bsmr: bsmrVersion,
		},
	}
	return m
}

// A command that copies text to the clipboard.
func copyToClipboardCmd(text string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(text)
		return copyStatusMsg{err: err}
	}
}

func (m model) Init() tea.Cmd {
	return m.spinNer.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth, m.termHeight = msg.Width, msg.Height
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		listHeight := m.termHeight - (headerHeight + footerHeight)
		viewportHeight := m.termHeight - (headerHeight + footerHeight + 4) // +4 for title/help margins

		if !m.ready {
			delegate := itemDelegate{}
			m.mainList = list.New(nil, delegate, 0, 0)
			m.backupList = list.New(nil, delegate, 0, 0)
			m.resultsCategoryList = list.New(nil, delegate, 0, 0)
			m.resultsResourceList = list.New(nil, delegate, 0, 0)
			m.configList = list.New(nil, delegate, 0, 0)
			m.viewPort = viewport.New(m.termWidth-4, viewportHeight)
			m.ready = true
			m.loadMainList()
		}

		m.mainList.SetSize(m.termWidth, listHeight)
		m.backupList.SetSize(m.termWidth, listHeight)
		m.resultsCategoryList.SetSize(m.termWidth, listHeight)
		m.resultsResourceList.SetSize(m.termWidth, listHeight)
		m.configList.SetSize(m.termWidth, listHeight)
		m.viewPort.Width = m.termWidth - 4
		m.viewPort.Height = viewportHeight
		return m, nil

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Pop view if not in main view
		if m.state != viewMain && (msg.String() == "q" || msg.String() == "esc") {
			m.state = m.popView()
			return m, nil
		}

	case spinner.TickMsg:
		m.spinNer, cmd = m.spinNer.Update(msg)
		return m, cmd

	case copyStatusMsg:
		if msg.err != nil {
			m.setNotification(fmt.Sprintf("Error copying: %v", msg.err), true)
		} else {
			m.setNotification("Copied to clipboard!", false)
		}
		return m, m.clearNotificationAfter(2 * time.Second)

	case clearNotificationMsg:
		m.notification = ""
		return m, nil

	case commandOutputMsg:
		m.commandRunner.stdout = msg.stdout
		m.commandRunner.stderr = msg.stderr
		m.commandRunner.exitError = msg.err
		// Append to in-memory logs
		source := "LOCAL"
		if *figs.Bool(argGitHub) {
			source = "GITHUB ACTIONS"
		}
		newLog := CommandExecutionLog{
			Command: m.commandRunner.cmd,
			Stdout:  msg.stdout,
			Stderr:  msg.stderr,
			Source:  source,
		}
		if msg.err != nil {
			newLog.Error = msg.err.Error()
			var exitErr *exec.ExitError
			if errors.As(msg.err, &exitErr) {
				newLog.ExitCode = exitErr.ExitCode()
			}
		}
		m.report.ExecutionLogs = append(m.report.ExecutionLogs, newLog)
		m.loadMainList() // Refresh the main list
		return m, nil

	case commandEditedMsg:
		if msg.err != nil {
			m.setNotification(fmt.Sprintf("Editor error: %v", msg.err), true)
			return m, m.clearNotificationAfter(3 * time.Second)
		}
		return m, copyToClipboardCmd(msg.newCmd)

	case error:
		m.err = msg
		return m, tea.Quit
	}

	// View-specific updates
	switch m.state {
	case viewMain:
		return m.updateMainView(msg)
	case viewBackup:
		m, cmd = m.updateBackupView(msg)
	case viewBackupDetail:
		m, cmd = m.updateBackupDetailView(msg)
	case viewResultsCategory:
		m, cmd = m.updateResultsCategoryView(msg)
	case viewResultsList:
		m, cmd = m.updateResultsListView(msg)
	case viewResultsDetail:
		m, cmd = m.updateResultsDetailView(msg)
	case viewExecutionLogDetail:
		m, cmd = m.updateExecutionLogDetailView(msg)
	case viewConfig:
		m, cmd = m.updateConfigView(msg)
	case viewConfigEdit:
		m, cmd = m.updateConfigEditView(msg)
	case viewCommandRunner:
		m, cmd = m.updateCommandRunnerView(msg)
	default:
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\nAn error occurred: %v\n\n", m.err))
	}
	if !m.ready {
		return fmt.Sprintf("\n  %s Initializing...\n", m.spinNer.View())
	}

	var mainContent string
	switch m.state {
	case viewMain:
		mainContent = m.mainList.View()
	case viewBackup:
		mainContent = m.backupList.View()
	case viewResultsCategory:
		mainContent = m.resultsCategoryList.View()
	case viewResultsList:
		mainContent = m.resultsResourceList.View()
	case viewConfig:
		mainContent = m.configList.View()
	case viewBackupDetail, viewResultsDetail, viewExecutionLogDetail:
		mainContent = m.viewPort.View()
	case viewConfigEdit:
		mainContent = fmt.Sprintf(
			"Editing %s:\n\n%s\n\n%s",
			keyStyle.Render(m.activeConfigItem.key),
			m.textInput.View(),
			helpStyle.Render("(esc to cancel, enter to save and quit)"),
		)
	case viewCommandRunner:
		mainContent = m.viewCommandRunner()
	default:
		mainContent = fmt.Sprintf("Unknown view %d", m.state)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.headerView(),
		mainContent,
		m.footerView(),
	)
}

func (m model) updateMainView(msg tea.Msg) (model, tea.Cmd) {
	if !m.ready {
		return m, nil
	}
	if m.mainList.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.mainList, cmd = m.mainList.Update(msg)
		return m, cmd
	}
	switch asMsg := msg.(type) {
	case tea.KeyMsg:
		switch asMsg.String() {
		case "q":
			return m, tea.Quit
		case "b":
			m.state = m.pushView(viewBackup)
			return m, m.loadBackupList()
		case "r":
			m.state = m.pushView(viewResultsCategory)
			return m, m.loadResultsCategories()
		case "c":
			m.state = m.pushView(viewConfig)
			return m, m.loadConfigList()
		case "s", "l", "enter":
			if i, ok := m.mainList.SelectedItem().(mainItem); ok {
				m.activeExecutionLog = i.log
				m.state = m.pushView(viewExecutionLogDetail)
				m.renderExecutionLogDetail()
			}
			return m, nil
		default:
		}
	}
	var cmd tea.Cmd
	m.mainList, cmd = m.mainList.Update(msg)
	return m, cmd
}

func (m model) updateBackupView(msg tea.Msg) (model, tea.Cmd) {
	switch asMsg := msg.(type) {
	case tea.KeyMsg:
		if asMsg.Type == tea.KeyEnter {
			if i, ok := m.backupList.SelectedItem().(backupItem); ok {
				m.activeBackupItem = i
				m.state = m.pushView(viewBackupDetail)
				m.renderBackupDetail()
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.backupList, cmd = m.backupList.Update(msg)
	return m, cmd
}

func (m model) updateBackupDetailView(msg tea.Msg) (model, tea.Cmd) {
	switch asMsg := msg.(type) {
	case tea.KeyMsg:
		switch asMsg.String() {
		case "c":
			return m, copyToClipboardCmd(m.activeBackupItem.val)
		case "enter":
			path := m.activeBackupItem.val
			if !isValidPath(path) {
				m.setNotification(fmt.Sprintf("Local path not found: %s", path), true)
				return m, m.clearNotificationAfter(3 * time.Second)
			}
			info, err := os.Stat(path)
			if err != nil {
				m.setNotification(fmt.Sprintf("Could not stat file: %v", err), true)
				return m, m.clearNotificationAfter(3 * time.Second)
			}
			details := fmt.Sprintf("File: %s\nSize: %d bytes\nPermissions: %s\nModified: %s",
				info.Name(), info.Size(), info.Mode(), info.ModTime().Format(time.RFC1123))
			m.viewPort.SetContent(details)
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.viewPort, cmd = m.viewPort.Update(msg)
	return m, cmd
}

func (m model) updateResultsCategoryView(msg tea.Msg) (model, tea.Cmd) {
	switch asMsg := msg.(type) {
	case tea.KeyMsg:
		if asMsg.Type == tea.KeyEnter {
			if i, ok := m.resultsCategoryList.SelectedItem().(resultCategoryItem); ok {
				m.activeResultCategory = i.name
				m.state = m.pushView(viewResultsList)
				return m, m.loadResultsList()
			}
		}
	}
	var cmd tea.Cmd
	m.resultsCategoryList, cmd = m.resultsCategoryList.Update(msg)
	return m, cmd
}

func (m model) updateResultsListView(msg tea.Msg) (model, tea.Cmd) {
	switch asMsg := msg.(type) {
	case tea.KeyMsg:
		if asMsg.Type == tea.KeyEnter {
			if i, ok := m.resultsResourceList.SelectedItem().(JSONResultItem); ok {
				m.state = m.pushView(viewResultsDetail)
				m.renderResourceDetail(i)
			}
		}
	}
	var cmd tea.Cmd
	m.resultsResourceList, cmd = m.resultsResourceList.Update(msg)
	return m, cmd
}

func (m model) updateResultsDetailView(msg tea.Msg) (model, tea.Cmd) {
	switch asMsg := msg.(type) {
	case tea.KeyMsg:
		switch asMsg.String() {
		case "X":
			if i, ok := m.resultsResourceList.SelectedItem().(JSONResultItem); ok && i.Command != "" {
				m.commandRunner.cmd = i.Command
				m.state = m.pushView(viewCommandRunner)
				return m, execCommand(i.Command)
			}
			m.setNotification("No command to execute for this item.", true)
			return m, m.clearNotificationAfter(2 * time.Second)
		}
	}
	var cmd tea.Cmd
	m.viewPort, cmd = m.viewPort.Update(msg)
	return m, cmd
}

func (m model) updateConfigView(msg tea.Msg) (model, tea.Cmd) {
	switch asMsg := msg.(type) {
	case tea.KeyMsg:
		if asMsg.Type == tea.KeyEnter {
			if i, ok := m.configList.SelectedItem().(configItem); ok {
				m.activeConfigItem = i
				m.textInput.SetValue(i.val)
				m.textInput.Focus()
				m.state = m.pushView(viewConfigEdit)
			}
		}
	}
	var cmd tea.Cmd
	m.configList, cmd = m.configList.Update(msg)
	return m, cmd
}

func (m model) updateConfigEditView(msg tea.Msg) (model, tea.Cmd) {
	switch asMsg := msg.(type) {
	case tea.KeyMsg:
		switch asMsg.String() {
		case "enter":
			m.quitMessage = fmt.Sprintf("export %s=\"%s\"", m.activeConfigItem.key, m.textInput.Value())
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) updateCommandRunnerView(msg tea.Msg) (model, tea.Cmd) {
	switch asMsg := msg.(type) {
	case tea.KeyMsg:
		if asMsg.Type == tea.KeyEnter {
			m.state = m.popView() // Return to the results detail view
		}
	}
	return m, nil
}

func (m *model) loadMainList() {
	items := make([]list.Item, len(m.report.ExecutionLogs))
	if len(m.report.ExecutionLogs) == 0 {
		items = []list.Item{mainItem{log: CommandExecutionLog{Command: "No execution logs found in this report."}}}
	} else {
		for i, log := range m.report.ExecutionLogs {
			items[i] = mainItem{log: log}
		}
	}
	m.mainList.Title = fmt.Sprintf("Execution Logs (%d)", len(m.report.ExecutionLogs))
	m.mainList.SetItems(items)
}

func (m *model) loadBackupList() tea.Cmd {
	items := toBackupItems(m.report.Backup)
	m.backupList.Title = "Backup File Paths"
	m.backupList.SetItems(items)
	return nil
}

func (m *model) loadResultsCategories() tea.Cmd {
	items := make([]list.Item, len(resultCategories))
	for i, cat := range resultCategories {
		items[i] = resultCategoryItem{
			name:  cat,
			count: len(m.report.Results.GetCategory(cat)),
		}
	}
	m.resultsCategoryList.Title = "Result Categories"
	m.resultsCategoryList.SetItems(items)
	return nil
}

func (m *model) loadResultsList() tea.Cmd {
	results := m.report.Results.GetCategory(m.activeResultCategory)
	items := make([]list.Item, len(results))
	for i, res := range results {
		items[i] = res
	}
	m.resultsResourceList.Title = fmt.Sprintf("Results: %s (%d)", m.activeResultCategory, len(results))
	m.resultsResourceList.SetItems(items)
	return nil
}

func (m *model) loadConfigList() tea.Cmd {
	var items []list.Item
	knownKeys := []string{envTfDir, envTfState}
	envMap := make(map[string]string)

	for _, key := range knownKeys {
		envMap[key] = os.Getenv(key)
	}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "FIGS_") || strings.HasPrefix(e, "OLLAMA_") {
			parts := strings.SplitN(e, "=", 2)
			envMap[parts[0]] = parts[1]
		}
	}
	for key, val := range envMap {
		items = append(items, configItem{key: key, val: val})
	}
	m.configList.Title = "Environment Configuration"
	m.configList.SetItems(items)
	return nil
}

func (m *model) renderBackupDetail() {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.activeBackupItem.key) + "\n\n")

	path := m.activeBackupItem.val
	s3Path := ""

	if !isValidPath(path) && strings.HasPrefix(m.report.State, "s3://") {
		if idx := strings.Index(path, "backups/"); idx != -1 {
			truncatedPath := path[idx+len("backups/"):]
			bucket, _, _ := parseS3Path(m.report.State)
			s3Path = fmt.Sprintf("s3://%s/state-backups/%s", bucket, truncatedPath)
		}
	}

	b.WriteString(valueStyle.Render(path))
	if s3Path != "" {
		b.WriteString(fmt.Sprintf("\n\n%s\n%s",
			titleStyle.Render("Derived S3 Path:"),
			valueStyle.Render(s3Path),
		))
	}
	m.viewPort.SetContent(b.String())
}

func (m *model) renderExecutionLogDetail() {
	var b strings.Builder
	execLog := m.activeExecutionLog
	b.WriteString(titleStyle.Render("Command:") + "\n")
	b.WriteString(wordwrap.String(prettyboy.Command(execLog.Command), m.viewPort.Width-12) + "\n\n")
	b.WriteString(titleStyle.Render(fmt.Sprintf("Exit Code: %d", execLog.ExitCode)) + "\n\n")
	b.WriteString(titleStyle.Render("STDOUT:") + "\n")
	b.WriteString(execLog.Stdout + "\n\n")
	b.WriteString(titleStyle.Render("STDERR:") + "\n")
	b.WriteString(errorStyle.Render(execLog.Stderr))
	m.viewPort.SetContent(b.String())
}

func (m *model) renderResourceDetail(item JSONResultItem) {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s\n", titleStyle.Render("Resource:"), item.Resource))
	b.WriteString(fmt.Sprintf("%s %s\n", titleStyle.Render("Kind:"), item.Kind))
	b.WriteString(fmt.Sprintf("%s %s\n", titleStyle.Render("Terraform ID:"), item.TFID))
	b.WriteString(fmt.Sprintf("%s %s\n", titleStyle.Render("AWS ID:"), item.AWSID))
	b.WriteString(fmt.Sprintf("\n%s\n%s", titleStyle.Render("Message:"), item.Message))
	if item.Command != "" {
		b.WriteString(fmt.Sprintf("\n\n%s\n%s",
			titleStyle.Render("Suggested Command:"),
			wordwrap.String(prettyboy.Command(item.Command), m.viewPort.Width-12),
		))
	}
	m.viewPort.SetContent(b.String())
}

func (m *model) pushView(state viewState) viewState {
	m.viewStack = append(m.viewStack, state)
	return state
}

func (m *model) popView() viewState {
	if len(m.viewStack) <= 1 {
		return m.state
	}
	m.viewStack = m.viewStack[:len(m.viewStack)-1]
	return m.viewStack[len(m.viewStack)-1]
}

func (m *model) setNotification(msg string, isError bool) {
	style := notificationStyle
	if isError {
		style = notificationStyle.Copy().Background(lipgloss.Color("196"))
	}
	m.notification = style.Render(msg)
}

func (m model) clearNotificationAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearNotificationMsg{}
	})
}

func (m model) viewCommandRunner() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Executing Command:") + "\n")
	b.WriteString(wordwrap.String(prettyboy.Command(m.commandRunner.cmd), m.viewPort.Width-12) + "\n\n")
	b.WriteString(titleStyle.Render("STDOUT:") + "\n")
	b.WriteString(m.commandRunner.stdout + "\n")
	b.WriteString(titleStyle.Render("STDERR:") + "\n")
	b.WriteString(errorStyle.Render(m.commandRunner.stderr) + "\n")
	b.WriteString(helpStyle.Render("\nPress Enter to return to the previous view."))
	return b.String()
}

func (m model) updateExecutionLogDetailView(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "c":
			return m, copyToClipboardCmd(m.activeExecutionLog.Command)
		case "e":
			// Suspend bubbletea to allow editor to take over terminal
			return m, tea.Batch(tea.Suspend, editCommand(m.activeExecutionLog.Command))
		}
	}
	var cmd tea.Cmd
	m.viewPort, cmd = m.viewPort.Update(msg)
	return m, cmd
}

func (m model) headerView() string {
	title := fmt.Sprintf("%s: %s", appName, *figs.String(argInputFile))
	versions := fmt.Sprintf("tf v%s | btfsm %s | bsmr %s   ", m.versions.tf, m.versions.bfsm, m.versions.bsmr)
	spaceWidth := m.termWidth - lipgloss.Width(title) - lipgloss.Width(versions)
	if spaceWidth < 1 {
		spaceWidth = 1
	}
	spacer := strings.Repeat(" ", spaceWidth)
	return titleStyle.Render(title + spacer + versions)
}

func (m model) footerView() string {
	if m.notification != "" {
		return m.notification
	}
	return helpStyle.Render(m.helpView())
}

func (m model) helpView() string {
	k := func(s string) string { return keyStyle.Render(s) }
	switch m.state {
	case viewMain:
		return fmt.Sprintf("↑/↓: Navigate | Enter: Details | %s: Backups | %s: Results | %s: Configs | %s: Quit", k("b"), k("r"), k("c"), k("q"))
	case viewBackup:
		return fmt.Sprintf("↑/↓: Navigate | Enter: Details | %s/%s: Back", k("q"), k("esc"))
	case viewBackupDetail:
		return fmt.Sprintf("↑/↓: Scroll | %s: Copy Path | Enter: Stat File | %s/%s: Back", k("c"), k("q"), k("esc"))
	case viewResultsCategory:
		return fmt.Sprintf("↑/↓: Navigate | Enter: View Items | %s/%s: Back", k("q"), k("esc"))
	case viewResultsList:
		return fmt.Sprintf("↑/↓: Navigate | Enter: Details | %s/%s: Back", k("q"), k("esc"))
	case viewResultsDetail:
		return fmt.Sprintf("↑/↓: Scroll | %s: Execute Command | %s/%s: Back", k("X"), k("q"), k("esc"))
	case viewConfig:
		return fmt.Sprintf("↑/↓: Navigate | Enter: Edit | %s/%s: Back", k("q"), k("esc"))
	case viewConfigEdit:
		return fmt.Sprintf("Enter: Save and Quit | %s: Back", k("esc"))
	case viewCommandRunner:
		return "Enter: Back"
	case viewExecutionLogDetail:
		return fmt.Sprintf("↑/↓: Scroll | %s: Copy | %s: Edit & Copy | %s/%s: Back", k("c"), k("e"), k("q"), k("esc"))
	default:
		return fmt.Sprintf("Use %s or %s to go back. %s to quit.", k("q"), k("esc"), k("ctrl+c"))
	}
}
