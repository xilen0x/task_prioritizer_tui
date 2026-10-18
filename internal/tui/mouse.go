package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// onMouse maps clicks on a best-effort basis: rows, the focus box and the
// sidebar columns select a task, a free slot opens the editor, the
// function-key bar acts like its key, the wheel moves the selection and
// any click closes help/about. The editor and confirm dialogs are
// keyboard-only.
func (m Model) onMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	l := m.layout()
	if l.tooSmall || msg.Action != tea.MouseActionPress {
		return m, nil
	}
	if m.editor != nil || m.confirmID != 0 {
		return m, nil
	}
	if m.help || m.about {
		if msg.Button == tea.MouseButtonLeft {
			m.help, m.about = false, false
		}
		return m, nil
	}
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		return m.onMainKey("up")
	case tea.MouseButtonWheelDown:
		return m.onMainKey("down")
	case tea.MouseButtonLeft:
	default:
		return m, nil
	}

	x, y := msg.X, msg.Y
	if y == l.h-1 {
		return m.clickFKey(x, l.compact)
	}
	if x >= marginX && x < marginX+l.leftW {
		if y >= headerRows && y < headerRows+l.focusH {
			return m.selectID(m.board.FocusID)
		}
		if len(m.board.Tasks) == 0 {
			return m, nil
		}
		others := m.others()
		for i := 0; i < 3; i++ {
			if top := l.rowTop(i); y >= top && y < top+l.rowH {
				if i < len(others) {
					return m.selectID(others[i].ID)
				}
				m.openNew()
				return m, nil
			}
		}
		return m, nil
	}
	cellW := (sideWidth - 2) / 4
	x0 := l.sideX() + 1
	if x >= x0 && x < x0+4*cellW && y >= headerRows && y < headerRows+l.barH+sideChrome {
		if i := (x - x0) / cellW; i < len(m.board.Tasks) {
			return m.selectID(m.board.Tasks[i].ID)
		}
	}
	return m, nil
}

func (m Model) selectID(id int) (tea.Model, tea.Cmd) {
	if !m.board.Has(id) {
		return m, nil
	}
	return m.mutate(func(m *Model) { m.board.Select(id) })
}

// clickFKey triggers the function key under column x of the key bar.
func (m Model) clickFKey(x int, compact bool) (tea.Model, tea.Cmd) {
	pos := marginX
	keys := m.fkeys()
	for i, seg := range m.fkeySegments(compact) {
		w := lenW(seg)
		if x >= pos && x < pos+w {
			f := keys[i]
			if f.disabled {
				return m, nil
			}
			k := strings.ToLower(f.key)
			if k == "f10" {
				return m.quit()
			}
			return m.onMainKey(k)
		}
		pos += w
	}
	return m, nil
}
