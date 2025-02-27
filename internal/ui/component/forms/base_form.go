package forms

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// BaseForm provides base implementation for forms
type BaseForm struct {
	FormModel
}

// NewBaseForm creates a new base form
func NewBaseForm(title, description string) *BaseForm {
	return &BaseForm{
		FormModel: FormModel{
			Title:       title,
			Description: description,
			Fields:      make([]FormField, 0),
		},
	}
}

// AddField adds a new field to the form
func (f *BaseForm) AddField(field FormField) {
	f.Fields = append(f.Fields, field)
}

func (f *BaseForm) Init() tea.Cmd {
	return textinput.Blink
}

func (f *BaseForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return f, tea.Quit

		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			if s == "up" || s == "shift+tab" {
				f.Focused--
			} else {
				f.Focused++
			}

			if f.Focused > len(f.Fields) {
				f.Focused = 0
			} else if f.Focused < 0 {
				f.Focused = len(f.Fields)
			}

			cmds := make([]tea.Cmd, len(f.Fields))
			for i := range f.Fields {
				if i == f.Focused {
					cmds[i] = f.Fields[i].Input.Focus()
				} else {
					f.Fields[i].Input.Blur()
				}
			}

			return f, tea.Batch(cmds...)

		case "enter":
			if f.Focused == len(f.Fields) {
				if err := f.ValidateAll(); err == nil {
					result := f.HandleSubmit()
					if result.Error == nil {
						f.Done = true
						return f, tea.Quit
					}
					f.Error = result.Error
				}
			}
		}
	}

	// Handle character input
	cmd := f.updateInputs(msg)

	return f, cmd
}

func (f *BaseForm) ValidateAll() error {
	for _, field := range f.Fields {
		if err := field.Validate(); err != nil {
			f.Error = err
			return err
		}
	}
	return nil
}

func (f *BaseForm) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(f.Fields))

	for i := range f.Fields {
		f.Fields[i].Input.PromptStyle = BlurredStyle
		f.Fields[i].Input.TextStyle = BlurredStyle
		if i == f.Focused {
			f.Fields[i].Input.PromptStyle = FocusedStyle
			f.Fields[i].Input.TextStyle = FocusedStyle
		}

		var cmd tea.Cmd
		f.Fields[i].Input, cmd = f.Fields[i].Input.Update(msg)
		cmds[i] = cmd
	}

	return tea.Batch(cmds...)
}

func (f *BaseForm) View() string {
	var b strings.Builder

	if f.Title != "" {
		b.WriteString(TitleStyle.Render(f.Title) + "\n\n")
	}

	if f.Description != "" {
		b.WriteString(f.Description + "\n\n")
	}

	// Fields
	for i := range f.Fields {
		label := fmt.Sprintf("%s:", f.Fields[i].Label)
		labelText := LabelStyle.Render(label)

		// Handle input display
		var displayValue string
		actualValue := f.Fields[i].Input.Value()
		if actualValue == "" && f.Fields[i].Default != "" {
			// Show placeholder if no value
			displayValue = PlaceholderStyle.Render(f.Fields[i].Default)
		} else {
			// Show actual value
			if i == f.Focused {
				displayValue = FocusedStyle.Render(actualValue)
			} else {
				displayValue = BlurredStyle.Render(actualValue)
			}
		}

		input := "> " + displayValue
		paddedInput := InputStyle.Render(input)

		b.WriteString(fmt.Sprintf("%-25s %s\n", labelText, paddedInput))
	}

	// Submit button with consistent alignment
	button := BlurredButton
	if f.Focused == len(f.Fields) {
		button = FocusedButton
	}
	b.WriteString("\n" + button + "\n")

	if f.Error != nil {
		b.WriteString(ErrorStyle.Render("Error: "+f.Error.Error()) + "\n")
	}

	return b.String()
}

func (f *BaseForm) GetFormModel() *FormModel {
	return &f.FormModel
}

func (f *BaseForm) HandleSubmit() FormResult {
	values := make(map[string]string)
	for _, field := range f.Fields {
		values[field.Name] = field.Input.Value()
	}

	if f.OnSubmit != nil {
		if err := f.OnSubmit(values); err != nil {
			return FormResult{Error: err}
		}
	}

	return FormResult{Values: values}
}
