package components

import (
	"context"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/google/uuid"
	"strings"
	"sync/atomic"
	"time"
)

var _ Component = (*EntryCreateCard)(nil)

type (
	EntryCreateCard struct {
		title      string
		back       Component
		entryUC    *usecases.EntryUC
		focusIndex int
		keyInput   textinput.Model
		typeInput  textinput.Model
		metaInput  textarea.Model
		dataInput  textarea.Model
		inputCount int
		syncing    atomic.Int64
	}
	createEntryMsg struct {
		id      uuid.UUID
		version int64
		err     error
	}
)

func NewEntryCreateCard(
	title string,
	back Component,
	entryUC *usecases.EntryUC,
) *EntryCreateCard {
	c := &EntryCreateCard{
		title:      title,
		back:       back,
		entryUC:    entryUC,
		focusIndex: 0,
	}
	c.keyInput = c.keyTextInput()
	c.typeInput = c.typeTextInput()
	c.metaInput = c.metaTextArea()
	c.dataInput = c.dataTextArea()
	c.inputCount = 4
	return c
}

func (c *EntryCreateCard) Title() string {
	return c.title
}

func (c *EntryCreateCard) Init() tea.Cmd {
	c.syncing.Store(0)
	return c.reset()
}

func (c *EntryCreateCard) Update(msg tea.Msg) (result UpdateResult, cmd tea.Cmd) {
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
		case "tab", "shift+tab", "enter", "up", "down":
			if k == "enter" && c.focusIndex == c.inputCount {
				if c.inputsValid() {
					result.Status = "invalid inputs"
					return result, c.createEntryCmd()
				}
				if !c.syncing.CompareAndSwap(0, 1) {
					result.Status = "entry creation in progress"
					return result, nil
				}
				result.Status = "entry creation"
				return result, c.createEntryCmd()
			}

			if k == "up" || k == "shift+tab" {
				c.focusIndex--
			} else {
				c.focusIndex++
			}
			if c.focusIndex > c.inputCount {
				c.focusIndex = 0
			} else if c.focusIndex < 0 {
				c.focusIndex = c.inputCount
			}

			c.keyInput.Blur()
			c.keyInput.PromptStyle = noStyle
			c.keyInput.TextStyle = noStyle
			c.typeInput.Blur()
			c.typeInput.PromptStyle = noStyle
			c.typeInput.TextStyle = noStyle
			c.metaInput.Blur()
			c.dataInput.Blur()
			if c.focusIndex < c.inputCount {
				switch c.focusIndex {
				case 0:
					c.keyInput.Focus()
					c.keyInput.PromptStyle = focusedStyle
					c.keyInput.TextStyle = focusedStyle
					return result, c.keyInput.Focus()
				case 1:
					c.typeInput.PromptStyle = focusedStyle
					c.typeInput.TextStyle = focusedStyle
					return result, c.typeInput.Focus()
				case 2:
					return result, c.metaInput.Focus()
				case 3:
					return result, c.dataInput.Focus()
				}
			}
		}
	}

	if msg, ok := msg.(createEntryMsg); ok {
		c.syncing.Store(0)
		if msg.err != nil {
			result.Status = msg.err.Error()
			return result, nil
		}
		result.Status = "entry created"
		result.Next = c.back
		return result, nil
	}

	cmds := make([]tea.Cmd, 4)
	c.keyInput, cmd = c.keyInput.Update(msg)
	cmds = append(cmds, cmd)
	c.typeInput, cmd = c.typeInput.Update(msg)
	cmds = append(cmds, cmd)
	c.metaInput, cmd = c.metaInput.Update(msg)
	cmds = append(cmds, cmd)
	c.dataInput, cmd = c.dataInput.Update(msg)
	cmds = append(cmds, cmd)
	cmd = tea.Batch(cmds...)

	return result, cmd
}

func (c *EntryCreateCard) View() string {
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

func (c *EntryCreateCard) reset() (cmd tea.Cmd) {
	c.focusIndex = 0

	c.keyInput.SetValue("")
	c.keyInput.Focus()
	c.keyInput.PromptStyle = focusedStyle
	c.keyInput.TextStyle = focusedStyle

	c.typeInput.SetValue("")
	c.typeInput.Blur()
	c.typeInput.PromptStyle = noStyle
	c.typeInput.TextStyle = noStyle

	c.metaInput.SetValue("")
	c.metaInput.Blur()

	c.dataInput.SetValue("")
	c.dataInput.Blur()

	return cmd
}

func (c *EntryCreateCard) keyTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "key"
	ti.CharLimit = 32
	return ti
}

func (c *EntryCreateCard) typeTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "type: password, note, card, binary"
	ti.CharLimit = 16
	return ti
}

func (c *EntryCreateCard) metaTextArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "metadata"
	ta.ShowLineNumbers = false
	return ta
}

func (c *EntryCreateCard) dataTextArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "data"
	ta.ShowLineNumbers = false
	return ta
}

func (c *EntryCreateCard) inputsValid() bool {
	valid := c.keyInput.Value() != "" &&
		c.typeInput.Value() != "" &&
		c.metaInput.Value() != "" &&
		c.dataInput.Value() != ""
	if !valid {
		return false
	}

	t := strings.ToLower(c.typeInput.Value())
	entryType := core.EntryType(t)
	return entryType.Valid()
}

func (c *EntryCreateCard) createEntryCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		resp, err := c.entryUC.Create(ctx, usecases.CreateEntryRequest{
			Key:  c.keyInput.Value(),
			Type: core.EntryType(c.typeInput.Value()),
			Meta: map[string]string{"description": c.metaInput.Value()},
			Data: []byte(c.dataInput.Value()),
		})
		if err != nil {
			return createEntryMsg{err: err}
		}
		return createEntryMsg{
			id:  resp.ID,
			err: nil,
		}
	}
}
