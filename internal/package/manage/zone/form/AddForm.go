package zoneForm

import (
	"dbaTools/internal/logger"
	"dbaTools/internal/ui/component/form"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// showZoneAddForm displays the form and returns the values if submitted
func ShowZoneAddForm(log *logger.Logger) (map[string]string, error) {
	// Get form fields and options
	fields := GetZoneFormFields()
	options := GetZoneAddFormOptions()

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

// GetZoneAddFormOptions returns form options for zone add operation
func GetZoneAddFormOptions() form.FormOptions {
	options := form.DefaultFormOptions()
	options.Title = "Add Zone"
	options.Description = "Enter new Zone information:"
	options.SubmitLabel = "Submit"
	return options
}

// ZoneFormData berisi definisi field untuk form zone
func GetZoneFormFields() []form.Field {
	return []form.Field{
		{
			Name:        "zone_name",
			Label:       "Zone Name",
			Placeholder: "Required",
			Required:    true,
			HelperText:  "Unique identifier for this zone",
			Width:       30,
		},
		{
			Name:        "zone_desc",
			Label:       "Zone Desc",
			Placeholder: "Optional",
			Required:    false,
			HelperText:  "Description of the zone's purpose",
			Width:       30,
		},
	}
}
