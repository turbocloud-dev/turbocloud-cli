package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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

	//New machine
	newMachineForm        *huh.Form
	newMachineJoinURLForm *huh.Form
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.HiddenBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(0, 2)

var (
	newMachineTypes string
	newMachineName  string
	newMachineIsAdd bool
)

type NewMachineJoinURLMsg int

func newMachineJoinURLMsg() tea.Msg {
	var msg NewMachineJoinURLMsg
	return msg
}

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

	model := model{
		list:         mainMenu,
		delegateKeys: delegateKeys,
		screenType:   1,
	}

	return model

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

	case NewMachineMsg:
		m.screenType = 3

		newMachineTypes = "" //newMachineTypes[:0]
		newMachineName = ""
		newMachineIsAdd = true
		m.newMachineForm = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose Machine Type").
					Options(
						huh.NewOption("Server", "lettuce"),
						huh.NewOption("Local Machine", "local_machine"),
					).
					Value(&newMachineTypes),

				huh.NewInput().
					Title("Machine Name").
					Value(&newMachineName).
					Validate(func(str string) error {
						if str == "Frank" {
							return errors.New("Sorry, we don’t serve customers named Frank.")
						}
						return nil
					}),
				huh.NewConfirm().
					Key("done").
					Title("Add a new machine?").
					Validate(func(v bool) error {
						if !v {
							m.screenType = 1
						}
						return nil
					}).
					Affirmative("Add").
					Negative("Cancel").
					Value(&newMachineIsAdd),
			),
		)

		m.newMachineForm.Init()

	case NewMachineJoinURLMsg:
		m.screenType = 4

		m.newMachineJoinURLForm = huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Connect a new machine to VPN").
					Description("• SSH into a server\n• Run a command:\n\n    curl turbocloud.dev/setup\n\n• After a server has finished provision you will see status Online near that machine in Machines\n\n").
					Next(true).
					NextLabel("OK"),
			),
		)

		m.newMachineJoinURLForm.Init()

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
		case "esc":
			if m.screenType == 2 || m.screenType == 3 {
				m.screenType = 1
				m.newMachineForm.State = huh.StateNormal
				return m, nil
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	}

	// This will also call our delegate's update function.
	if m.screenType == 1 {
		newListModel, cmd := m.list.Update(msg)
		m.list = newListModel
		cmds = append(cmds, cmd)

	}

	if m.screenType == 2 {
		m.machineList, cmd = m.machineList.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.screenType == 3 {
		form, cmd := m.newMachineForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.newMachineForm = f
		}

		cmds = append(cmds, cmd)

		if m.newMachineForm.State == huh.StateAborted {
			m.newMachineForm.State = huh.StateNormal
			m.screenType = 1
		} else if m.newMachineForm.State == huh.StateCompleted {
			m.newMachineForm.State = huh.StateNormal
			if newMachineIsAdd {
				cmds = append(cmds, newMachineJoinURLMsg)
			} else {
				m.screenType = 1
			}
		}

	}

	if m.screenType == 4 {
		form, cmd := m.newMachineJoinURLForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.newMachineJoinURLForm = f
		}

		cmds = append(cmds, cmd)

		if m.newMachineJoinURLForm.State == huh.StateCompleted {
			m.screenType = 1
		}

	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	switch m.screenType {
	case 1:
		return appStyle.Render(m.list.View())
	case 2:
		return breadhumbPositionStyle.Render(breadhumbStyle.Render("TurboCloud > Machines")) + topHintPositionStyle.Render(topHintStyle.Render("Press Enter to select a machine\nPress ← or ESC to return to main menu")) + baseStyle.Render(m.machineList.View()) + "\n  " + m.machineList.HelpView() + "\n"
	case 3:
		{
			if m.newMachineForm.State == huh.StateCompleted {

				//class := m.newMachineForm.GetString("class")
				//level := m.newMachineForm.GetString("level")
				//return fmt.Sprintf("You selected: %s, Lvl. %d", class, level)
			}
			return breadhumbPositionStyle.Render(breadhumbStyle.Render("TurboCloud > Add Machine")) + topHintPositionStyle.Render(topHintStyle.Render("Press X or Space to select options\nPress Enter to confirm\nPress ESC to return to main menu")) + baseStyle.Render(m.newMachineForm.View()) + "\n"
		}
	case 4:
		{
			if m.newMachineForm.State == huh.StateCompleted {

				//class := m.newMachineForm.GetString("class")
				//level := m.newMachineForm.GetString("level")
				//return fmt.Sprintf("You selected: %s, Lvl. %d", class, level)
			}
			return breadhumbPositionStyle.Render(breadhumbStyle.Render("TurboCloud > Add Machine")) + topHintPositionStyle.Render(topHintStyle.Render("Press ESC to return to main menu")) + baseStyle.Render(m.newMachineJoinURLForm.View()) + "\n"
		}
	}

	return ""

}

func main() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
