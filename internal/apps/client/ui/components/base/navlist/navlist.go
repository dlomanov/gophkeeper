package navlist

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base"
	"io"
	"strings"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(2)
)

type (
	List struct {
		inner list.Model
		items []Item
	}
	Item struct {
		Name string
		Next base.Component
	}

	listItem         string
	listItemDelegate struct{}
)

func (i listItem) FilterValue() string                             { return "" }
func (d listItemDelegate) Height() int                             { return 1 }
func (d listItemDelegate) Spacing() int                            { return 0 }
func (d listItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d listItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(listItem)
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

func New(items []Item) List {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = listItem(item.Name)
	}
	l := list.New(listItems, listItemDelegate{}, 20, len(items))
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.SetShowPagination(false)
	l.Styles.PaginationStyle = paginationStyle
	l.SetShowHelp(false)
	return List{inner: l, items: items}
}

func (l List) Selected() *Item {
	i := l.inner.SelectedItem().(listItem)
	choice := string(i)
	for _, v := range l.items {
		if v.Name == choice {
			return &v
		}
	}
	return nil
}

func (l List) SetWidth(width int) {
	l.inner.SetWidth(width)
}

func (l List) Update(msg tea.Msg) (List, tea.Cmd) {
	var cmd tea.Cmd
	l.inner, cmd = l.inner.Update(msg)
	return l, cmd
}

func (l List) View() string {
	return l.inner.View()
}
