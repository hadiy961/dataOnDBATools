package lokal_rpm

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define styles
var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
)

type TransferForm struct {
	inputs     []textinput.Model
	focusIndex int
	done       bool
}

func NewTransferForm() *TransferForm {
	inputs := make([]textinput.Model, 4)

	// Initialize text inputs
	for i := range inputs {
		input := textinput.New()
		input.Cursor.Style = cursorStyle
		input.CharLimit = 64

		// Set field-specific properties
		switch i {
		case 0:
			input.Placeholder = "root"
			input.Focus()
		case 1:
			input.Placeholder = "192.168.101.131"
		case 2:
			input.Placeholder = "2930"
		case 3:
			input.Placeholder = "/root/mariadb_packages"
		}

		inputs[i] = input
	}

	return &TransferForm{
		inputs: inputs,
	}
}

func (f *TransferForm) Init() tea.Cmd {
	return textinput.Blink
}

func (f *TransferForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return f, tea.Quit
		case "tab", "shift+tab", "enter", "up", "down":
			// Handle navigation
			if msg.String() == "enter" && f.focusIndex == len(f.inputs)-1 {
				f.done = true
				return f, tea.Quit
			}

			// Cycle through inputs
			if msg.String() == "up" || msg.String() == "shift+tab" {
				f.focusIndex--
				if f.focusIndex < 0 {
					f.focusIndex = len(f.inputs) - 1
				}
			} else {
				f.focusIndex++
				if f.focusIndex >= len(f.inputs) {
					f.focusIndex = 0
				}
			}

			for i := 0; i < len(f.inputs); i++ {
				if i == f.focusIndex {
					cmds = append(cmds, f.inputs[i].Focus())
				} else {
					f.inputs[i].Blur()
				}
			}

			return f, tea.Batch(cmds...)
		}
	}

	// Handle character input
	cmd := f.updateInputs(msg)

	return f, cmd
}

func (f *TransferForm) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	for i := range f.inputs {
		if i == f.focusIndex {
			f.inputs[i], cmd = f.inputs[i].Update(msg)
		}
	}

	return cmd
}

func (f *TransferForm) View() string {
	var b strings.Builder

	b.WriteString("\n  Transfer Configuration\n\n")

	for i := range f.inputs {
		field := ""
		switch i {
		case 0:
			field = "Username"
		case 1:
			field = "IP Address"
		case 2:
			field = "Port"
		case 3:
			field = "Destination Directory"
		}

		b.WriteString(fmt.Sprintf("  %-23s %s\n", field, f.inputs[i].View()))
	}

	b.WriteString("\n  Press tab/shift+tab to navigate • enter to confirm • ctrl+c to quit\n")

	return b.String()
}

func (f *TransferForm) GetConfig() TransferConfig {
	return TransferConfig{
		User:      getInputValue(f.inputs[0], "root"),
		IPAddress: getInputValue(f.inputs[1], "192.168.101.131"),
		Port:      getInputValue(f.inputs[2], "2930"),
		DestDir:   getInputValue(f.inputs[3], "/root/mariadb_packages"),
	}
}

func getInputValue(input textinput.Model, defaultValue string) string {
	value := strings.TrimSpace(input.Value())
	if value == "" {
		return defaultValue
	}
	return value
}

// promptTransferConfig now uses the Bubble Tea form
func promptTransferConfig() TransferConfig {
	form := NewTransferForm()
	p := tea.NewProgram(form)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running form: %v\n", err)
		return TransferConfig{}
	}

	if !form.done {
		return TransferConfig{} // User cancelled
	}

	return form.GetConfig()
}
