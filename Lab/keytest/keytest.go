package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}
func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		fmt.Printf("\nType=%v String=%q\n", k.Type, k.String())
		if k.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}
func (m model) View() string { return "PgDn 눌러보세요 (q=종료)\n" }

func main() {
	tea.NewProgram(model{}).Run()
}
