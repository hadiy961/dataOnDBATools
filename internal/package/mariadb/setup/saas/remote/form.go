package saas_remote

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Custom styles for the form
var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// Model for the remote server setup form
type RemoteServerFormModel struct {
	inputs     []textinput.Model
	focusIndex int
	errorMsg   string
	submitted  bool
	values     map[string]string
}

// Initialize the form model
func NewRemoteServerFormModel() RemoteServerFormModel {
	// Create four text inputs for IP, port, user, and password
	inputs := make([]textinput.Model, 4)

	// IP Address input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Required"
	inputs[0].Focus()
	inputs[0].Width = 30
	inputs[0].Prompt = "> "

	// SSH Port input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "2930"
	inputs[1].Width = 30
	inputs[1].Prompt = "> "

	// SSH User input
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "root"
	inputs[2].Width = 30
	inputs[2].Prompt = "> "

	// SSH Password input - with masking
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "••••••••••"
	inputs[3].Width = 30
	inputs[3].Prompt = "> "
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].EchoCharacter = '•'

	return RemoteServerFormModel{
		inputs:     inputs,
		focusIndex: 0,
		values:     make(map[string]string),
	}
}

func (m RemoteServerFormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m RemoteServerFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Handle navigation between fields
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			// Wrap around
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			} else if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			}

			// Update focus for inputs
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i < len(m.inputs); i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}

			return m, tea.Batch(cmds...)

		// Handle form submission with enter
		case "enter":
			// If focusIndex is beyond inputs, this means we're at the submit button
			if m.focusIndex == len(m.inputs) {
				// Validate IP address - the only required field
				if m.inputs[0].Value() == "" {
					m.errorMsg = "IP Address is required"
					return m, nil
				}

				// Form submitted successfully
				m.submitted = true

				// Collect values
				m.values["ip_address"] = m.inputs[0].Value()
				m.values["ssh_port"] = m.inputs[1].Value()
				m.values["ssh_user"] = m.inputs[2].Value()
				m.values["ssh_pass"] = m.inputs[3].Value()

				// Use defaults for empty values
				if m.values["ssh_port"] == "" {
					m.values["ssh_port"] = "2930"
				}
				if m.values["ssh_user"] == "" {
					m.values["ssh_user"] = "root"
				}
				if m.values["ssh_pass"] == "" {
					m.values["ssh_pass"] = "2025Dataon!"
				}

				return m, tea.Quit
			}
		}
	}

	// Handle characters typed into the fields
	cmd := m.updateInputs(msg)
	return m, cmd
}

// Function to update all inputs
func (m *RemoteServerFormModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds[i] = cmd
	}

	return tea.Batch(cmds...)
}

func (m RemoteServerFormModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Remote Server Setup") + "\n\n")
	b.WriteString("Enter connection details for the remote server:\n\n")

	// Display all inputs with proper labels
	labels := []string{"IP Address", "SSH Port", "SSH User", "SSH Password"}

	for i := range m.inputs {
		var style lipgloss.Style
		if i == m.focusIndex {
			style = focusedStyle
		} else {
			style = blurredStyle
		}

		// Format the label with right alignment
		label := fmt.Sprintf("%s:", labels[i])
		paddedLabel := fmt.Sprintf("%-15s", label)

		// Apply styling
		styledLabel := style.Render(paddedLabel)
		styledInput := m.inputs[i].View()

		b.WriteString(styledLabel + " " + styledInput + "\n")
	}

	// Submit button
	submitButton := "[ Submit ]"
	if m.focusIndex == len(m.inputs) {
		submitButton = focusedStyle.Render(submitButton)
	} else {
		submitButton = blurredStyle.Render(submitButton)
	}
	b.WriteString("\n" + submitButton + "\n")

	// Display error if any
	if m.errorMsg != "" {
		b.WriteString("\n" + errorStyle.Render("Error: "+m.errorMsg) + "\n")
	}

	return b.String()
}
