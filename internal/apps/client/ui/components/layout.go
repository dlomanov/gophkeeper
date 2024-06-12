package components

import (
	"strings"
)

type Layout struct {
}

func NewLayout() *Layout {
	return &Layout{}
}

func (l *Layout) View(c Component, quitting bool, status string) string {
	sb := strings.Builder{}
	sb.WriteByte('\n')
	sb.WriteString(c.Title())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(c.View())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(statusStyle.Render(status))
	if quitting {
		sb.WriteString("\nSee you later! 😊\n\n")
	}

	return mainStyle.Render(sb.String())
}
