package tables

import (
	"dbaTools/internal/utils"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TableColumn struct {
	Title       string
	Width       int
	HeaderStyle *lipgloss.Style
	Truncate    bool // tambahan flag untuk mengontrol truncate per kolom
}

type TableRow struct {
	ID     string
	Values []string
}

type Table struct {
	Columns         []TableColumn
	Title           string
	Rows            []TableRow
	SelectedRow     int
	ShowMenu        bool
	MenuOption      int
	Width           int
	UpdateFunction  func(id string) // Fungsi update yang menerima id
	DeleteFunction  func(id string) // Fungsi delete yang menerima id
	Detail          func(id string) // Fungsi delete yang menerima id
	AddFunction     func()          // Fungsi untuk menambahkan data baru
	MenuManager     interface{}
	ShutdownHandler interface{}
	NavigationText  string   // Custom navigation text
	MenuOptions     []string // Custom menu options
	BackFunction    func()   // Function for back action
}

var (
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("8"))
)

var (
	menuItemStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginRight(1).
			Background(lipgloss.Color("8")).
			Foreground(lipgloss.Color("15"))

	selectedMenuItemStyle = lipgloss.NewStyle().
				Padding(0, 1).
				MarginRight(1).
				Bold(true).
				Background(lipgloss.Color("63"))

	menuBoxStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8"))
)

// TableOption is a function type for configuring the table
type TableOption func(*Table)

func WithUpdateFunction(updateFunc func(id string)) TableOption {
	return func(t *Table) {
		t.UpdateFunction = updateFunc
	}
}

func WithDetailFunction(detailfunc func(id string)) TableOption {
	return func(t *Table) {
		t.Detail = detailfunc
	}
}

func WithDeleteFunction(deleteFunc func(id string)) TableOption {
	return func(t *Table) {
		t.DeleteFunction = deleteFunc
	}
}

// Tambahkan option function untuk mengatur AddFunction
func WithAddFunction(addFunc func()) TableOption {
	return func(t *Table) {
		t.AddFunction = addFunc
	}
}

// Add new option functions
func WithCustomNavigation(text string) TableOption {
	return func(t *Table) {
		t.NavigationText = text
	}
}

func WithBackFunction(backFunc func()) TableOption {
	return func(t *Table) {
		t.BackFunction = backFunc
	}
}

func WithMenuOptions(options []string) TableOption {
	return func(t *Table) {
		t.MenuOptions = options
	}
}

// Modifikasi juga di constructor
func NewTable(columns []TableColumn, opts ...TableOption) *Table {
	t := &Table{
		Columns:        columns,
		Rows:           make([]TableRow, 0),
		SelectedRow:    -1,
		MenuOptions:    []string{"Detail", "Update", "Delete", "Cancel"},   // Default options
		NavigationText: "↑/↓: Navigate • Enter: Select • a: Add • q: Quit", // Default navigation
	}

	// Set default styles untuk header
	for i := range columns {
		if columns[i].HeaderStyle == nil {
			style := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15"))
			columns[i].HeaderStyle = &style
		}
		// Set initial width ke 0 agar dihitung secara dinamis
		columns[i].Width = 0
	}

	// Apply options
	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *Table) AddRow(id string, values []string) {
	t.Rows = append(t.Rows, TableRow{ID: id, Values: values})
}

func (t *Table) Init() tea.Cmd {
	t.calculateDynamicWidths()
	return nil
}

func (t *Table) calculateDynamicWidths() {
	// Inisialisasi width dengan panjang header
	maxWidths := make([]int, len(t.Columns))
	for i, col := range t.Columns {
		maxWidths[i] = len(col.Title)
	}

	// Cek panjang maksimum untuk setiap kolom dari data
	for _, row := range t.Rows {
		for i, val := range row.Values {
			if i < len(maxWidths) && len(val) > maxWidths[i] {
				maxWidths[i] = len(val)
			}
		}
	}

	// Hitung total width yang tersedia
	size, err := utils.GetTerminalSize()
	terminalWidth := 80 // default width
	if err == nil {
		terminalWidth = size.Width
	}

	// Kurangi space untuk border dan separator
	availableWidth := terminalWidth - (len(t.Columns)-1)*3 - 4 // 3 untuk separator " | ", 4 untuk margin

	// Jika total width melebihi terminal, sesuaikan proportionally
	totalWidth := 0
	for _, width := range maxWidths {
		totalWidth += width
	}

	if totalWidth > availableWidth {
		// Sesuaikan width secara proporsional
		for i := range maxWidths {
			maxWidths[i] = (maxWidths[i] * availableWidth) / totalWidth
			if maxWidths[i] < 3 { // minimal width
				maxWidths[i] = 3
			}
		}
	}

	// Update column widths
	for i := range t.Columns {
		t.Columns[i].Width = maxWidths[i]
	}
}

