package input

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

var _ Input = (*Area)(nil)

type Area struct {
	model textarea.Model
}

func NewArea(
	placeholder string,
	charLimit int,
) *Area {
	model := textarea.New()
	model.Placeholder = placeholder
	model.CharLimit = charLimit
	model.ShowLineNumbers = false
	a := &Area{
		model: model,
	}
	a.Reset()
	return a
}

func (a *Area) Value() string {
	return a.model.Value()
}

func (a *Area) SetValue(v string) {
	a.model.SetValue(v)
}

func (a *Area) Focus() tea.Cmd {
	return a.model.Focus()
}

func (a *Area) Blur() {
	a.model.Blur()
}

func (a *Area) Reset() {
	a.model.Reset()
}

func (a *Area) View() string {
	return a.model.View()
}

func (a *Area) Update(msg tea.Msg) (cmd tea.Cmd) {
	a.model, cmd = a.model.Update(msg)
	return cmd
}
