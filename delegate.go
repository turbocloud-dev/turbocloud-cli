package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title string

		if i, ok := m.SelectedItem().(item); ok {
			title = i.Title()
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.choose):
				if title == "Getting Started" {
					openbrowser("https://turbocloud.dev/docs/getting-started")
					return nil
				} else if title == "Machines" {
					return getMachines
				} else if title == "Add Machine" {
					return newMachineMsg
				} else if title == "Services" {
					return getServices
				} else if title == "Add Service" {
					return newServiceMsg
				} else if title == "Docs" {
					openbrowser("https://turbocloud.dev/docs")
					return nil
				}
				return m.NewStatusMessage(statusMessageStyle("You chose " + title))

			}
		}

		return nil
	}

	help := []key.Binding{keys.choose}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

func newMachineMsg() tea.Msg {
	var msg NewMachineMsg
	return msg
}

type NewMachineMsg int

func newServiceMsg() tea.Msg {
	var msg NewServiceMsg
	return msg
}

type NewServiceMsg int

type delegateKeyMap struct {
	choose key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}
}
