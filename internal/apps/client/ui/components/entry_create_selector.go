package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/navlist"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/styles"
	"github.com/dlomanov/gophkeeper/internal/core"
	"go.uber.org/zap"
	"strings"
)

var _ base.Component = (*EntryCreateSelector)(nil)

type (
	EntryCreateSelector struct {
		title  string
		back   base.Component
		list   navlist.List
		logger *zap.Logger
	}
)

func NewEntryCreateSelector(
	title string,
	logger *zap.Logger,
	back base.Component,
	entryUC EntryCreateUC,
) *EntryCreateSelector {
	l := navlist.New([]navlist.Item{
		{
			Name: string(core.EntryTypePassword),
			Next: NewEntryCreatePassword(title, logger, back, entryUC),
		},
		{
			Name: string(core.EntryTypeNote),
			Next: NewEntryCreateNote(title, logger, back, entryUC),
		},
		{
			Name: string(core.EntryTypeCard),
			Next: NewEntryCreateCard(title, logger, back, entryUC),
		},
		{
			Name: string(core.EntryTypeBinary),
			Next: NewEntryCreateBinary(title, logger, back, entryUC),
		},
	})
	c := &EntryCreateSelector{
		title:  title,
		logger: logger,
		back:   back,
		list:   l,
	}
	return c
}

func (c *EntryCreateSelector) Title() string {
	return c.title
}

func (c *EntryCreateSelector) Init() (result base.InitResult) {
	return result
}

func (c *EntryCreateSelector) Update(msg tea.Msg) (result base.UpdateResult) {
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

func (c *EntryCreateSelector) View() string {
	sb := strings.Builder{}
	sb.WriteString(c.list.View())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(styles.SubtleStyle.Render("q: quit"))
	sb.WriteString(styles.DotStyle)
	sb.WriteString(styles.SubtleStyle.Render("esc: back"))
	return sb.String()
}

func (c *EntryCreateSelector) updateKeyMsg(
	msg tea.KeyMsg,
	result base.UpdateResult,
) base.UpdateResult {
	k := msg.String()
	switch k {
	case "q", "ctrl+c":
		result.Quitting = true
		return result.AppendCmd(tea.Quit)
	case "esc":
		result.Prev = c.back
		return result
	case "enter":
		if v := c.list.Selected(); v != nil {
			result.Next = v.Next
		}
		return result
	}
	return result
}

func (c *EntryCreateSelector) updateList(msg tea.Msg, result base.UpdateResult) base.UpdateResult {
	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	return result.AppendCmd(cmd)
}
