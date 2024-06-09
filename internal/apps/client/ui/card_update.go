package ui

import (
	"context"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/google/uuid"
	"strings"
	"sync/atomic"
	"time"
)

var _ Component = (*UpdateCard)(nil)

type (
	UpdateCard struct {
		title      string
		back       Component
		entryUC    *usecases.EntryUC
		entry      entities.Entry
		focusIndex int
		keyInput   textinput.Model
		typeInput  textinput.Model
		metaInput  textarea.Model
		dataInput  textarea.Model
		inputCount int
		creating   atomic.Int64
	}
	updateEntryMsg struct {
		id      uuid.UUID
		version int64
		err     error
	}
)

func NewUpdateCard(
	title string,
	back Component,
	entryUC *usecases.EntryUC,
	entry entities.Entry,
) *UpdateCard {
	c := &UpdateCard{
		title:      title,
		back:       back,
		entryUC:    entryUC,
		entry:      entry,
		focusIndex: 0,
	}
	c.keyInput = c.keyTextInput()
	c.typeInput = c.typeTextInput()
	c.metaInput = c.metaTextArea()
	c.dataInput = c.dataTextArea()
	c.inputCount = 4
	return c
}

func (c *UpdateCard) Title() string {
	return c.title
}

func (c *UpdateCard) Init() tea.Cmd {
	return c.reset()
}

func (c *UpdateCard) Update(msg tea.Msg) (result UpdateResult, cmd tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		switch k {
		case "q", "ctrl+c":
			result.Quitting = true
			return result, tea.Quit
		case "esc":
			result.Prev = c.back
			return result, nil
		case "tab", "shift+tab", "enter", "up", "down":
			if k == "enter" && c.focusIndex == c.inputCount {
				if c.inputsValid() {
					result.Status = "invalid inputs"
					return result, c.updateEntryCmd()
				}
				if !c.creating.CompareAndSwap(0, 1) {
					result.Status = "entry creation in progress"
					return result, nil
				}
				if !c.inputsUpdated() {
					result.Status = "no changes"
					result.Prev = c.back
					return result, nil
				}
				result.Status = "entry creation"
				return result, c.updateEntryCmd()
			}

			if k == "up" || k == "shift+tab" {
				c.focusIndex--
			} else {
				c.focusIndex++
			}
			if c.focusIndex > c.inputCount {
				c.focusIndex = 2
			} else if c.focusIndex < 2 {
				c.focusIndex = c.inputCount
			}

			c.metaInput.Blur()
			c.dataInput.Blur()
			if c.focusIndex < c.inputCount {
				switch c.focusIndex {
				case 2:
					return result, c.metaInput.Focus()
				case 3:
					return result, c.dataInput.Focus()
				}
			}
		}
	}

	if msg, ok := msg.(updateEntryMsg); ok {
		c.creating.Store(0)
		if msg.err != nil {
			result.Status = msg.err.Error()
			return result, nil
		}
		result.Status = "entry created"
		result.Next = c.back
		return result, nil
	}

	cmds := make([]tea.Cmd, 2)
	c.metaInput, cmd = c.metaInput.Update(msg)
	cmds = append(cmds, cmd)
	c.dataInput, cmd = c.dataInput.Update(msg)
	cmds = append(cmds, cmd)
	cmd = tea.Batch(cmds...)

	return result, cmd
}

func (c *UpdateCard) View() string {
	sb := strings.Builder{}
	sb.WriteString(c.keyInput.View())
	sb.WriteByte('\n')
	sb.WriteString(c.typeInput.View())
	sb.WriteByte('\n')
	sb.WriteString(c.metaInput.View())
	sb.WriteByte('\n')
	sb.WriteString(c.dataInput.View())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	if c.focusIndex == c.inputCount {
		sb.WriteString(focusedButton)
	} else {
		sb.WriteString(blurredButton)
	}
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(subtleStyle.Render("q: quit"))
	sb.WriteString(dotStyle)
	sb.WriteString(subtleStyle.Render("esc: back"))
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	return sb.String()
}

func (c *UpdateCard) reset() (cmd tea.Cmd) {
	c.focusIndex = 0

	c.keyInput.SetValue(c.entry.Key)
	c.keyInput.Blur()
	c.keyInput.PromptStyle = noStyle
	c.keyInput.TextStyle = noStyle

	c.typeInput.SetValue(string(c.entry.Type))
	c.typeInput.Blur()
	c.typeInput.PromptStyle = noStyle
	c.typeInput.TextStyle = noStyle

	c.metaInput.SetValue(c.entry.Meta["description"])
	c.metaInput.Focus()

	c.dataInput.SetValue(string(c.entry.Data))
	c.dataInput.Blur()

	return cmd
}

func (c *UpdateCard) keyTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "key"
	ti.CharLimit = 32
	return ti
}

func (c *UpdateCard) typeTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "type"
	ti.Placeholder = "type: password, note, card, binary"
	ti.CharLimit = 16
	return ti
}

func (c *UpdateCard) metaTextArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "metadata"
	ta.ShowLineNumbers = false
	return ta
}

func (c *UpdateCard) dataTextArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "data"
	ta.ShowLineNumbers = false
	return ta
}

func (c *UpdateCard) inputsValid() bool {
	valid := c.keyInput.Value() != "" &&
		c.typeInput.Value() != "" &&
		c.metaInput.Value() != "" &&
		c.dataInput.Value() != ""
	if !valid {
		return false
	}

	t := strings.ToLower(c.typeInput.Value())
	entryType := entities.EntryType(t)
	return entryType.Valid()
}

func (c *UpdateCard) inputsUpdated() bool {
	return c.keyInput.Value() != c.entry.Key ||
		c.typeInput.Value() != string(c.entry.Type) ||
		c.metaInput.Value() != c.entry.Meta["description"] ||
		c.dataInput.Value() != string(c.entry.Data)
}

func (c *UpdateCard) updateEntryCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		resp, err := c.entryUC.Update(ctx, usecases.UpdateEntryRequest{
			ID:      c.entry.ID,
			Version: c.entry.Version,
			Meta:    map[string]string{"description": c.metaInput.Value()},
			Data:    []byte(c.dataInput.Value()),
		})
		if err != nil {
			return updateEntryMsg{err: err}
		}
		return updateEntryMsg{
			id:      resp.ID,
			version: resp.Version,
			err:     nil,
		}
	}
}
