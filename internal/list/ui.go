package list

import (
	"fmt"
	"giles/internal/database"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type DeleteMsg struct {
	Path string
}

type TableModel struct {
	table     table.Model
	err       error
	fileStore *database.FileStore
	files     []database.File
}

func NewTableModel(files []database.File, store *database.FileStore) TableModel {
	columns := []table.Column{
		{Title: "Filename", Width: 30},
		{Title: "Path", Width: 50},
		{Title: "Type", Width: 20},
	}

	rows := make([]table.Row, len(files))
	for i, file := range files {
		rows[i] = table.Row{file.Name, file.Path}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	t.SetStyles(s)

	return TableModel{
		table:     t,
		fileStore: store,
		files:     files,
	}
}

func (m TableModel) Init() tea.Cmd {
	return nil
}

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "d":
			if len(m.files) > 0 {
				selectedRow := m.table.SelectedRow()
				filePath := selectedRow[1] // Path is in second column
				return m, func() tea.Msg {
					err := m.fileStore.DeleteFile(filePath)
					if err != nil {
						m.err = err
						return nil
					}
					// Refresh file list after deletion
					files, err := m.fileStore.GetAllFiles()
					if err != nil {
						m.err = err
						return nil
					}
					m.files = files
					rows := make([]table.Row, len(files))
					for i, file := range files {
						rows[i] = table.Row{file.Name, file.Path}
					}
					m.table.SetRows(rows)
					return nil
				}
			}
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m TableModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	return m.table.View() + "\nPress q to quit, d to delete selected file\n"
}
