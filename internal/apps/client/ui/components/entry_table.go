package components

import (
	"context"
	"errors"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/styles"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"slices"
	"strings"
	"time"
)

var _ base.Component = (*EntryTable)(nil)

type (
	EntryTable struct {
		title   string
		back    base.Component
		table   table.Model
		logger  *zap.Logger
		entryUC EntryUC
		entries []entities.GetEntryResponse
		syncing bool
	}
	EntryUC interface {
		EntryCreateUC
		EntryUpdateUC
		Sync(ctx context.Context) error
		GetAll(ctx context.Context) (entities.GetEntriesResponse, error)
		Delete(ctx context.Context, request entities.DeleteEntryRequest) error
	}
	syncMsg struct {
		entries []entities.GetEntryResponse
		err     error
	}
	deleteMsg struct {
		err error
	}
)

func NewEntryTable(
	title string,
	entryUC EntryUC,
	logger *zap.Logger,
) *EntryTable {
	c := &EntryTable{
		title:   title,
		back:    nil,
		entryUC: entryUC,
		logger:  logger,
	}
	c.table = c.newTable()
	return c
}

func (c *EntryTable) Title() string {
	return c.title
}

func (c *EntryTable) Init() (result base.InitResult) {
	c.reset()
	c.syncing = true
	return result.AppendCmd(c.syncCmd())
}

func (c *EntryTable) Update(msg tea.Msg) (result base.UpdateResult) {
	switch msg := msg.(type) {
	case syncMsg:
		return c.updateSyncMsg(msg, result)
	case deleteMsg:
		return c.updateDeleteMsg(msg, result)
	case tea.KeyMsg:
		if result = c.updateKeyMsg(msg, result); result.Cmd != nil {
			return result
		}
	}
	var cmd tea.Cmd
	c.table, cmd = c.table.Update(msg)
	return result.AppendCmd(cmd)
}

func (c *EntryTable) View() string {
	sb := strings.Builder{}
	sb.WriteString(c.table.View())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(styles.SubtleStyle.Render("s: sync"))
	sb.WriteString(styles.DotStyle)
	sb.WriteString(styles.SubtleStyle.Render("enter: select"))
	sb.WriteString(styles.DotStyle)
	sb.WriteString(styles.SubtleStyle.Render("d: delete"))
	sb.WriteByte('\n')
	sb.WriteString(styles.SubtleStyle.Render("esc: back"))
	sb.WriteString(styles.DotStyle)
	sb.WriteString(styles.SubtleStyle.Render("q: quit"))
	return sb.String()
}

func (c *EntryTable) SetPrev(back base.Component) {
	c.back = back
}

