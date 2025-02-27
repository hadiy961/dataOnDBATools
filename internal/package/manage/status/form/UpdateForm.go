package StatusForm

import (
	"dbaTools/internal/logger"
	StatusService "dbaTools/internal/package/manage/status/service"
	"dbaTools/internal/ui/component/form"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// GetStatusUpdateFormFields mengembalikan definisi field form update Status
func GetStatusUpdateFormFields(StatusData StatusService.StatusData) []form.Field {
	return []form.Field{
		{
			Name:        "Status_name",
			Label:       "Status Name",
			Placeholder: "Required",
			Required:    true,
			HelperText:  "Unique identifier for this Status",
			Width:       30,
			Value:       StatusData.Name,
		},
		{
			Name:        "Status_desc",
			Label:       "Status Desc",
			Placeholder: "Optional",
			Required:    false,
			HelperText:  "Description of the Status's purpose",
			Width:       30,
			Value:       StatusData.Description,
		},
	}
}

// GetStatusUpdateFormOptions mengembalikan opsi form untuk update Status
func GetStatusUpdateFormOptions(StatusId string) form.FormOptions {
	options := form.DefaultFormOptions()
	options.Title = "Update Status"
	options.Description = fmt.Sprintf("Update Status #%s information:", StatusId)
	options.SubmitLabel = "Update"
	return options
}

// showStatusUpdateForm menampilkan dan mengelola form
func ShowStatusUpdateForm(log *logger.Logger, StatusData StatusService.StatusData) (map[string]string, error) {
	// Get form fields and options
	fields := GetStatusUpdateFormFields(StatusData)
	options := GetStatusUpdateFormOptions(StatusData.Id)

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
