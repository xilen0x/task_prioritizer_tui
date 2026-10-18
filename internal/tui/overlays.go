package tui

import tea "github.com/charmbracelet/bubbletea"

// editor is the state of the edit/new modal (implemented with the
// overlays).
type editor struct{}

func (m *Model) openEdit()  {}
func (m *Model) openNew()   {}
func (m *Model) askDelete() {}

func (m Model) onConfirmKey(string) (tea.Model, tea.Cmd)    { return m, nil }
func (m Model) onEditorKey(tea.KeyMsg) (tea.Model, tea.Cmd) { return m, nil }
func (m Model) onMouse(tea.MouseMsg) (tea.Model, tea.Cmd)   { return m, nil }
func (m Model) overlay(base []string) []string              { return base }
