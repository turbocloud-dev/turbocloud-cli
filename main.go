package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle  = lipgloss.NewStyle().Padding(1, 2)
	listStyle = lipgloss.NewStyle().Padding(5, 4)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	breadhumbPositionStyle = lipgloss.NewStyle().
				Padding(1, 4)

	breadhumbStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	topHintPositionStyle = lipgloss.NewStyle().
				Padding(1, 3)

	topHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#bfbfbf")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type model struct {
	list         list.Model
	delegateKeys *delegateKeyMap
	machineMsg   MachineMsg
	machineList  table.Model
	screenType   int
	screenWidth  int
	screenHeight int
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.HiddenBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(0, 2)

func newModel() model {
	var (
		delegateKeys = newDelegateKeyMap()
	)

	// Make initial list of items
	//const numItems = 24
	items := []list.Item{
		item{title: "Getting Started", description: "How to deploy the first project"},
		item{title: "Add Machine", description: "Add a new server or local machine"},
		item{title: "Machines", description: "Manage servers and local machines"},
		item{title: "Add Service", description: "Deploy a new service"},
		item{title: "Services", description: "Deploy and manage services"},
		item{title: "Docs", description: "Detailed documentation and examples"},
	}

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	mainMenu := list.New(items, delegate, 0, 0)
	mainMenu.Title = "TurboCloud"
	mainMenu.Styles.Title = titleStyle
	mainMenu.SetShowStatusBar(false)

	return model{
		list:         mainMenu,
		delegateKeys: delegateKeys,
		screenType:   1,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

		h, v = listStyle.GetFrameSize()
		m.machineList.SetWidth(m.screenWidth - h)
		m.machineList.SetHeight(m.screenHeight - v)

	case MachineMsg:
		// The server returned a status message. Save it to our model. Also
		// tell the Bubble Tea runtime we want to exit because we have nothing
		// else to do. We'll still be able to render a final view with our
		// status message.
		m.machineMsg = msg
		m.screenType = 2

		//Reload machine list

		columns := []table.Column{
			{Title: "ID", Width: 10},
			{Title: "Name", Width: 20},
			{Title: "VPN Ip", Width: 18},
			{Title: "Public Ip", Width: 18},
			{Title: "Status", Width: 10},
			{Title: "CPU", Width: 10},
			{Title: "RAM", Width: 10},
			{Title: "Disk", Width: 10},
		}
		//{"1", "Tokyo", "Japan", "37,274,000"}
		rows := []table.Row{}

		for _, machine := range m.machineMsg {
			var tableRow []string
			tableRow = append(tableRow, machine.Id)
			tableRow = append(tableRow, machine.Name)
			tableRow = append(tableRow, machine.VPNIp)
			tableRow = append(tableRow, machine.PublicIp)
			tableRow = append(tableRow, machine.Status)
			tableRow = append(tableRow, "34%")
			tableRow = append(tableRow, "12GB")
			tableRow = append(tableRow, "24GB")

			rows = append(rows, tableRow)
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(20),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.HiddenBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("#bfbfbf")).
			Bold(true)
		s.Cell = s.Cell.Height(1)
		t.SetStyles(s)

		m.machineList = t
		h, v := listStyle.GetFrameSize()
		m.machineList.SetWidth(m.screenWidth - h)
		m.machineList.SetHeight(m.screenHeight - v)

		return m, nil

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {

		// These keys should exit the program.
		case "left", "esc":
			if m.screenType == 2 {
				m.screenType = 1
			}
			return m, nil
		}

	}

	// This will also call our delegate's update function.
	m.machineList, cmd = m.machineList.Update(msg)
	cmds = append(cmds, cmd)

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.screenType == 1 {
		return appStyle.Render(m.list.View())
	} else {
		return breadhumbPositionStyle.Render(breadhumbStyle.Render("TurboCloud > Machines")) + topHintPositionStyle.Render(topHintStyle.Render("Press Enter to select a machine\nPress ← or ESC to return to main menu")) + baseStyle.Render(m.machineList.View()) + "\n  " + m.machineList.HelpView() + "\n"
	}

}

func main() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
