package components

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"strings"
)

const listHeight = 14

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(2)
)

var _ Component = (*Menu)(nil)

type (
	Menu struct {
		title  string
		navs   []Nav
		list   list.Model
		choice string
	}
	Nav struct {
		Name string
		Next Component
	}
	item         string
	itemDelegate struct{}
)

func (i item) FilterValue() string                             { return "" }
func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

func NewMenu(title string, navs []Nav) *Menu {
	items := make([]list.Item, len(navs))
	for i, nav := range navs {
		items[i] = item(nav.Name)
	}
	l := list.New(items, itemDelegate{}, 20, 7)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.SetShowHelp(false)
	return &Menu{
		title: title,
		navs:  navs,
		list:  l,
	}
}

func (c *Menu) Title() string {
	return c.title
}

func (c *Menu) Init() tea.Cmd {
	return nil
}

func (c *Menu) Update(msg tea.Msg) (result UpdateResult, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.list.SetWidth(msg.Width)
		return result, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		switch k {
		case "q", "ctrl+c":
			result.Quitting = true
			return result, tea.Quit
		case "enter":
			i, ok := c.list.SelectedItem().(item)
			if ok {
				c.choice = string(i)
				for _, v := range c.navs {
					if v.Name == c.choice {
						result.Next = v.Next
						break
					}
				}
			}
			return result, nil
		}
	}

	c.list, cmd = c.list.Update(msg)
	return result, cmd
}

func (c *Menu) View() string {
	sb := strings.Builder{}
	sb.WriteString(c.list.View())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(subtleStyle.Render("q: quit"))
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	return sb.String()
}
