package components

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/styles"
	"strings"
)

var _ base.Component = (*About)(nil)

type (
	About struct {
		title     string
		back      base.Component
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

func (c *About) Init() (result base.InitResult) {
	return result
}

func (c *About) Update(msg tea.Msg) (result base.UpdateResult) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		switch k {
		case "q", "ctrl+c":
			result.Quitting = true
			return result.AppendCmd(tea.Quit)
		case "esc":
			result.Prev = c.back
		}
	}
	return result
}

func (c *About) View() string {
	sb := strings.Builder{}
	_, err := fmt.Fprintf(&sb, `build into:
 - version: %s
 - date:    %s
 - commit:  %s
`,
		c.buildInfo.Version,
		c.buildInfo.Date,
		c.buildInfo.Commit)
	if err != nil { // err unexpected
		panic(err)
	}

	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(styles.SubtleStyle.Render("esc: back"))
	sb.WriteString(styles.DotStyle)
	sb.WriteString(styles.SubtleStyle.Render("q: quit"))
	return sb.String()
}

func (c *About) SetPrev(back base.Component) {
	c.back = back
}
