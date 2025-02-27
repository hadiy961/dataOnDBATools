package forms

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	FocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	BlurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	LabelStyle = lipgloss.NewStyle().
			Width(25).
			Align(lipgloss.Right)

	InputStyle = lipgloss.NewStyle().
			PaddingLeft(1)

	PlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				PaddingLeft(1)

	FocusedButton = FocusedStyle.Copy().
			Render("[ Submit ]")

	BlurredButton = BlurredStyle.Copy().
			Render("[ Submit ]")

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// FormField represents a single form field
type FormField struct {
	Label       string
	Name        string
	Value       string
	Default     string
	Required    bool
	Placeholder string
	Input       textinput.Model
	Validator   func(string) error
}

// FormModel represents the base form model
type FormModel struct {
	Fields      []FormField
	Focused     int
	Done        bool
	Error       error
	Title       string
	Description string
	OnSubmit    func(map[string]string) error
}

// FormResult contains the form submission results
type FormResult struct {
	Values map[string]string
	Error  error
}

// Validatable interface for form validation
type Validatable interface {
	Validate() error
}

// Submittable interface for form submission
type Submittable interface {
	Submit() FormResult
}

// FormComponent interface that all forms must implement
type FormComponent interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
	GetFormModel() *FormModel
	HandleSubmit() FormResult
}

// NewPasswordFormField creates a form field specifically for passwords
func NewPasswordFormField(name, label, defaultValue string, required bool) FormField {
	// Create a new input with password masking enabled from the start
	input := textinput.New()
	input.Prompt = ""
	input.PromptStyle = BlurredStyle
	input.TextStyle = BlurredStyle
	input.Width = 40
	input.EchoMode = textinput.EchoPassword
	input.EchoCharacter = '•'

	return FormField{
		Label:    label,
		Name:     name,
		Default:  defaultValue,
		Required: required,
		Input:    input,
	}
}

func NewFormField(name, label, placeholder string, required bool) FormField {
	input := textinput.New()
	input.Prompt = "" // Remove default prompt
	input.PromptStyle = BlurredStyle
	input.TextStyle = BlurredStyle
	input.Width = 40
	input.PlaceholderStyle = PlaceholderStyle

	return FormField{
		Label:    label,
		Name:     name,
		Default:  placeholder, // Store placeholder in Default field
		Required: required,
		Input:    input,
	}
}

// SetValidator sets a custom validator for the form field
func (f *FormField) SetValidator(validator func(string) error) {
	f.Validator = validator
}

// Validate validates the form field
func (f *FormField) Validate() error {
	if f.Required && f.Input.Value() == "" {
		return fmt.Errorf("%s is required", f.Label)
	}
	if f.Validator != nil {
		return f.Validator(f.Input.Value())
	}
	return nil
}
