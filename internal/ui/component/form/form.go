// Package form menyediakan komponen form yang bisa digunakan kembali
package form

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Field mendefinisikan sebuah field dalam form
type Field struct {
	Name        string
	Label       string
	Placeholder string
	Required    bool
	HelperText  string
	Width       int
	Value       string
}

// Styles berisi semua informasi styling untuk form
type Styles struct {
	Focused     lipgloss.Style
	Blurred     lipgloss.Style
	Title       lipgloss.Style
	Error       lipgloss.Style
	Helper      lipgloss.Style
	Button      lipgloss.Style
	FocusedBtn  lipgloss.Style
	BlurredBtn  lipgloss.Style
	TextStyle   lipgloss.Style
	Placeholder lipgloss.Style
}

// FormOptions berisi konfigurasi untuk form
type FormOptions struct {
	Title         string
	Description   string
	SubmitLabel   string
	CancelLabel   string
	NavigationMsg string
}

// Model mengelola state form
type Model struct {
	inputs     []textinput.Model
	focusIndex int
	errorMsg   string
	submitted  bool
	canceled   bool
	values     map[string]string
	fields     []Field
	styles     Styles
	options    FormOptions
}

// DefaultStyles mengembalikan konfigurasi styling default
func DefaultStyles() Styles {
	return Styles{
		Focused:     lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
		Blurred:     lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")),
		Error:       lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Helper:      lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Italic(true),
		Button:      lipgloss.NewStyle().Padding(0, 3),
		FocusedBtn:  lipgloss.NewStyle().Padding(0, 3).Foreground(lipgloss.Color("205")).Bold(true),
		BlurredBtn:  lipgloss.NewStyle().Padding(0, 3).Foreground(lipgloss.Color("252")),
		TextStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		Placeholder: lipgloss.NewStyle().Foreground(lipgloss.Color("246")),
	}
}

// DefaultFormOptions mengembalikan opsi form default
func DefaultFormOptions() FormOptions {
	return FormOptions{
		Title:         "Form",
		Description:   "Please fill in the form:",
		SubmitLabel:   "Submit",
		CancelLabel:   "Cancel",
		NavigationMsg: "Tab/Up/Down: Navigate • Enter: Select • Esc: Cancel",
	}
}

// New membuat instance form baru dengan styling dan opsi default
func New(fields []Field) Model {
	return NewCustom(fields, DefaultStyles(), DefaultFormOptions())
}

// NewCustom membuat form baru dengan opsi styling dan konfigurasi kustom
func NewCustom(fields []Field, styles Styles, options FormOptions) Model {
	inputs := make([]textinput.Model, len(fields))

	// Inisialisasi inputs berdasarkan definisi field
	for i, field := range fields {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = field.Placeholder
		inputs[i].Width = field.Width
		inputs[i].Prompt = "> "
		inputs[i].TextStyle = styles.TextStyle
		inputs[i].PlaceholderStyle = styles.Placeholder

		// Set nilai default jika ada
		if field.Value != "" {
			inputs[i].SetValue(field.Value)
		}
	}

	// Set focus pada field pertama
	if len(inputs) > 0 {
		inputs[0].Focus()
	}

	return Model{
		inputs:     inputs,
		focusIndex: 0,
		values:     make(map[string]string),
		fields:     fields,
		styles:     styles,
		options:    options,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		// Menangani navigasi antar field
		case "tab", "shift+tab", "up", "down", "left", "right":
			s := msg.String()

			// Menangani navigasi horizontal antar tombol
			if m.focusIndex >= len(m.inputs) {
				if s == "left" && m.focusIndex == len(m.inputs)+1 {
					m.focusIndex = len(m.inputs)
					return m, nil
				} else if s == "right" && m.focusIndex == len(m.inputs) {
					m.focusIndex = len(m.inputs) + 1
					return m, nil
				}
			}

			// Navigasi vertikal normal
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else if s == "down" || s == "tab" {
				m.focusIndex++
			}

			// Wrap around
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + 1
			} else if m.focusIndex > len(m.inputs)+1 {
				m.focusIndex = 0
			}

			// Update focus untuk inputs
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i < len(m.inputs); i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}

			return m, tea.Batch(cmds...)

		// Menangani submisi form dengan enter
		case "enter":
			// Tombol Submit
			if m.focusIndex == len(m.inputs) {
				// Validasi field yang required
				for i, field := range m.fields {
					if field.Required && m.inputs[i].Value() == "" {
						m.errorMsg = fmt.Sprintf("%s is required", field.Label)
						return m, nil
					}
				}

				// Form berhasil disubmit
				m.submitted = true

				// Kumpulkan nilai
				for i, field := range m.fields {
					m.values[field.Name] = m.inputs[i].Value()
				}

				return m, tea.Quit
			}

			// Tombol Cancel
			if m.focusIndex == len(m.inputs)+1 {
				m.canceled = true
				return m, tea.Quit
			}

		// Menangani tombol escape untuk membatalkan
		case "esc":
			m.canceled = true
			return m, tea.Quit
		}
	}

	// Menangani karakter yang diketik dalam field
	cmd := m.updateInputs(msg)
	return m, cmd
}

