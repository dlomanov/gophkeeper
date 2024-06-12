package components

import (
	"context"
	"errors"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	"github.com/google/uuid"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"slices"
	"strings"
	"sync/atomic"
	"time"
)

var _ Component = (*EntryTable)(nil)

type (
	EntryTable struct {
		title   string
		back    Component
		table   table.Model
		logger  *zap.Logger
		entryUC *usecases.EntryUC
		entries []entities.Entry
		syncing atomic.Int64
	}
	syncMsg struct {
		entries []entities.Entry
		err     error
	}
	deleteMsg struct {
		err error
	}
)

func NewEntryTable(
	title string,
	back Component,
	entryUC *usecases.EntryUC,
	logger *zap.Logger,
) *EntryTable {
	c := &EntryTable{
		title:   title,
		back:    back,
		entryUC: entryUC,
		logger:  logger,
	}
	c.table = c.newTable()
	return c
}

func (c *EntryTable) Title() string {
	return c.title
}

func (c *EntryTable) Init() tea.Cmd {
	c.reset()
	c.syncing.Store(1)
	return c.syncCmd()
}

func (c *EntryTable) Update(msg tea.Msg) (result UpdateResult, cmd tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		switch k {
		case "q", "ctrl+c":
			if c.syncing.Load() != 0 {
				result.Status = result.Status + "."
				return result, nil
			}
			result.Quitting = true
			return result, tea.Quit
		case "esc":
			if c.syncing.Load() != 0 {
				result.Status = result.Status + "."
				return result, nil
			}
			result.Prev = c.back
			return result, nil
		case "j", "k", "up", "down":
			if !c.table.Focused() {
				c.table.Focus()
			}
		case "enter":
			if c.syncing.Load() != 0 {
				result.Status = result.Status + "."
				return result, nil
			}
			row := c.table.SelectedRow()
			if len(row) == 0 {
				return result, nil
			}
			key := c.table.SelectedRow()[0]
			idx := slices.IndexFunc(c.entries, func(entry entities.Entry) bool { return entry.Key == key })

			// create
			if idx == -1 {
				result.Next = NewEntryCreateCard(
					c.title+"/create",
					c,
					c.entryUC,
				)
				return result, nil
			}

			// update
			result.Next = NewEntryUpdateCard(
				c.title+"/update",
				c,
				c.entryUC,
				c.entries[idx],
			)
			return result, nil
		case "s", "S":
			if c.syncing.CompareAndSwap(0, 1) {
				result.Status = result.Status + "."
				return result, c.syncCmd()
			}
			result.Status = "syncing 2"
			return result, nil
		case "delete", "d":
			if !c.syncing.CompareAndSwap(0, 1) {
				result.Status = result.Status + "."
				return result, nil
			}
			row := c.table.SelectedRow()
			if len(row) == 0 {
				return result, nil
			}
			key := c.table.SelectedRow()[0]
			idx := slices.IndexFunc(c.entries, func(entry entities.Entry) bool { return entry.Key == key })
			if idx == -1 {
				return result, nil
			}
			result.Status = "deleting..."
			id := c.entries[idx].ID
			return result, c.deleteCmd(id)
		}
	}

	if msg, ok := msg.(syncMsg); ok {
		c.syncing.Store(0)
		result.Status = "synced"
		if msg.err != nil {
			switch {
			case errors.Is(msg.err, entities.ErrServerUnavailable):
				result.Status = "server unavailable"
			default:
				result.Status = msg.err.Error() // TODO: internal error
			}
		}
		c.entries = msg.entries
		rows := make([]table.Row, len(c.entries)+1)
		rows[0] = table.Row{"", "press enter to create new entry"}
		for i, entry := range c.entries {
			rows[i+1] = table.Row{
				entry.Key,
				entry.Meta["description"],
				entry.UpdatedAt.Format(time.DateTime),
			}
		}
		c.table.SetRows(rows)
		return result, nil
	}

	if msg, ok := msg.(deleteMsg); ok {
		c.syncing.Store(0)
		if msg.err != nil {
			result.Status = msg.err.Error()
			return result, nil
		}
		result.Status = "entry deleted"
		return result, c.syncCmd()
	}

	c.table, cmd = c.table.Update(msg)
	return result, cmd
}

func (c *EntryTable) View() string {
	sb := strings.Builder{}
	sb.WriteString(c.table.View())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(subtleStyle.Render("esc: back"))
	sb.WriteString(dotStyle)
	sb.WriteString(subtleStyle.Render("q: quit"))
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	return sb.String()
}

func (c *EntryTable) SetPrev(back Component) {
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

func (c *EntryTable) reset() {
	c.table.Blur()
	c.table.SetRows(nil)
}

func (c *EntryTable) deleteCmd(id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err := c.entryUC.Delete(ctx, usecases.DeleteEntryRequest{ID: id})
		return deleteMsg{err: err}
	}
}

func (c *EntryTable) syncCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		var merr error
		if err := c.entryUC.Sync(ctx); err != nil {
			c.logger.Error("failed to sync entries", zap.Error(err))
			merr = multierr.Append(merr, err)
		}
		resp, err := c.entryUC.GetAll(ctx)
		if err != nil {
			merr = multierr.Append(merr, err)
			return syncMsg{err: merr}
		}
		return syncMsg{entries: resp.Entries, err: merr}
	}
}
