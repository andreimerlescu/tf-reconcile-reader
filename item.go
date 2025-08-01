package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (i mainItem) Title() string       { return i.log.Command }
func (i mainItem) Description() string { return fmt.Sprintf("Exit: %d", i.log.ExitCode) }
func (i mainItem) FilterValue() string { return i.log.Command }

func (i backupItem) Title() string       { return i.key }
func (i backupItem) Description() string { return i.val }
func (i backupItem) FilterValue() string { return i.key }

func (i resultCategoryItem) Title() string       { return i.name }
func (i resultCategoryItem) Description() string { return fmt.Sprintf("%d items", i.count) }
func (i resultCategoryItem) FilterValue() string { return i.name }

func (i configItem) Title() string       { return i.key }
func (i configItem) Description() string { return i.val }
func (i configItem) FilterValue() string { return i.key }

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var title string
	switch item := listItem.(type) {
	case mainItem:
		title = item.Title()
	case backupItem:
		title = fmt.Sprintf("%-25s = %s", item.Title(), item.Description())
	case resultCategoryItem:
		title = fmt.Sprintf("%-25s %s", item.Title(), item.Description())
	case JSONResultItem:
		title = item.Title()
	case configItem:
		title = fmt.Sprintf("%s = %s", item.Title(), item.Description())
	default:
		return
	}

	str := strings.Split(title, "\n")[0] // Ensure single line for list view

	style := itemStyle
	if index == m.Index() {
		style = selectedItemStyle
		_, _ = fmt.Fprint(w, style.Render("> "+str))
	} else {
		_, _ = fmt.Fprint(w, style.Render("  "+str))
	}
}
