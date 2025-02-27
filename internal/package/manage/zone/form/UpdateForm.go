package zoneForm

import (
	"dbaTools/internal/logger"
	zoneService "dbaTools/internal/package/manage/zone/service"
	"dbaTools/internal/ui/component/form"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// GetZoneUpdateFormFields mengembalikan definisi field form update zone
func GetZoneUpdateFormFields(zoneData zoneService.ZoneData) []form.Field {
	return []form.Field{
		{
			Name:        "zone_name",
			Label:       "Zone Name",
			Placeholder: "Required",
			Required:    true,
			HelperText:  "Unique identifier for this zone",
			Width:       30,
			Value:       zoneData.Name,
		},
		{
			Name:        "zone_desc",
			Label:       "Zone Desc",
			Placeholder: "Optional",
			Required:    false,
			HelperText:  "Description of the zone's purpose",
			Width:       30,
			Value:       zoneData.Description,
		},
	}
}

// GetZoneUpdateFormOptions mengembalikan opsi form untuk update zone
func GetZoneUpdateFormOptions(zoneId string) form.FormOptions {
	options := form.DefaultFormOptions()
	options.Title = "Update Zone"
	options.Description = fmt.Sprintf("Update Zone #%s information:", zoneId)
	options.SubmitLabel = "Update"
	return options
}

// showZoneUpdateForm menampilkan dan mengelola form
func ShowZoneUpdateForm(log *logger.Logger, zoneData zoneService.ZoneData) (map[string]string, error) {
	// Get form fields and options
	fields := GetZoneUpdateFormFields(zoneData)
	options := GetZoneUpdateFormOptions(zoneData.Id)

	// Create form model
	formModel := form.NewCustom(fields, form.DefaultStyles(), options)

	// Run the form program
	p := tea.NewProgram(formModel)
	finalModel, err := p.Run()
	if err != nil {
		log.Error("Form error", err)
		return nil, fmt.Errorf("error running form: %w", err)
	}

	// Cast ke form.Model untuk mengakses methods
	updatedForm, ok := finalModel.(form.Model)
	if !ok {
		return nil, fmt.Errorf("failed to get updated form model")
	}

	// Check if canceled
	if updatedForm.IsCanceled() {
		return nil, fmt.Errorf("canceled")
	}

	// Check if not submitted
	if !updatedForm.IsSubmitted() {
		return nil, fmt.Errorf("not submitted")
	}

	// Return values
	return updatedForm.GetValues(), nil
}
