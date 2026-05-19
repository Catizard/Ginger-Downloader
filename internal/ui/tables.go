package ui

import (
	"fmt"
	"log"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Catizard/Ginger-Downloader/internal/promise"
	"github.com/Catizard/bmstable"
)

type tablesModelState int

const (
	INPUT_TABLE_URL     = iota
	FETCHING_TABLE_DATA = iota
)

// TablesModel reads the user specified table's url
type TablesModel struct {
	textInput textinput.Model
	state     tablesModelState
	err       error
	ctx       *viewContext
	waitTable *promise.Await[bmstable.DifficultTable]
}

func InitializeTablesModel(ctx *viewContext) TablesModel {
	ti := textinput.New()
	// ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 256
	ti.SetWidth(128)
	return TablesModel{
		textInput: ti,
		ctx:       ctx,
		waitTable: promise.NewAwait[bmstable.DifficultTable](),
	}
}

func (m TablesModel) Init() tea.Cmd {
	return nil
}

func (m TablesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	select {
	case p := <-m.waitTable.Done:
		if p.Err != nil {
			m.err = p.Err
		} else {
			if p.Data == nil {
				log.Fatal("table data is empty")
			}
			m.ctx.candidateTable = *p.Data
			return m, tea.Sequence(newTransferRequest(SCENE_PREPARE))
		}
	default:
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "enter" {
			url := m.textInput.Value()
			if m.state == INPUT_TABLE_URL && url != "" {
				go m.fetchTableData(url)
				m.state = FETCHING_TABLE_DATA
			}
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m TablesModel) View() tea.View {
	s := lipgloss.JoinVertical(
		lipgloss.Top,
		"Input the url of the difficult table you want to download with",
		"Press enter to submit",
		m.textInput.View(),
	)
	if m.state == FETCHING_TABLE_DATA {
		s = lipgloss.JoinVertical(
			lipgloss.Top,
			s,
			"Fetching table data...Please wait :)",
		)
	} else if m.state == INPUT_TABLE_URL && m.err != nil {
		s = lipgloss.JoinVertical(
			lipgloss.Top,
			s,
			fmt.Sprintf("Failed to load table data: %v", m.err),
		)
	}
	return tea.NewView(s)
}

func (m TablesModel) fetchTableData(url string) {
	dt, err := bmstable.ParseFromURL(url)
	if err != nil {
		m.waitTable.Done <- promise.Fail[bmstable.DifficultTable](err)
		return
	}

	m.waitTable.Done <- promise.Ok(&dt)
}
