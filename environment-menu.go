package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	menuItemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	menuSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("170"))
)

type envMenuItemDelegate struct{}

func (d envMenuItemDelegate) Height() int                             { return 1 }
func (d envMenuItemDelegate) Spacing() int                            { return 0 }
func (d envMenuItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d envMenuItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	//str := fmt.Sprintf("%s", i.title)

	fn := menuItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return menuSelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(i.title))
}

func createEnvironmentDetails(m *model, machineOptions []huh.Option[string], confirmationTitle string, confirmationBtn string) {

	m.newEnvironmentForm = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Environment Name").
				Value(&newEnvironmentName).
				Validate(func(str string) error {
					/*if str == "Frank" {
					}*/
					return nil
				}),
			huh.NewInput().
				Title("Branch").
				Placeholder("main, master, dev, etc").
				Value(&newEnvironmentBranchName).
				Validate(func(str string) error {
					/*if str == "Frank" {
					}*/
					return nil
				}),
			huh.NewInput().
				Title("Port").
				Placeholder("4008, 5005, etc").
				Value(&newEnvironmentPort).
				Validate(func(str string) error {
					/*if str == "Frank" {
					}*/
					return nil
				}),
			huh.NewInput().
				Title("Domain").
				Description("Without HTTPS—for example, project.com—the DNS A record for that domain or subdomain should resolve to the IP address of the load balancer machine").
				Placeholder("project.com").
				Value(&newEnvironmentDomain).
				Validate(func(str string) error {
					/*if str == "Frank" {
					}*/
					return nil
				}),
			huh.NewMultiSelect[string]().
				Title("Choose Servers to Deploy").
				Value(&newEnvironmentMachines).
				Options(machineOptions...),
			huh.NewConfirm().
				Key("done").
				Title(confirmationTitle).
				Validate(func(v bool) error {
					if !v {
						screenType = 1
					}
					return nil
				}).
				Affirmative(confirmationBtn).
				Negative("Cancel").
				Value(&newEnvironmentIsAdd),
		),
	).WithHeight(m.screenHeight - 14)

	m.newEnvironmentForm.Init()
}
