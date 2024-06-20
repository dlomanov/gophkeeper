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
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

type (
	EntryUpdate struct {
		title      string
		logger     *zap.Logger
		back       base.Component
		focusIndex int
		syncing    bool

		inputs    []input.Input
		entryUC   EntryUpdateUC
		entry     entities.GetEntryResponse
		reseter   func(entities.GetEntryResponse, []input.Input)
		validator func([]input.Input) error
		requester func([]input.Input) (entities.UpdateEntryRequest, error)
	}
	EntryUpdateUC interface {
		Update(ctx context.Context, request entities.UpdateEntryRequest) error
	}
	entryUpdateMsg struct {
		id      uuid.UUID
		version int64
		err     error
	}
	binaryDownloadMsg struct {
		err    error
		status string
	}
)

func NewEntryUpdatePassword(
	title string,
	logger *zap.Logger,
	back base.Component,
	entryUC EntryUpdateUC,
	entry entities.GetEntryResponse,
) *EntryUpdate {
	const (
		loginIndex = iota
		passwordIndex
		descIndex
		inputCount
	)
	inputs := make([]input.Input, inputCount)
	inputs[loginIndex] = input.NewText("login", 32)
	inputs[passwordIndex] = input.NewText("password", 32)
	inputs[descIndex] = input.NewArea("description", 280)
	return &EntryUpdate{
		title:      title,
		logger:     logger,
		back:       back,
		focusIndex: 0,

		entry:   entry,
		entryUC: entryUC,
		inputs:  inputs,
		reseter: func(entry entities.GetEntryResponse, inputs []input.Input) {
			passData := entry.Data.(entities.EntryDataPassword)
			inputs[loginIndex].SetValue(passData.Login)
			inputs[passwordIndex].SetValue(passData.Password)
			inputs[descIndex].SetValue(entry.Meta["description"])
		},
		validator: func(inputs []input.Input) error {
			if inputs[loginIndex].Value() == "" {
				return errors.New("login should not be empty")
			}
			if inputs[passwordIndex].Value() == "" {
				return errors.New("password should not be empty")
			}
			return nil
		},
		requester: func(inputs []input.Input) (entities.UpdateEntryRequest, error) {
			return entities.UpdateEntryRequest{
				ID:   entry.ID,
				Meta: map[string]string{"description": inputs[descIndex].Value()},
				Data: entities.EntryDataPassword{
					Login:    inputs[loginIndex].Value(),
					Password: inputs[passwordIndex].Value(),
				},
			}, nil
		},
	}
}

