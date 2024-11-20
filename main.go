package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Screen Types
const SCREEN_TYPE_ENVIRONMENTS = 6
const SCREEN_TYPE_NEW_ENVIRONMENT = 8
const SCREEN_TYPE_ENV_MENU = 12
const SCREEN_TYPE_ENV_DELETE_CONFIRMATION = 13
const SCREEN_TYPE_EDIT_ENVIRONMENT = 14
const SCREEN_TYPE_DEPLOYMENT_SCHEDULED = 15

// Strings
const ADD_ENVIRONMENT_STRING = "Add Environment"
const MENU_EDIT = "Edit / Details"
const MENU_DEPLOY = "Deploy"
const MENU_DELETE = "Delete"
const MENU_BACK = "Back"

var (
	appStyle         = lipgloss.NewStyle().Padding(1, 2)
	listStyle        = lipgloss.NewStyle().Padding(1, 4).Width(0).Height(0)
	listHelpStyle    = lipgloss.NewStyle().Padding(0, 4)
	listTopHintHeght = 11

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

	//New machine styles
	newMachineHintTitleStyle = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"})

	codeHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#e07a5f", Dark: "#e07a5f"})
)

type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

var screenType = 1

type model struct {
	list         list.Model
	delegateKeys *delegateKeyMap

	//Machines
	machineList table.Model

	//Services
	serviceList     table.Model
	selectedService Service

	//New service
	newServiceForm *huh.Form

	//Environments
	environmentList       table.Model
	envMenu               list.Model
	selectedEnvironment   Environment
	deleteEnvConfirmation textinput.Model

	//New environment
	newEnvironmentForm *huh.Form
	newEnvironmentHint string

	screenWidth  int
	screenHeight int

	//New machine
	newMachineForm     *huh.Form
	newMachineJoinHint string
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.HiddenBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(2, 0)

var (
	newMachineTypes string
	newMachineName  string
	newMachineIsAdd bool
)

var (
	newServiceName   string
	newServiceGitURL string
	newServiceIsAdd  bool
)

var (
	newEnvironmentName       string
	newEnvironmentBranchName string
	newEnvironmentIsAdd      bool
	newEnvironmentPort       string
	newEnvironmentDomain     string
	newEnvironmentMachines   []string
)

type TickMsg time.Time

type MainMenuMsg int

func mainMenuMsg() tea.Msg {
	var msg MainMenuMsg
	return msg
}

type NewMachineJoinURLMsg struct {
	newMachine Machine
}

type NewEnvironmentMsg struct {
	service Service
}

func newEnvironmentMsg(serviceId string, serviceName string) tea.Cmd {
	return func() tea.Msg {
		var msg NewEnvironmentMsg
		if serviceId == "" {
			msg.service = postService(newServiceName, newServiceGitURL)
		} else {
			msg.service.Id = serviceId
			msg.service.Name = serviceName

		}
		return msg
	}

}

type EditEnvironmentMsg struct {
	environment Environment
}

func editEnvironmentMsg(environmentId string, serviceId string) tea.Cmd {
	return func() tea.Msg {
		var environments = getEnvironments(serviceId).(EnvironmentsMsg)
		var msg EditEnvironmentMsg

		for _, environment := range environments {
			if environment.Id == environmentId {
				msg.environment = environment
			}
		}

		return msg
	}

}

type MenuEnvironmentMsg struct {
	environment Environment
}

func menuEnvironmentMsg(envId string, envName string) tea.Cmd {
	return func() tea.Msg {
		var msg MenuEnvironmentMsg
		msg.environment.Id = envId
		msg.environment.Name = envName

		return msg
	}

}

func newMachineJoinURLMsg() tea.Msg {
	var msg NewMachineJoinURLMsg
	//Send a request to create a new machine
	msg.newMachine = postMachine(newMachineName, newMachineTypes)

	return msg
}

func newModel() model {
	var (
		delegateKeys = newDelegateKeyMap()
	)

	// Make initial list of items
	items := []list.Item{
		item{title: "Getting Started", description: "How to deploy the first project"},
		item{title: "Add Machine", description: "Add a new server or local machine"},
		item{title: "Machines", description: "Manage servers and local machines"},
		item{title: "Add Service", description: "Deploy a new service"},
		item{title: "Services", description: "Deploy and manage services and environments"},
		item{title: "Docs", description: "Detailed documentation and examples"},
	}

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	mainMenu := list.New(items, delegate, 0, 0)
	mainMenu.Title = "TurboCloud"
	mainMenu.Styles.Title = titleStyle
	mainMenu.SetShowStatusBar(false)

	//Setup deleteEnvConfirmation
	deleteEnvConfirmation := textinput.New()
	deleteEnvConfirmation.Focus()
	deleteEnvConfirmation.CharLimit = 156
	deleteEnvConfirmation.Width = 20

	model := model{
		list:                  mainMenu,
		delegateKeys:          delegateKeys,
		deleteEnvConfirmation: deleteEnvConfirmation,
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

		v, _ = listStyle.GetFrameSize()
		m.machineList.SetWidth(m.screenWidth - 2*v)
		m.machineList.SetHeight(m.screenHeight - listTopHintHeght)

		if screenType == 3 {
			m.newMachineForm.WithWidth(m.screenWidth - 2*v)
			m.newMachineForm.WithHeight(m.screenHeight - listTopHintHeght)
		}

		m.serviceList.SetWidth(m.screenWidth - 2*v)
		m.serviceList.SetHeight(m.screenHeight - listTopHintHeght)

		m.environmentList.SetWidth(m.screenWidth - 2*v)
		m.environmentList.SetHeight(m.screenHeight - listTopHintHeght)

		if screenType == SCREEN_TYPE_ENV_MENU {
			v, _ := listStyle.GetFrameSize()
			m.envMenu.SetSize(m.screenWidth-2*v, m.screenHeight-listTopHintHeght)
		}

		ClearTerminal()

	case MainMenuMsg:
		screenType = 1

	case NewMachineMsg:
		screenType = 3

		newMachineTypes = "" //newMachineTypes[:0]
		newMachineName = ""
		newMachineIsAdd = true
		m.newMachineForm = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose Machine Type").
					Options(
						huh.NewOption("Server", "workload"),
						huh.NewOption("Local Machine", "local_machine"),
					).
					Value(&newMachineTypes),

				huh.NewInput().
					Title("Machine Name").
					Value(&newMachineName).
					Validate(func(str string) error {
						/*if str == "Frank" {
						}*/
						return nil
					}),
				huh.NewConfirm().
					Key("done").
					Title("Add a new machine?").
					Validate(func(v bool) error {
						if !v {
							screenType = 1
						}
						return nil
					}).
					Affirmative("Add").
					Negative("Cancel").
					Value(&newMachineIsAdd),
			),
		)

		m.newMachineForm.Init()

		v, _ := listStyle.GetFrameSize()

		m.newMachineForm.WithWidth(m.screenWidth - 2*v)
		m.newMachineForm.WithHeight(m.screenHeight - listTopHintHeght + 1)
		m.newMachineForm.Help()

	case NewServiceMsg:
		screenType = 7

		newServiceName = ""
		newServiceGitURL = ""
		newServiceIsAdd = true
		m.newServiceForm = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Service Name").
					Value(&newServiceName).
					Validate(func(str string) error {
						/*if str == "Frank" {
						}*/
						return nil
					}),
				huh.NewInput().
					Title("Git clone URL").
					Placeholder("get@...for private repos and https://... for public repos").
					Value(&newServiceGitURL).
					Validate(func(str string) error {
						/*if str == "Frank" {
						}*/
						return nil
					}),
				huh.NewConfirm().
					Key("done").
					Title("Add a new service?").
					Validate(func(v bool) error {
						if !v {
							screenType = 1
						}
						return nil
					}).
					Affirmative("Add").
					Negative("Cancel").
					Value(&newServiceIsAdd),
			),
		)

		m.newServiceForm.Init()

	case NewEnvironmentMsg:
		screenType = SCREEN_TYPE_NEW_ENVIRONMENT

		newEnvironmentName = ""
		newEnvironmentBranchName = ""
		newEnvironmentIsAdd = true
		newEnvironmentPort = ""
		newEnvironmentDomain = ""
		newEnvironmentMachines = newEnvironmentMachines[:0]

		m.selectedService.Id = msg.service.Id
		m.selectedService.Name = msg.service.Name

		machineOptions, _ := getMachineOptions()
		createEnvironmentDetails(&m, machineOptions, "Add a new environment?", "Add")

	case EditEnvironmentMsg:
		screenType = SCREEN_TYPE_EDIT_ENVIRONMENT
		machineOptions, machines := getMachineOptions()

		newEnvironmentMachines = newEnvironmentMachines[:0]
		for _, machine := range machines {
			if slices.Contains(msg.environment.MachineIds, machine.Id) {
				newEnvironmentMachines = append(newEnvironmentMachines, machine.Name)
			}
		}

		newEnvironmentName = msg.environment.Name
		newEnvironmentBranchName = msg.environment.Branch
		newEnvironmentIsAdd = true
		newEnvironmentPort = msg.environment.Port
		newEnvironmentDomain = msg.environment.Domains[0]

		createEnvironmentDetails(&m, machineOptions, "Save environment details?", "Save")

	case NewEnvironmentAddedMsg:
		screenType = 9

		//Get the first builder machine
		machines := getMachines().(MachineMsg) //type asssertion
		var machineBuilder Machine
		for _, machine := range machines {
			if slices.Contains(machine.Types, "builder") {
				machineBuilder = machine
				break
			}
		}

		if len(machineBuilder.Domains) == 0 {
			m.newEnvironmentHint = "Before deploying this environment you should add at least one domain to the builder machine. Contact us at hey@turbocloud.dev iff you don't know how to do that."
		} else {

			sshKeyHint := "    To allow cloning the git repository from your build machine, you should add public SSH key below to GitHub, Bitbucket repository access/deploy keys (only read permission is required):\n\n" + codeHintStyle.Render(strings.Replace(machineBuilder.PublicSSHKey, "\n", "", -1)) + "\n\n"
			webhookHint := "    To deploy after each Git push to a remote repository automatically, you should add a webhook below to GitHub (don't forget to select application/json in the Content-Type dropdown), Bitbucket repository webhooks:\n\n" + codeHintStyle.Render("https://"+machineBuilder.Domains[0]+"/deploy/"+m.selectedService.Id)

			//This public SSH key also can be found if ssh into your build machine (usually the first server you provisioned in this project) and run 'cat ~/.ssh/id_rsa.pub'`
			m.newEnvironmentHint = sshKeyHint + webhookHint + "\n\n    Options to deploy:\n\n    • Push any changes to the branch you specified in the previous step.\n    • Send a POST request to https://" + machineBuilder.Domains[0] + "/deploy/environment/" + msg.Id + "\n\n    To manage environments, go to Services and select a service from the list.\n\n"
		}

		cmd := tea.Tick(100*time.Microsecond, func(t time.Time) tea.Msg {
			return TickMsg(t)
		})
		cmds = append(cmds, cmd)

	case NewMachineJoinURLMsg:

		screenType = 4

		m.newMachineJoinHint = "    • SSH into the new machine\n    • Copy and run the following command (shown only once):\n\n" + codeHintStyle.Render("    curl https://turbocloud.dev/setup | bash -s -- -j https://"+msg.newMachine.JoinURL) + "\n\n    • Once provisioning is complete, the status will show as 'Online' next to the machine in the Machines list.\n\n"
		cmd := tea.Tick(100*time.Microsecond, func(t time.Time) tea.Msg {
			return TickMsg(t)
		})
		cmds = append(cmds, cmd)

	case MachineMsg:
		// The server returned a status message. Save it to our model. Also
		// tell the Bubble Tea runtime we want to exit because we have nothing
		// else to do. We'll still be able to render a final view with our
		// status message.
		screenType = 2

		//Reload machine list
		selectedRow := m.machineList.SelectedRow()
		indexToSelect := 0

		columns := []table.Column{
			{Title: "ID", Width: 10},
			{Title: "Name", Width: 20},
			{Title: "VPN Ip", Width: 18},
			{Title: "Public Ip", Width: 18},
			{Title: "Status", Width: 10},
			{Title: "CPU(%)", Width: 8},
			{Title: "RAM(MB)", Width: 9},
			{Title: "Disk(MB)", Width: 9},
		}
		//{"1", "Tokyo", "Japan", "37,274,000"}
		rows := []table.Row{}

		for index, machine := range msg {

			if selectedRow != nil && selectedRow[0] == machine.Id {
				indexToSelect = index
			}
			var tableRow []string
			tableRow = append(tableRow, machine.Id)
			tableRow = append(tableRow, machine.Name)
			tableRow = append(tableRow, machine.VPNIp)
			tableRow = append(tableRow, machine.PublicIp)
			tableRow = append(tableRow, machine.Status)
			tableRow = append(tableRow, machine.CPUUsage)
			tableRow = append(tableRow, machine.MEMUsage)
			tableRow = append(tableRow, machine.DiskUsage)

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

		v, _ := listStyle.GetFrameSize()
		m.machineList.SetWidth(m.screenWidth - 2*v)
		m.machineList.SetHeight(m.screenHeight - listTopHintHeght)

		m.machineList.MoveDown(indexToSelect)
		cmd := tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			if screenType == 2 {
				return getMachines()
			} else {
				return TickMsg(t)
			}
		})
		cmds = append(cmds, cmd)

	case ServicesMsg:
		// The server returned a status message. Save it to our model. Also
		// tell the Bubble Tea runtime we want to exit because we have nothing
		// else to do. We'll still be able to render a final view with our
		// status message.
		screenType = 5

		//Reload machine list

		columns := []table.Column{
			{Title: "ID", Width: 8},
			{Title: "Name", Width: 16},
			{Title: "GitURL", Width: 50},
		}
		//{"1", "Tokyo", "Japan", "37,274,000"}
		rows := []table.Row{}

		for _, service := range msg {
			var tableRow []string
			tableRow = append(tableRow, service.Id)
			tableRow = append(tableRow, service.Name)
			tableRow = append(tableRow, service.GitURL)

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

		m.serviceList = t
		v, _ := listStyle.GetFrameSize()
		m.serviceList.SetWidth(m.screenWidth - 2*v)
		m.serviceList.SetHeight(m.screenHeight - listTopHintHeght)

	case EnvironmentsMsg:
		// The server returned a status message. Save it to our model. Also
		// tell the Bubble Tea runtime we want to exit because we have nothing
		// else to do. We'll still be able to render a final view with our
		// status message.
		screenType = SCREEN_TYPE_ENVIRONMENTS

		//Reload machine list

		columns := []table.Column{
			{Title: "ID", Width: 15},
			{Title: "Name", Width: 16},
			{Title: "Branch", Width: 16},
		}
		rows := []table.Row{}

		var tableRow []string
		tableRow = append(tableRow, ADD_ENVIRONMENT_STRING)
		tableRow = append(tableRow, "")
		tableRow = append(tableRow, "")

		rows = append(rows, tableRow)

		for _, environment := range msg {
			var tableRow []string
			tableRow = append(tableRow, environment.Id)
			tableRow = append(tableRow, environment.Name)
			tableRow = append(tableRow, environment.Branch)

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

		m.environmentList = t
		v, _ := listStyle.GetFrameSize()
		m.environmentList.SetWidth(m.screenWidth - 2*v)
		m.environmentList.SetHeight(m.screenHeight - listTopHintHeght)

		return m, nil

	case MenuEnvironmentMsg:
		screenType = SCREEN_TYPE_ENV_MENU

		envMenuItems := []list.Item{
			item{title: MENU_DEPLOY, description: ""},
			item{title: MENU_EDIT, description: ""},
			item{title: MENU_DELETE, description: ""},
			item{title: MENU_BACK, description: ""},
		}

		m.envMenu = list.New(envMenuItems, envMenuItemDelegate{}, defaultWidth, listHeight)
		m.envMenu.SetShowStatusBar(false)
		m.envMenu.SetFilteringEnabled(false)
		m.envMenu.SetShowHelp(false)
		m.envMenu.SetShowTitle(false)

		v, _ := listStyle.GetFrameSize()
		m.envMenu.SetSize(m.screenWidth-2*v, m.screenHeight-listTopHintHeght)

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {

		// These keys should exit the program.
		case "esc":
			if screenType == 2 || screenType == 3 || screenType == 5 {
				screenType = 1
				if m.newMachineForm != nil {
					m.newMachineForm.State = huh.StateNormal
				}
				return m, nil
			}
			if screenType == SCREEN_TYPE_ENVIRONMENTS {
				screenType = 5
				return m, nil
			}
			if screenType == 7 {
				if m.newServiceForm != nil {
					m.newServiceForm.State = huh.StateNormal
				}
				screenType = 1
				return m, nil
			}
			if screenType == SCREEN_TYPE_NEW_ENVIRONMENT {
				if m.newEnvironmentForm != nil {
					m.newEnvironmentForm.State = huh.StateNormal
				}
				screenType = 1
				return m, nil
			}
			if screenType == SCREEN_TYPE_ENV_MENU {
				screenType = SCREEN_TYPE_ENVIRONMENTS
				return m, nil
			} else if screenType == SCREEN_TYPE_ENV_DELETE_CONFIRMATION {
				screenType = SCREEN_TYPE_ENV_MENU
				return m, nil
			}
			if screenType == SCREEN_TYPE_EDIT_ENVIRONMENT {
				screenType = SCREEN_TYPE_ENVIRONMENTS
				return m, nil
			}
		case "left":
			if screenType == 2 || screenType == 5 {
				screenType = 1
				return m, nil
			}
			if screenType == SCREEN_TYPE_ENVIRONMENTS {
				screenType = 5
				return m, nil
			}
			if screenType == SCREEN_TYPE_ENV_MENU {
				screenType = SCREEN_TYPE_ENVIRONMENTS
				return m, nil
			}
		case "enter":
			if screenType == 5 {
				//A service has been selected
				m.selectedService.Id = m.serviceList.SelectedRow()[0]
				m.selectedService.Name = m.serviceList.SelectedRow()[1]
				return m, getEnvironmentsCmd(m.selectedService.Id)
			} else if screenType == SCREEN_TYPE_ENVIRONMENTS {
				//A new environment has been selected
				if m.environmentList.SelectedRow()[0] == ADD_ENVIRONMENT_STRING {
					cmds = append(cmds, newEnvironmentMsg(m.serviceList.SelectedRow()[0], m.serviceList.SelectedRow()[1]))
				} else {
					cmds = append(cmds, menuEnvironmentMsg(m.environmentList.SelectedRow()[0], m.environmentList.SelectedRow()[1]))
					m.selectedEnvironment.Id = m.environmentList.SelectedRow()[0]
					m.selectedEnvironment.Name = m.environmentList.SelectedRow()[1]
				}

			} else if screenType == SCREEN_TYPE_ENV_MENU {
				//A new environment has been selected
				if m.envMenu.SelectedItem().(item).title == MENU_BACK {
					//Go back
					screenType = SCREEN_TYPE_ENVIRONMENTS
					return m, nil
				} else if m.envMenu.SelectedItem().(item).title == MENU_DEPLOY {
					//Deploy
					screenType = SCREEN_TYPE_DEPLOYMENT_SCHEDULED
					deployEnvironment(m.selectedEnvironment.Id)
					return m, nil
				} else if m.envMenu.SelectedItem().(item).title == MENU_EDIT {
					//Edit environment
					cmds = append(cmds, editEnvironmentMsg(m.selectedEnvironment.Id, m.selectedService.Id))
				} else if m.envMenu.SelectedItem().(item).title == MENU_DELETE {
					//Delete environment
					m.deleteEnvConfirmation.SetValue("")
					m.deleteEnvConfirmation.Focus()
					screenType = SCREEN_TYPE_ENV_DELETE_CONFIRMATION
					return m, nil
				}

			} else if screenType == SCREEN_TYPE_ENV_DELETE_CONFIRMATION {
				//A new environment has been selected
				if strings.ToLower(m.deleteEnvConfirmation.Value()) == "y" {
					deleteEnvironment(m.selectedEnvironment.Id)
					return m, getEnvironmentsCmd(m.selectedService.Id)
				}

			} else if screenType == SCREEN_TYPE_DEPLOYMENT_SCHEDULED {
				screenType = SCREEN_TYPE_ENVIRONMENTS
				return m, nil
			}
		case "ctrl+c":
			return m, tea.Quit

		}

	}

	// This will also call our delegate's update function.
	if screenType == 1 {
		newList, cmd := m.list.Update(msg)
		m.list = newList
		cmds = append(cmds, cmd)

	}

	if screenType == 2 {
		m.machineList, cmd = m.machineList.Update(msg)
		cmds = append(cmds, cmd)
	}

	if screenType == 5 {
		m.serviceList, cmd = m.serviceList.Update(msg)
		cmds = append(cmds, cmd)
	}

	if screenType == SCREEN_TYPE_ENVIRONMENTS {
		m.environmentList, cmd = m.environmentList.Update(msg)
		cmds = append(cmds, cmd)
	}

	if screenType == 3 {
		form, cmd := m.newMachineForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.newMachineForm = f
		}

		cmds = append(cmds, cmd)

		if m.newMachineForm.State == huh.StateAborted {
			m.newMachineForm.State = huh.StateNormal
			screenType = 1
		} else if m.newMachineForm.State == huh.StateCompleted {
			m.newMachineForm.State = huh.StateNormal
			if newMachineIsAdd {
				cmds = append(cmds, newMachineJoinURLMsg)
			} else {
				screenType = 1
			}
		}

	}

	if screenType == 7 {
		form, cmd := m.newServiceForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.newServiceForm = f
		}

		cmds = append(cmds, cmd)

		if m.newServiceForm.State == huh.StateAborted {
			m.newServiceForm.State = huh.StateNormal
			screenType = 1
		} else if m.newServiceForm.State == huh.StateCompleted {
			m.newServiceForm.State = huh.StateNormal
			if newServiceIsAdd {
				cmds = append(cmds, newEnvironmentMsg("", ""))
			} else {
				screenType = 1
			}
		}

	}

	if screenType == SCREEN_TYPE_NEW_ENVIRONMENT || screenType == SCREEN_TYPE_EDIT_ENVIRONMENT {
		form, cmd := m.newEnvironmentForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.newEnvironmentForm = f
		}

		cmds = append(cmds, cmd)

		if screenType == SCREEN_TYPE_NEW_ENVIRONMENT {
			if m.newEnvironmentForm.State == huh.StateAborted {
				m.newEnvironmentForm.State = huh.StateNormal
				screenType = 1
			} else if m.newEnvironmentForm.State == huh.StateCompleted {
				m.newEnvironmentForm.State = huh.StateNormal
				if newEnvironmentIsAdd {
					//Send a request to create a new environment
					var newEnvironment Environment
					newEnvironment.ServiceId = m.selectedService.Id
					newEnvironment.Name = newEnvironmentName
					newEnvironment.Branch = newEnvironmentBranchName
					newEnvironment.Domains = append(newEnvironment.Domains, newEnvironmentDomain)
					newEnvironment.Port = newEnvironmentPort
					newEnvironment.GitTag = ""
					//Get Machine Ids to deploy
					machines := getMachines().(MachineMsg) //type asssertion
					machineIds := []string{}

					for _, machine := range machines {
						if slices.Contains(newEnvironmentMachines, machine.Name) {
							machineIds = append(machineIds, machine.Id)
						}
					}
					newEnvironment.MachineIds = machineIds
					cmds = append(cmds, postEnvironment(newEnvironment))

				} else {
					screenType = 1
				}
			}
		} else {
			if m.newEnvironmentForm.State == huh.StateAborted {
				m.newEnvironmentForm.State = huh.StateNormal
				screenType = SCREEN_TYPE_ENV_MENU
			} else if m.newEnvironmentForm.State == huh.StateCompleted {
				m.newEnvironmentForm.State = huh.StateNormal
				if newEnvironmentIsAdd {
					//Send a request to create a new environment
					var editedEnvironment Environment
					editedEnvironment.Id = m.selectedEnvironment.Id
					editedEnvironment.Name = newEnvironmentName
					editedEnvironment.Branch = newEnvironmentBranchName
					editedEnvironment.Domains = append(editedEnvironment.Domains, newEnvironmentDomain)
					editedEnvironment.Port = newEnvironmentPort
					editedEnvironment.GitTag = ""
					//Get Machine Ids to deploy
					machines := getMachines().(MachineMsg) //type asssertion
					machineIds := []string{}

					for _, machine := range machines {
						if slices.Contains(newEnvironmentMachines, machine.Name) {
							machineIds = append(machineIds, machine.Id)
						}
					}
					editedEnvironment.MachineIds = machineIds
					updateEnvironment(editedEnvironment)
					cmds = append(cmds, getEnvironmentsCmd(m.selectedService.Id))

				} else {
					screenType = SCREEN_TYPE_ENV_MENU
				}
			}
		}

	}

	if screenType == SCREEN_TYPE_ENV_MENU {
		newList, cmd := m.envMenu.Update(msg)
		m.envMenu = newList
		cmds = append(cmds, cmd)

	}

	if screenType == SCREEN_TYPE_ENV_DELETE_CONFIRMATION {
		m.deleteEnvConfirmation, cmd = m.deleteEnvConfirmation.Update(msg)
		cmds = append(cmds, cmd)

	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	switch screenType {
	case 1:
		return appStyle.Render(m.list.View())
	case 2:
		return breadhumbPositionStyle.Render(breadhumbStyle.Render("Machines")) + topHintPositionStyle.Render(topHintStyle.Render("Press Enter to select a machine\nPress ← or ESC to return to main menu")) + listStyle.Render(m.machineList.View()) + "\n\n\n" + listHelpStyle.Render(m.machineList.HelpView()) + "\n"
	case 3:
		{
			return breadhumbPositionStyle.Render(breadhumbStyle.Render("Add Machine")) + topHintPositionStyle.Render(topHintStyle.Render("Press X or Space to select options\nPress Enter to confirm\nPress ESC to return to main menu")) + listStyle.Render(m.newMachineForm.View()) + "\n"
		}
	case 4:
		{
			screenType = 10

			return breadhumbPositionStyle.Render(breadhumbStyle.Render("Connect a new machine to VPN")) + "\n"
		}
	case 5:
		return breadhumbPositionStyle.Render(breadhumbStyle.Render("Services")) + topHintPositionStyle.Render(topHintStyle.Render("Press Enter to select a service\nPress ← or ESC to return to main menu")) + listStyle.Render(m.serviceList.View()) + "\n\n\n" + listHelpStyle.Render(m.serviceList.HelpView()) + "\n"
	case 6:
		return breadhumbPositionStyle.Render(breadhumbStyle.Render("Services > "+m.selectedService.Name)) + topHintPositionStyle.Render(topHintStyle.Render("Press Enter to add or select an environment\nPress ← or ESC to return to Services")) + listStyle.Render(m.environmentList.View()) + "\n\n\n" + listHelpStyle.Render(m.environmentList.HelpView()) + "\n"
	case 7:
		{
			return breadhumbPositionStyle.Render(breadhumbStyle.Render("Add Service")) + topHintPositionStyle.Render(topHintStyle.Render("Press X or Space to select options\nPress Enter to confirm\nPress ESC to return to main menu")) + baseStyle.Render(m.newServiceForm.View()) + "\n"
		}

	case SCREEN_TYPE_NEW_ENVIRONMENT:
		{
			return breadhumbPositionStyle.Render(breadhumbStyle.Render("Add Environment")) + topHintPositionStyle.Render(topHintStyle.Render("Press X or Space to select options\nPress Enter to confirm\nPress ESC to return to main menu")) + baseStyle.Render(m.newEnvironmentForm.View()) + "\n"
		}
	case 9:
		{
			screenType = 11

			return breadhumbPositionStyle.Render(breadhumbStyle.Render("Environment has been added")) + "\n"
		}
	case 10:
		{

			//tea.ClearScreen()
			app.ReleaseTerminal()
			fmt.Print(m.newMachineJoinHint)

			ok := YesNoPrompt(newMachineHintTitleStyle.Render("\n    Press Enter to return to Main Menu"), true)
			if ok {
				screenType = 1
				app.RestoreTerminal()
			} else {
				app.RestoreTerminal()
			}

		}
	case 11:
		{

			//tea.ClearScreen()
			app.ReleaseTerminal()
			fmt.Print(m.newEnvironmentHint)

			ok := YesNoPrompt(newMachineHintTitleStyle.Render("\n    Press Enter to return to Main Menu"), true)
			if ok {
				screenType = 1
				app.RestoreTerminal()
			} else {
				app.RestoreTerminal()
			}

		}
	case SCREEN_TYPE_ENV_MENU:
		return breadhumbPositionStyle.Render(breadhumbStyle.Render("Services > "+m.selectedService.Name+" > "+m.selectedEnvironment.Name)) + topHintPositionStyle.Render(topHintStyle.Render("Press Enter to select\nPress ← or ESC to return to Environments")) + listStyle.Render(m.envMenu.View()) + "\n"

	case SCREEN_TYPE_ENV_DELETE_CONFIRMATION:
		return breadhumbPositionStyle.Render(breadhumbStyle.Render("Services > "+m.selectedService.Name+" > "+m.selectedEnvironment.Name)) + topHintPositionStyle.Render(fmt.Sprintf(
			"\n Do you really want to delete this environment? Type 'y' to confirm or press ESC to cancel.\n\n %s\n\n %s",
			m.deleteEnvConfirmation.View(),
			"(esc to quit)"))

	case SCREEN_TYPE_EDIT_ENVIRONMENT:
		{
			return breadhumbPositionStyle.Render(breadhumbStyle.Render("Edit Environment")) + topHintPositionStyle.Render(topHintStyle.Render("Press X or Space to select options\nPress Enter to confirm\nPress ESC to return to main menu")) + baseStyle.Render(m.newEnvironmentForm.View()) + "\n"
		}
	case SCREEN_TYPE_DEPLOYMENT_SCHEDULED:
		{
			return breadhumbPositionStyle.Render(breadhumbStyle.Render("Services > "+m.selectedService.Name+" > "+m.selectedEnvironment.Name)) + topHintPositionStyle.Render(newMachineHintTitleStyle.Render("\n Deployment is scheduled. \n Press Enter to dismiss this message")) + "\n"
		}
	}
	return ""
}

var app *tea.Program

func runCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func ClearTerminal() {
	switch runtime.GOOS {
	case "darwin":
		runCmd("clear")
	case "linux":
		runCmd("clear")
	case "windows":
		runCmd("cmd", "/c", "cls")
	default:
		runCmd("clear")
	}
}

func main() {

	ClearTerminal()

	//executeScriptString("lsof -i tcp:5445 | awk 'NR!=1 {print $2}' | xargs kill\nssh -o ExitOnForwardFailure=yes -f -N -L 5445:localhost:5445 root@188.245.224.58")

	app = tea.NewProgram(newModel() /*, tea.WithAltScreen()*/)

	if _, err := app.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
