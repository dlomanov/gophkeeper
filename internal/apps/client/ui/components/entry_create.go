package components

import (
	"context"
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/input"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/styles"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type (
	EntryCreate struct {
		title      string
		back       base.Component
		focusIndex int
		syncing    bool

		inputs    []input.Input
		entryUC   EntryCreateUC
		entryType core.EntryType
		validator func([]input.Input) error
		requester func([]input.Input) (entities.CreateEntryRequest, error)
	}
	EntryCreateUC interface {
		Create(ctx context.Context, request entities.CreateEntryRequest) (entities.CreateEntryResponse, error)
	}
	entryCreateMsg struct {
		id      uuid.UUID
		version int64
		err     error
	}
)

func NewEntryCreatePassword(
	title string,
	back base.Component,
	entryUC EntryCreateUC,
) *EntryCreate {
	typ := core.EntryTypePassword
	const (
		keyIndex = iota
		loginIndex
		passwordIndex
		descIndex
		inputCount
	)
	inputs := make([]input.Input, inputCount)
	inputs[keyIndex] = input.NewText("key", 32)
	inputs[loginIndex] = input.NewText("login", 32)
	inputs[passwordIndex] = input.NewText("password", 32)
	inputs[descIndex] = input.NewArea("description", 280)
	return &EntryCreate{
		title:      fmt.Sprintf("%s/%s", title, typ),
		back:       back,
		focusIndex: 0,

		entryType: typ,
		entryUC:   entryUC,
		inputs:    inputs,
		validator: func(inputs []input.Input) error {
			if inputs[keyIndex].Value() == "" {
				return errors.New("key should not be empty")
			}
			if inputs[loginIndex].Value() == "" {
				return errors.New("login should not be empty")
			}
			if inputs[passwordIndex].Value() == "" {
				return errors.New("password should not be empty")
			}
			return nil
		},
		requester: func(inputs []input.Input) (entities.CreateEntryRequest, error) {
			return entities.CreateEntryRequest{
				Key:  inputs[keyIndex].Value(),
				Type: typ,
				Meta: map[string]string{"description": inputs[descIndex].Value()},
				Data: entities.EntryDataPassword{
					Login:    inputs[loginIndex].Value(),
					Password: inputs[passwordIndex].Value(),
				},
			}, nil
		},
	}
}

func NewEntryCreateCard(
	title string,
	back base.Component,
	entryUC EntryCreateUC,
) *EntryCreate {
	typ := core.EntryTypeCard
	const (
		keyIndex = iota
		numberIndex
		ownerIndex
		expiresIndex
		cvvIndex
		descIndex
		inputCount
	)
	inputs := make([]input.Input, inputCount)
	inputs[keyIndex] = input.NewText("key", 32)
	inputs[numberIndex] = input.NewText("number", 32)
	inputs[ownerIndex] = input.NewText("owner", 32)
	inputs[expiresIndex] = input.NewText("expires", 32)
	inputs[cvvIndex] = input.NewText("cvv", 32)
	inputs[descIndex] = input.NewArea("description", 280)
	return &EntryCreate{
		title:      fmt.Sprintf("%s/%s", title, typ),
		back:       back,
		focusIndex: 0,

		entryType: typ,
		entryUC:   entryUC,
		inputs:    inputs,
		validator: func(inputs []input.Input) error {
			if inputs[keyIndex].Value() == "" {
				return errors.New("key should not be empty")
			}
			if inputs[numberIndex].Value() == "" {
				return errors.New("number should not be empty")
			}
			return nil
		},
		requester: func(inputs []input.Input) (entities.CreateEntryRequest, error) {
			return entities.CreateEntryRequest{
				Key:  inputs[keyIndex].Value(),
				Type: typ,
				Meta: map[string]string{"description": inputs[descIndex].Value()},
				Data: entities.EntryDataCard{
					Number:  inputs[numberIndex].Value(),
					Owner:   inputs[ownerIndex].Value(),
					Expires: inputs[expiresIndex].Value(),
					Cvc:     inputs[cvvIndex].Value(),
				},
			}, nil
		},
	}
}

func NewEntryCreateNote(
	title string,
	back base.Component,
	entryUC EntryCreateUC,
) *EntryCreate {
	typ := core.EntryTypeNote
	const (
		keyIndex = iota
		noteIndex
		inputCount
	)
	inputs := make([]input.Input, inputCount)
	inputs[keyIndex] = input.NewText("key", 32)
	inputs[noteIndex] = input.NewArea("note", 280)
	return &EntryCreate{
		title:      fmt.Sprintf("%s/%s", title, typ),
		back:       back,
		focusIndex: 0,

		entryType: typ,
		entryUC:   entryUC,
		inputs:    inputs,
		validator: func(inputs []input.Input) error {
			if inputs[keyIndex].Value() == "" {
				return errors.New("key should not be empty")
			}
			if inputs[noteIndex].Value() == "" {
				return errors.New("note should not be empty")
			}
			return nil
		},
		requester: func(inputs []input.Input) (entities.CreateEntryRequest, error) {
			return entities.CreateEntryRequest{
				Key:  inputs[keyIndex].Value(),
				Type: typ,
				Meta: nil,
				Data: entities.EntryDataNote(inputs[noteIndex].Value()),
			}, nil
		},
	}
}

func NewEntryCreateBinary(
	title string,
	back base.Component,
	entryUC EntryCreateUC,
) *EntryCreate {
	typ := core.EntryTypeBinary
	const (
		keyIndex = iota
		pathIndex
		descIndex
		inputCount
	)
	inputs := make([]input.Input, inputCount)
	inputs[keyIndex] = input.NewText("key", 32)
	inputs[pathIndex] = input.NewText("filepath", 64)
	inputs[descIndex] = input.NewArea("description", 280)
	return &EntryCreate{
		title:      fmt.Sprintf("%s/%s", title, typ),
		back:       back,
		focusIndex: 0,

		entryType: typ,
		entryUC:   entryUC,
		inputs:    inputs,
		validator: func(inputs []input.Input) error {
			if inputs[keyIndex].Value() == "" {
				return errors.New("key should not be empty")
			}
			if inputs[pathIndex].Value() == "" {
				return errors.New("path should not be empty")
			}
			return nil
		},
		requester: func(inputs []input.Input) (request entities.CreateEntryRequest, err error) {
			path := inputs[pathIndex].Value()
			fileInfo, err := os.Stat(path)
			if err != nil {
				return request, fmt.Errorf("failed to get file info: %w", err)
			}
			if fileInfo.IsDir() {
				return request, fmt.Errorf("path must be a file")
			}
			if fileInfo.Size() > entities.EntryMaxDataSize {
				return request, fmt.Errorf("file size must be less than %d bytes", entities.EntryMaxDataSize)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return request, fmt.Errorf("failed to read file: %w", err)
			}
			name := filepath.Base(path)
			return entities.CreateEntryRequest{
				Key:  inputs[keyIndex].Value(),
				Type: typ,
				Meta: map[string]string{
					"filename": name,
				},
				Data: entities.EntryDataBinary(data),
			}, nil
		},
	}
}

func (c *EntryCreate) Title() string {
	return c.title
}

func (c *EntryCreate) Init() (result base.InitResult) {
	c.syncing = false
	return result.AppendCmd(c.reset())
}

func (c *EntryCreate) Update(msg tea.Msg) (result base.UpdateResult) {
	switch msg := msg.(type) {
	case entryCreateMsg:
		return c.updateEntryMsg(msg, result)
	case tea.KeyMsg:
		if result = c.updateKeyMsg(msg, result); result.Cmd != nil {
			return result
		}
	}
	cmds := make([]tea.Cmd, len(c.inputs))
	for i, v := range c.inputs {
		cmds[i] = v.Update(msg)
	}
	return result.AppendCmd(cmds...)
}

func (c *EntryCreate) View() string {
	sb := strings.Builder{}
	for i := range c.inputs {
		sb.WriteString(c.inputs[i].View())
		sb.WriteByte('\n')
	}
	sb.WriteByte('\n')
	if c.focusIndex == len(c.inputs) {
		sb.WriteString(styles.FocusedButton)
	} else {
		sb.WriteString(styles.BlurredButton)
	}
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(styles.SubtleStyle.Render("q: quit"))
	sb.WriteString(styles.DotStyle)
	sb.WriteString(styles.SubtleStyle.Render("esc: back"))
	return sb.String()
}

func (c *EntryCreate) updateEntryMsg(
	msg entryCreateMsg,
	result base.UpdateResult,
) base.UpdateResult {
	c.syncing = false
	if msg.err != nil {
		result.Status = msg.err.Error()
		return result
	}
	result.Status = "entry created 🔥"
	result.Next = c.back
	return result
}

func (c *EntryCreate) updateKeyMsg(
	msg tea.KeyMsg,
	result base.UpdateResult,
) base.UpdateResult {
	k := msg.String()
	switch k {
	case "q", "ctrl+c":
		if c.syncing {
			result.Status = result.Status + "."
			return result
		}
		result.Quitting = true
		return result.AppendCmd(tea.Quit)
	case "esc":
		if c.syncing {
			result.Status = result.Status + "."
			return result
		}
		result.Prev = c.back
		return result
	case "tab", "shift+tab", "enter", "up", "down":
		if k == "enter" && c.focusIndex == len(c.inputs) {
			if err := c.validator(c.inputs); err != nil {
				result.Status = err.Error() + " ⛔"
				return result
			}
			if c.syncing {
				result.Status = "🤔"
				return result
			}
			c.syncing = true
			result.Status = "creating..."
			cmd := c.entryCreateCmd()
			return result.AppendCmd(cmd)
		}

		if k == "up" || k == "shift+tab" {
			c.focusIndex--
		} else {
			c.focusIndex++
		}
		if c.focusIndex > len(c.inputs) {
			c.focusIndex = 0
		} else if c.focusIndex < 0 {
			c.focusIndex = len(c.inputs)
		}
		for i, v := range c.inputs {
			if i == c.focusIndex {
				result = result.AppendCmd(v.Focus())
			} else {
				v.Blur()
			}
		}
	}
	return result
}

func (c *EntryCreate) reset() (cmd tea.Cmd) {
	c.focusIndex = 0
	for _, v := range c.inputs {
		v.Reset()
	}
	return c.inputs[0].Focus()
}

func (c *EntryCreate) entryCreateCmd() tea.Cmd {
	return func() tea.Msg {
		request, err := c.requester(c.inputs)
		if err != nil {
			return entryCreateMsg{err: err}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		resp, err := c.entryUC.Create(ctx, request)
		if err != nil {
			return entryCreateMsg{err: err}
		}
		return entryCreateMsg{
			id:  resp.ID,
			err: nil,
		}
	}
}