func NewEntryUpdateCard(
	title string,
	logger *zap.Logger,
	back base.Component,
	entryUC EntryUpdateUC,
	entry entities.GetEntryResponse,
) *EntryUpdate {
	const (
		numberIndex = iota
		ownerIndex
		expiresIndex
		cvvIndex
		descIndex
		inputCount
	)
	inputs := make([]input.Input, inputCount)
	inputs[numberIndex] = input.NewText("number", 32)
	inputs[ownerIndex] = input.NewText("owner", 32)
	inputs[expiresIndex] = input.NewText("expires", 32)
	inputs[cvvIndex] = input.NewText("cvv", 32)
	inputs[descIndex] = input.NewArea("description", 280)
	return &EntryUpdate{
		title:      title,
		logger:     logger,
		back:       back,
		focusIndex: 0,

		entryUC: entryUC,
		entry:   entry,
		inputs:  inputs,
		reseter: func(entry entities.GetEntryResponse, inputs []input.Input) {
			cardData := entry.Data.(entities.EntryDataCard)
			inputs[numberIndex].SetValue(cardData.Number)
			inputs[ownerIndex].SetValue(cardData.Owner)
			inputs[expiresIndex].SetValue(cardData.Expires)
			inputs[cvvIndex].SetValue(cardData.Cvc)
			inputs[descIndex].SetValue(entry.Meta["description"])
		},
		validator: func(inputs []input.Input) error {
			if inputs[numberIndex].Value() == "" {
				return errors.New("number should not be empty")
			}
			return nil
		},
		requester: func(inputs []input.Input) (entities.UpdateEntryRequest, error) {
			return entities.UpdateEntryRequest{
				ID:   entry.ID,
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

func NewEntryUpdateNote(
	title string,
	logger *zap.Logger,
	back base.Component,
	entryUC EntryUpdateUC,
	entry entities.GetEntryResponse,
) *EntryUpdate {
	const (
		noteIndex = iota
		inputCount
	)
	inputs := make([]input.Input, inputCount)
	inputs[noteIndex] = input.NewArea("note", 280)
	return &EntryUpdate{
		title:      title,
		logger:     logger,
		back:       back,
		focusIndex: 0,

		entryUC: entryUC,
		entry:   entry,
		inputs:  inputs,
		reseter: func(entry entities.GetEntryResponse, inputs []input.Input) {
			noteData := entry.Data.(entities.EntryDataNote)
			inputs[noteIndex].SetValue(string(noteData))
		},
		validator: func(inputs []input.Input) error {
			if inputs[noteIndex].Value() == "" {
				return errors.New("note should not be empty")
			}
			return nil
		},
		requester: func(inputs []input.Input) (entities.UpdateEntryRequest, error) {
			return entities.UpdateEntryRequest{
				ID:   entry.ID,
				Meta: nil,
				Data: entities.EntryDataNote(inputs[noteIndex].Value()),
			}, nil
		},
	}
}

func NewEntryUpdateBinary(
	title string,
	logger *zap.Logger,
	back base.Component,
	entryUC EntryUpdateUC,
	entry entities.GetEntryResponse,
) *EntryUpdate {
	const (
		pathIndex = iota
		descIndex
		inputCount
	)
	inputs := make([]input.Input, inputCount)
	inputs[pathIndex] = input.NewTextReadonly("filepath", 64)
	inputs[descIndex] = input.NewArea("description", 280)
	return &EntryUpdate{
		title:      title,
		logger:     logger,
		back:       back,
		focusIndex: 0,

		entry:   entry,
		entryUC: entryUC,
		inputs:  inputs,
		validator: func(inputs []input.Input) error {
			if inputs[pathIndex].Value() == "" {
				return errors.New("path should not be empty")
			}
			return nil
		},
		reseter: func(entry entities.GetEntryResponse, inputs []input.Input) {
			path := entry.Meta["filename"]
			inputs[pathIndex].SetValue(path)
		},
		requester: func(inputs []input.Input) (request entities.UpdateEntryRequest, err error) {
			return entities.UpdateEntryRequest{
				ID: entry.ID,
				Meta: map[string]string{
					"filename":    entry.Meta["filename"],
					"description": inputs[descIndex].Value(),
				},
				Data: entry.Data,
			}, nil
		},
	}
}

func (c *EntryUpdate) Title() string {
	return fmt.Sprintf("%s/%s/%s", c.title, c.entry.Type, c.entry.Key)
}

func (c *EntryUpdate) Init() (result base.InitResult) {
	c.syncing = false
	return result.AppendCmd(c.reset())
}

func (c *EntryUpdate) Update(msg tea.Msg) (result base.UpdateResult) {
	switch msg := msg.(type) {
	case entryUpdateMsg:
		return c.updateEntryMsg(msg, result)
	case binaryDownloadMsg:
		return c.updateBinaryDownloadMsg(msg, result)
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

func (c *EntryUpdate) View() string {
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
	if c.entry.Type == core.EntryTypeBinary {
		sb.WriteString(styles.DotStyle)
		sb.WriteString(styles.SubtleStyle.Render("d: download file"))
	}
	return sb.String()
}

func (c *EntryUpdate) updateEntryMsg(
	msg entryUpdateMsg,
	result base.UpdateResult,
) base.UpdateResult {
	c.syncing = false
	if msg.err != nil {
		result.Status = msg.err.Error()
		return result
	}
	result.Status = "entry updated 🔥"
	result.Next = c.back
	return result
}

func (c *EntryUpdate) updateKeyMsg(
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
	case "d":
		if c.entry.Type != core.EntryTypeBinary {
			return result
		}
		if c.syncing {
			result.Status = "🤔"
			return result
		}
		c.syncing = true
		cmd := c.downloadBinaryCmd()
		return result.AppendCmd(cmd)
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
			cmd := c.entryUpdateCmd()
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
		result.AppendCmd(func() tea.Msg { return nil })
	}
	return result
}

func (c *EntryUpdate) reset() (cmd tea.Cmd) {
	c.focusIndex = 0
	for _, v := range c.inputs {
		v.Reset()
	}
	cmd = c.inputs[0].Focus()
	c.reseter(c.entry, c.inputs)
	return cmd
}

func (c *EntryUpdate) entryUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		request, err := c.requester(c.inputs)
		if err != nil {
			c.logger.Error("failed to create entry request", zap.Error(err))
			return entryUpdateMsg{err: err}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err = c.entryUC.Update(ctx, request); err != nil {
			c.logger.Error("failed to update entry", zap.Error(err))
			return entryUpdateMsg{err: err}
		}
		return entryUpdateMsg{
			err: nil,
		}
	}
}

func (c *EntryUpdate) downloadBinaryCmd() tea.Cmd {
	return func() tea.Msg {
		entry := c.entry
		filename := entry.Meta["filename"]
		if filename == "" {
			c.logger.Error("filename is empty")
			return binaryDownloadMsg{err: errors.New("filename is empty")}
		}
		filepath := "./downloads/" + filename
		data := entry.Data.(entities.EntryDataBinary)
		if err := os.MkdirAll("./downloads", 0755); err != nil {
			c.logger.Error("failed to create downloads directory", zap.Error(err))
			return binaryDownloadMsg{err: err}
		}
		if err := os.WriteFile(filepath, data, 0644); err != nil {
			c.logger.Error("failed to write file", zap.Error(err))
			return binaryDownloadMsg{err: err}
		}
		return binaryDownloadMsg{
			err:    nil,
			status: "binary downloaded to " + filepath + " 🔥",
		}
	}
}

func (c *EntryUpdate) updateBinaryDownloadMsg(
	msg binaryDownloadMsg,
	result base.UpdateResult,
) base.UpdateResult {
	c.syncing = false
	if msg.err != nil {
		result.Status = msg.err.Error()
		return result
	}
	result.Status = msg.status
	return result
}
