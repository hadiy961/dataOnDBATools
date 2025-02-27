package tables

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DetailTableColumn struct {
	Title       string
	Width       int
	HeaderStyle *lipgloss.Style
	Truncate    bool
}

type DetailTableRow struct {
	Label string
	Value string
}

type DetailTable struct {
	Title          string
	Rows           []DetailTableRow
	Width          int
	UpdateFunction func()
	DeleteFunction func()
	BackFunction   func()
	NavigationText string
}

type DetailTableOption func(*DetailTable)

func WithDetailTitle(title string) DetailTableOption {
	return func(t *DetailTable) {
		t.Title = title
	}
}

func WithDetailBackFunction(fn func()) DetailTableOption {
	return func(t *DetailTable) {
		t.BackFunction = fn
	}
}

func WithDetailUpdateFunction(fn func()) DetailTableOption {
	return func(t *DetailTable) {
		t.UpdateFunction = fn
	}
}

func WithDetailDeleteFunction(fn func()) DetailTableOption {
	return func(t *DetailTable) {
		t.DeleteFunction = fn
	}
}

func NewDetailTable(opts ...DetailTableOption) *DetailTable {
	t := &DetailTable{
		NavigationText: "b: Back • u: Update • d: Delete • q: Exit",
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *DetailTable) AddDetailRow(label, value string) {
	t.Rows = append(t.Rows, DetailTableRow{Label: label, Value: value})
}

func (t *DetailTable) Init() tea.Cmd {
	return nil
}

// Remove all selection-related code and simplify the Update method
func (t *DetailTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return t, tea.Quit
		case "b":
			if t.BackFunction != nil {
				t.BackFunction()
				return t, tea.Quit
			}
		case "u":
			if t.UpdateFunction != nil {
				t.UpdateFunction()
				return t, tea.Quit
			}
		case "d":
			if t.DeleteFunction != nil {
				t.DeleteFunction()
				return t, tea.Quit
			}
		}
	}
	return t, nil
}

func (t *DetailTable) View() string {
	var b strings.Builder

	// Title
	if t.Title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
		b.WriteString("\n " + titleStyle.Render(t.Title) + "\n")
	}

	// Calculate widths
	labelWidth := 15
	valueWidth := 40

	for _, row := range t.Rows {
		if len(row.Label) > labelWidth {
			labelWidth = len(row.Label)
		}
		if len(row.Value) > valueWidth {
			valueWidth = len(row.Value)
		}
	}

	labelWidth += 2
	valueWidth += 2
	totalWidth := labelWidth + valueWidth + 3

	// Divider
	divider := strings.Repeat("-", totalWidth)
	b.WriteString(divider + "\n")

	// Header
	headerFormat := fmt.Sprintf(" %%-%ds | %%-%ds\n", labelWidth-1, valueWidth-1)
	b.WriteString(fmt.Sprintf(headerFormat, "Title", "Value"))
	b.WriteString(divider + "\n")

	// Rows - without selection highlighting
	rowFormat := fmt.Sprintf(" %%-%ds | %%-%ds\n", labelWidth-1, valueWidth-1)
	for _, row := range t.Rows {
		b.WriteString(fmt.Sprintf(rowFormat, row.Label, row.Value))
	}

	b.WriteString(divider)
	// Update navigation text to remove up/down navigation
	b.WriteString("\nb: Back • u: Update • d: Delete • q: Exit\n")

	return b.String()
}

func RunDetailTable(table *DetailTable) error {
	p := tea.NewProgram(table)
	_, err := p.Run()
	return err
}
