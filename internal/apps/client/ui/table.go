package ui

import (
	"context"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/google/uuid"
	"slices"
	"strings"
	"sync/atomic"
	"time"
)

var _ Component = (*Table)(nil)

type (
	Table struct {
		title      string
		back       Component
		table      table.Model
		entryUC    *usecases.EntryUC
		entries    []entities.Entry
		refreshing atomic.Int64
		deleting   atomic.Int64
	}
	refreshMsg struct {
		entries []entities.Entry
		err     error
	}
	deleteMsg struct {
		err     error
		id      uuid.UUID
		version int64
	}
)

func NewTable(
	title string,
	back Component,
	entryUC *usecases.EntryUC,
) *Table {
	c := &Table{
		title:   title,
		back:    back,
		entryUC: entryUC,
	}
	c.table = c.newTable()
	return c
}

func (c *Table) Title() string {
	return c.title
}

func (c *Table) Init() tea.Cmd {
	c.reset()
	return c.refreshCmd()
}

func (c *Table) Update(msg tea.Msg) (result UpdateResult, cmd tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		switch k {
		case "q", "ctrl+c":
			result.Quitting = true
			return result, tea.Quit
		case "esc":
			if c.table.Focused() {
				c.table.Blur()
			} else {
				result.Prev = c.back
				return result, nil
			}
		case "j", "k", "up", "down":
			if !c.table.Focused() {
				c.table.Focus()
			}
		case "enter":
			row := c.table.SelectedRow()
			if len(row) == 0 {
				return result, nil
			}
			key := c.table.SelectedRow()[0]
			idx := slices.IndexFunc(c.entries, func(entry entities.Entry) bool { return entry.Key == key })

			// create
			if idx == -1 {
				result.Next = NewCreateCard(
					c.title+"/create",
					c,
					c.entryUC,
				)
				return result, nil
			}

			// update
			result.Next = NewUpdateCard(
				c.title+"/update",
				c,
				c.entryUC,
				c.entries[idx],
			)
			return result, nil
		case "u", "U":
			if c.refreshing.CompareAndSwap(0, 1) {
				result.Status = "refreshing 1"
				return result, c.refreshCmd()
			}
			result.Status = "refreshing 2"
			return result, nil
		case "delete", "d":
			row := c.table.SelectedRow()
			if len(row) == 0 {
				return result, nil
			}
			key := c.table.SelectedRow()[0]
			idx := slices.IndexFunc(c.entries, func(entry entities.Entry) bool { return entry.Key == key })
			if idx == -1 {
				return result, nil
			}
			if !c.deleting.CompareAndSwap(0, 1) {
				result.Status = "deleting 1"
				return result, nil
			}
			result.Status = "deleting..."
			id := c.entries[idx].ID
			return result, c.deleteCmd(id)
		}
	}

	if msg, ok := msg.(refreshMsg); ok {
		c.refreshing.Store(0)
		if msg.err != nil {
			c.entries = nil
			result.Status = msg.err.Error()
			return result, nil
		}
		c.entries = msg.entries
		rows := make([]table.Row, len(c.entries)+1)
		rows[0] = table.Row{"", "press enter to create new entry"}
		for i, entry := range c.entries {
			rows[i+1] = table.Row{
				entry.Key,
				entry.Meta["description"],
			}
		}
		c.table.SetRows(rows)
		result.Status = "table updated"
		return result, nil
	}

	if msg, ok := msg.(deleteMsg); ok {
		c.deleting.Store(0)
		if msg.err != nil {
			result.Status = msg.err.Error()
			return result, nil
		}
		result.Status = "entry deleted"
		return result, c.refreshCmd()
	}

	c.table, cmd = c.table.Update(msg)
	return result, cmd
}

func (c *Table) View() string {
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

func (c *Table) SetPrev(back Component) {
	c.back = back
}

func (c *Table) newTable() table.Model {
	columns := []table.Column{
		{Title: "Key", Width: 15},
		{Title: "Description", Width: 30},
	}
	t := table.New(
		table.WithColumns(columns),
		//table.WithRows(rows),
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

func (c *Table) reset() {
	c.table.Blur()
	c.table.SetRows(nil)
}

func (c *Table) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		resp, err := c.entryUC.GetAll(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}
		return refreshMsg{entries: resp.Entries}
	}
}

func (c *Table) deleteCmd(id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		response, err := c.entryUC.Delete(ctx, usecases.DeleteEntryRequest{ID: id})
		if err != nil {
			return deleteMsg{err: err}
		}
		return deleteMsg{
			id:      response.ID,
			version: response.Version,
			err:     nil,
		}
	}
}
