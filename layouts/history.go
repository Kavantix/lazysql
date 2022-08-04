package layouts

import (
	"fmt"
	"strings"

	"github.com/Kavantix/lazysql/database"
	. "github.com/Kavantix/lazysql/pane"
	"github.com/awesome-gocui/gocui"
)

type HistoryPane struct {
	queries        []*paneableQuery
	g              *gocui.Gui
	pane           *Pane[*paneableQuery]
	lastId         int
	onExecuteQuery func(query database.Query)
}

type paneableQuery struct {
	id    int
	value database.Query
}

func (p *paneableQuery) String() string {
	return fmt.Sprintf("%d: %s", p.id, strings.ReplaceAll(string(p.value), "\n", " "))
}

func (p *paneableQuery) EqualsPaneable(other Paneable) bool {
	return p == other
}

func NewHistoryPane(g *gocui.Gui, onSelectQuery func(query database.Query)) *HistoryPane {
	historyPane := &HistoryPane{
		g:              g,
		pane:           NewPane[*paneableQuery](g, "History"),
		onExecuteQuery: onSelectQuery,
	}
	historyPane.pane.OnSelectItem(func(item *paneableQuery) {
		onSelectQuery(database.Query(item.value))
	})

	return historyPane
}

func (h *HistoryPane) Position(left, top, right, bottom int) {
	h.pane.Position(left, top, right, bottom)
}

func (h *HistoryPane) Paint() {
	h.pane.Paint()
}

func (h *HistoryPane) AddQuery(newQuery database.Query) {
	if len(h.queries) > 0 && h.queries[0].value == newQuery {
		return
	}
	h.lastId += 1
	h.queries = append([]*paneableQuery{{h.lastId, newQuery}}, h.queries...)
	h.pane.SetContent(h.queries)
	h.pane.Selected = h.queries[0]
}
