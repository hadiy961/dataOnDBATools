package StatusForm

import (
	"dbaTools/internal/logger"
	"dbaTools/internal/ui/component/form"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// showStatusAddForm displays the form and returns the values if submitted
func ShowStatusAddForm(log *logger.Logger) (map[string]string, error) {
	// Get form fields and options
	fields := GetStatusFormFields()
	options := GetStatusAddFormOptions()

	// Create form model
	formModel := form.NewCustom(fields, form.DefaultStyles(), options)

	// Run the form program
	p := tea.NewProgram(formModel)
	finalModel, err := p.Run()
	if err != nil {
		log.Error("Form error", err)
		return nil, err
	}

	// Get form data after submission
	m, ok := finalModel.(form.Model)
	if !ok {
		return nil, fmt.Errorf("unexpected model type")
	}

	// Check form status
	if m.IsCanceled() {
		return nil, fmt.Errorf("setup cancelled")
	}

	if !m.IsSubmitted() {
		return nil, fmt.Errorf("setup not completed")
	}

	// Return the form values
	return m.GetValues(), nil
}

// GetStatusAddFormOptions returns form options for Status add operation
func GetStatusAddFormOptions() form.FormOptions {
	options := form.DefaultFormOptions()
	options.Title = "Add Status"
	options.Description = "Enter new Status information:"
	options.SubmitLabel = "Submit"
	return options
}

// StatusFormData berisi definisi field untuk form Status
func GetStatusFormFields() []form.Field {
	return []form.Field{
		{
			Name:        "Status_name",
			Label:       "Status Name",
			Placeholder: "Required",
			Required:    true,
			HelperText:  "Unique identifier for this Status",
			Width:       30,
		},
		{
			Name:        "Status_desc",
			Label:       "Status Desc",
			Placeholder: "Optional",
			Required:    false,
			HelperText:  "Description of the Status's purpose",
			Width:       30,
		},
	}
}