func (t *Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "b": // Add back key handler
			if t.BackFunction != nil {
				t.BackFunction()
				return t, tea.Quit
			}
			// ... rest of the key handlers ...
		case "q", "esc", "ctrl+c":
			return t, tea.Quit
		case "up":
			if t.ShowMenu {
				t.MenuOption--
				if t.MenuOption < 0 {
					t.MenuOption = len(t.MenuOptions) - 1
				}
			} else if t.SelectedRow > 0 {
				t.SelectedRow--
			}
		case "down":
			if t.ShowMenu {
				t.MenuOption++
				if t.MenuOption >= len(t.MenuOptions) { // Change to use length of menu options
					t.MenuOption = 0
				}
			} else if t.SelectedRow < len(t.Rows)-1 {
				t.SelectedRow++
			}
		case "enter":
			if t.SelectedRow >= 0 {
				if t.ShowMenu {
					switch t.MenuOptions[t.MenuOption] { // Use menu option text instead of index
					case "Detail":
						if t.Detail != nil && t.SelectedRow >= 0 && t.SelectedRow < len(t.Rows) {
							selectedID := t.Rows[t.SelectedRow].ID
							t.Detail(selectedID)
							return t, tea.Quit
						}
					case "Update":
						if t.UpdateFunction != nil && t.SelectedRow >= 0 && t.SelectedRow < len(t.Rows) {
							selectedID := t.Rows[t.SelectedRow].ID
							t.UpdateFunction(selectedID)
							return t, tea.Quit
						}
					case "Delete":
						if t.DeleteFunction != nil && t.SelectedRow >= 0 && t.SelectedRow < len(t.Rows) {
							selectedID := t.Rows[t.SelectedRow].ID
							t.DeleteFunction(selectedID)
							return t, tea.Quit
						}
					case "Cancel":
						t.ShowMenu = false
					}
				} else {
					t.ShowMenu = true
					t.MenuOption = 0
				}
			}
		case "a":
			// Panggil AddFunction jika tersedia
			if t.AddFunction != nil {
				t.AddFunction()    // Panggil fungsi add
				return t, tea.Quit // Keluar dari UI tabel
			}
			return t, nil
		}
	}
	return t, nil
}

// Tambahkan option untuk set title
func WithTitle(title string) TableOption {
	return func(t *Table) {
		t.Title = title
	}
}

func (t *Table) View() string {
	t.calculateDynamicWidths()
	var b strings.Builder

	// Hitung total lebar untuk divider
	dividerWidth := 1
	for i, col := range t.Columns {
		dividerWidth += col.Width
		if i < len(t.Columns)-1 {
			dividerWidth += 3 // space untuk " | "
		}
	}
	dividerWidth += 1
	divider := strings.Repeat("-", dividerWidth)

	// Tampilkan judul jika ada
	if t.Title != "" {
		b.WriteString("\n")
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
		b.WriteString(" " + titleStyle.Render(t.Title) + "\n")
	}

	b.WriteString(divider + "\n")
	b.WriteString(t.renderHeader())
	b.WriteString("\n" + divider + "\n")

	for i, row := range t.Rows {
		if i == t.SelectedRow {
			b.WriteString(t.renderSelectedRow(row))
		} else {
			b.WriteString(t.renderRow(row))
		}
		b.WriteString("\n")
	}
	b.WriteString(divider)

	if t.ShowMenu && t.SelectedRow >= 0 {
		b.WriteString(t.renderMenu())
	}

	// Use custom navigation text if provided
	b.WriteString("\n" + t.NavigationText + "\n")

	return b.String()
}

// Modifikasi fungsi renderHeader untuk menggunakan truncate yang baru
func (t *Table) renderHeader() string {
	var parts []string
	for _, col := range t.Columns {
		title := truncateString(col.Title, col.Width, col.Truncate)
		parts = append(parts, col.HeaderStyle.Render(title))
	}
	return " " + strings.Join(parts, " | ") + " "
}

// Fungsi helper untuk truncate yang lebih baik
func truncateString(s string, width int, shouldTruncate bool) string {
	if !shouldTruncate || len(s) <= width {
		return fmt.Sprintf("%-*s", width, s) // padding dengan spasi jika tidak perlu truncate
	}

	if width <= 3 {
		return s[:width]
	}

	// Truncate dengan elipsis
	return fmt.Sprintf("%-*s", width, s[:width-3]+"...")
}

// Modifikasi fungsi renderRow untuk menggunakan truncate yang baru
func (t *Table) renderRow(row TableRow) string {
	var parts []string
	for i, val := range row.Values {
		part := truncateString(val, t.Columns[i].Width, t.Columns[i].Truncate)
		parts = append(parts, part)
	}
	return " " + strings.Join(parts, " | ") + " "
}

func (t *Table) renderSelectedRow(row TableRow) string {
	var parts []string
	for i, val := range row.Values {
		part := fmt.Sprintf("%-*s", t.Columns[i].Width, truncateString(val, t.Columns[i].Width, t.Columns[i].Truncate))
		parts = append(parts, part)
	}
	return selectedStyle.Render(" " + strings.Join(parts, " | ") + " ")
}

func (t *Table) renderMenu() string {
	var menu strings.Builder
	menu.WriteString("\n")

	var menuItems []string
	options := t.MenuOptions // Use custom menu options

	for i, opt := range options {
		if i == t.MenuOption {
			menuItems = append(menuItems, selectedMenuItemStyle.Render(opt))
		} else {
			menuItems = append(menuItems, menuItemStyle.Render(opt))
		}
	}

	menu.WriteString(menuBoxStyle.Render(strings.Join(menuItems, " ")))
	return menu.String()
}

func RunTable(table *Table) error {
	p := tea.NewProgram(table)
	_, err := p.Run()
	return err
}

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