func (c *EntryTable) newTable() table.Model {
	columns := []table.Column{
		{Title: "Key", Width: 15},
		{Title: "Description", Width: 30},
		{Title: "Updated", Width: 25},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(7),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return t
}

func (c *EntryTable) updateSyncMsg(
	msg syncMsg,
	result base.UpdateResult,
) base.UpdateResult {
	c.syncing = false
	result.Status = "synced 🦾"
	if msg.err != nil {
		switch {
		case errors.Is(msg.err, entities.ErrServerUnavailable):
			result.Status = "can't sync 🤨: server unavailable"
		case errors.Is(msg.err, entities.ErrUserTokenInvalid):
			result.Status = "for syncing try sign-in/sign-up first 🤔"
		case errors.Is(msg.err, entities.ErrUserTokenNotFound):
			result.Status = "for syncing try sign-in/sign-up first 🤔"
		default:
			result.Status = "can't sync 🤨: internal server error 💀"
		}
	}
	c.entries = msg.entries
	rows := make([]table.Row, len(c.entries)+1)
	rows[0] = table.Row{"press enter", "to create new entry"}
	for i, entry := range c.entries {
		rows[i+1] = table.Row{
			entry.Key,
			entry.Meta["description"],
			entry.UpdatedAt.Format(time.DateTime),
		}
	}
	c.table.SetRows(rows)
	return result
}

func (c *EntryTable) updateDeleteMsg(
	msg deleteMsg,
	result base.UpdateResult,
) base.UpdateResult {
	c.syncing = false
	if msg.err != nil {
		result.Status = msg.err.Error()
		return result
	}
	result.Status = "entry deleted"
	return result.AppendCmd(c.syncCmd())
}

func (c *EntryTable) updateKeyMsg(
	msg tea.KeyMsg,
	result base.UpdateResult,
) base.UpdateResult {
	k := msg.String()
	switch k {
	case "q", "ctrl+c":
		if c.syncing {
			result.Status = "🤔"
			return result
		}
		result.Quitting = true
		return result.AppendCmd(tea.Quit)
	case "esc":
		if c.syncing {
			result.Status = "🤔"
			return result
		}
		result.Prev = c.back
		return result
	case "j", "k", "up", "down":
		if !c.table.Focused() {
			c.table.Focus()
		}
	case "enter":
		if c.syncing {
			result.Status = "🤔"
			return result
		}
		row := c.table.SelectedRow()
		if len(row) == 0 {
			return result
		}
		key := c.table.SelectedRow()[0]
		idx := slices.IndexFunc(c.entries, func(entry entities.GetEntryResponse) bool { return entry.Key == key })

		if idx == -1 {
			return c.entryCreateSelector(result)
		}
		return c.entryUpdate(c.entries[idx], result)
	case "s":
		result.Status = "🤔"
		if c.syncing {
			return result
		}
		c.syncing = true
		return result.AppendCmd(c.syncCmd())
	case "delete", "d":
		if c.syncing {
			result.Status = "🤔"
			return result
		}
		row := c.table.SelectedRow()
		if len(row) == 0 {
			return result
		}
		key := c.table.SelectedRow()[0]
		idx := slices.IndexFunc(c.entries, func(entry entities.GetEntryResponse) bool { return entry.Key == key })
		if idx == -1 {
			return result
		}
		result.Status = "🤔"
		id := c.entries[idx].ID
		return result.AppendCmd(c.deleteCmd(id))
	}
	return result
}

func (c *EntryTable) syncCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		var merr error
		if err := c.entryUC.Sync(ctx); err != nil {
			c.logger.Error("failed to sync entries", zap.Error(err))
			merr = errors.Join(merr, err)
		}
		resp, err := c.entryUC.GetAll(ctx)
		if err != nil {
			merr = errors.Join(merr, err)
			return syncMsg{err: merr}
		}
		return syncMsg{entries: resp.Entries, err: merr}
	}
}

func (c *EntryTable) deleteCmd(id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err := c.entryUC.Delete(ctx, entities.DeleteEntryRequest{ID: id})
		return deleteMsg{err: err}
	}
}

func (c *EntryTable) reset() {
	c.table.Blur()
	c.table.SetRows(nil)
}

func (c *EntryTable) entryCreateSelector(
	result base.UpdateResult,
) base.UpdateResult {
	result.Next = NewEntryCreateSelector(
		c.title+"/create",
		c,
		c.entryUC,
	)
	return result
}
func (c *EntryTable) entryUpdate(
	entry entities.GetEntryResponse,
	result base.UpdateResult,
) base.UpdateResult {
	title := c.title + "/update"
	switch entry.Type {
	case core.EntryTypePassword:
		result.Next = NewEntryUpdatePassword(title, c, c.entryUC, entry)
	case core.EntryTypeNote:
		result.Next = NewEntryUpdateNote(title, c, c.entryUC, entry)
	case core.EntryTypeCard:
		result.Next = NewEntryUpdateCard(title, c, c.entryUC, entry)
	case core.EntryTypeBinary:
		result.Next = NewEntryUpdateBinary(title, c, c.entryUC, entry)
	}
	return result
}
