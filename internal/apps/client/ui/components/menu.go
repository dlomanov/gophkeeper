package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/navlist"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/styles"
	"strings"
)

var _ base.Component = (*Menu)(nil)

type (
	Menu struct {
		title string
		list  navlist.List
	}
)

func NewMenu(title string, navs []navlist.Item) *Menu {
	return &Menu{
		title: title,
		list:  navlist.New(navs),
	}
}

func (c *Menu) Title() string {
	return c.title
}

func (c *Menu) Init() (result base.InitResult) {
	return result
}

func (c *Menu) Update(msg tea.Msg) (result base.UpdateResult) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.list.SetWidth(msg.Width)
	case tea.KeyMsg:
		if result = c.updateKeyMsg(msg, result); result.Cmd != nil {
			return result
		}
	}
	return c.updateList(msg, result)
}

func (c *Menu) View() string {
	sb := strings.Builder{}
	sb.WriteString(c.list.View())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(styles.SubtleStyle.Render("q: quit"))
	return sb.String()
}

func (c *Menu) updateKeyMsg(msg tea.KeyMsg, result base.UpdateResult) base.UpdateResult {
	k := msg.String()
	switch k {
	case "q", "ctrl+c":
		result.Quitting = true
		return result.AppendCmd(tea.Quit)
	case "enter":
		if v := c.list.Selected(); v != nil {
			result.Next = v.Next
		}
		return result
	}
	return result
}

func (c *Menu) updateList(msg tea.Msg, result base.UpdateResult) base.UpdateResult {
	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	return result.AppendCmd(cmd)
}