// updateInputs memperbarui semua input
func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		if i == m.focusIndex {
			var cmd tea.Cmd
			m.inputs[i], cmd = m.inputs[i].Update(msg)
			cmds[i] = cmd
		}
	}

	return tea.Batch(cmds...)
}

// GetValues mengembalikan nilai-nilai form
func (m *Model) GetValues() map[string]string {
	return m.values
}

// IsSubmitted mengembalikan true jika form telah disubmit
func (m *Model) IsSubmitted() bool {
	return m.submitted
}

// IsCanceled mengembalikan true jika form telah dibatalkan
func (m *Model) IsCanceled() bool {
	return m.canceled
}

func (m Model) View() string {
	var b strings.Builder

	// Judul
	b.WriteString(m.styles.Title.Render("\n"+m.options.Title) + "\n")
	b.WriteString(m.styles.Blurred.Render(m.options.Description) + "\n\n")

	// Tampilkan semua input dengan label yang sesuai
	for i, field := range m.fields {
		var style lipgloss.Style
		if i == m.focusIndex {
			style = m.styles.Focused
		} else {
			style = m.styles.Blurred
		}

		// Format label dengan right alignment
		paddedLabel := fmt.Sprintf("%-15s", field.Label)

		// Terapkan styling
		styledLabel := style.Render(paddedLabel)
		styledInput := m.inputs[i].View()

		b.WriteString(styledLabel + " " + styledInput + "\n")

		// Tambahkan helper text
		if i == m.focusIndex {
			b.WriteString(fmt.Sprintf("%15s %s\n", "", m.styles.Helper.Render(field.HelperText)))
		} else {
			b.WriteString("\n")
		}
	}

	// Baris tombol
	submitButton := "[ " + m.options.SubmitLabel + " ]"
	cancelButton := "[ " + m.options.CancelLabel + " ]"

	// Terapkan styling berdasarkan fokus
	if m.focusIndex == len(m.inputs) {
		submitButton = m.styles.FocusedBtn.Render(submitButton)
		cancelButton = m.styles.BlurredBtn.Render(cancelButton)
	} else if m.focusIndex == len(m.inputs)+1 {
		submitButton = m.styles.BlurredBtn.Render(submitButton)
		cancelButton = m.styles.FocusedBtn.Render(cancelButton)
	} else {
		submitButton = m.styles.BlurredBtn.Render(submitButton)
		cancelButton = m.styles.BlurredBtn.Render(cancelButton)
	}

	b.WriteString("\n" + submitButton + " " + cancelButton + "\n")

	// Tampilkan error jika ada
	if m.errorMsg != "" {
		b.WriteString("\n" + m.styles.Error.Render("Error: "+m.errorMsg) + "\n")
	}

	// Bantuan navigasi
	b.WriteString("\n" + m.styles.Helper.Render(m.options.NavigationMsg) + "\n")

	return b.String()
}
