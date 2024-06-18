package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

var _ Component = (*About)(nil)

type (
	About struct {
		title     string
		back      Component
		buildInfo BuildInfo
	}
	BuildInfo struct {
		Version string
		Date    string
		Commit  string
	}
)

func NewSettings(
	title string,
	buildInfo BuildInfo,
) *About {
	return &About{
		title:     title,
		buildInfo: buildInfo,
	}
}

func (c *About) Title() string {
	return c.title
}

func (c *About) Init() tea.Cmd {
	return nil
}

func (c *About) Update(msg tea.Msg) (result UpdateResult, cmd tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		switch k {
		case "q", "ctrl+c":
			result.Quitting = true
			return result, tea.Quit
		case "esc":
			result.Prev = c.back
			return result, nil
		}
	}

	return result, nil
}

func (c *About) View() string {
	sb := strings.Builder{}
	sb.WriteString("build info\n")
	sb.WriteString(" - version: ")
	sb.WriteString(c.buildInfo.Version)
	sb.WriteByte('\n')
	sb.WriteString(" - date:    ")
	sb.WriteString(c.buildInfo.Date)
	sb.WriteByte('\n')
	sb.WriteString(" - commit:  ")
	sb.WriteString(c.buildInfo.Commit)
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(subtleStyle.Render("esc: back"))
	sb.WriteString(dotStyle)
	sb.WriteString(subtleStyle.Render("q: quit"))
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	return sb.String()
}

func (c *About) SetPrev(back Component) {
	c.back = back
}
