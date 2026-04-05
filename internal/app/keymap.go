package app

import "github.com/charmbracelet/bubbles/key"

// KeyMap содержит все глобальные горячие клавиши приложения.
type KeyMap struct {
	NextPanel key.Binding
	PrevPanel key.Binding
	Send      key.Binding
	Save      key.Binding
	New       key.Binding
	Search    key.Binding
	Duplicate key.Binding
	Delete    key.Binding
	Help      key.Binding
	Quit      key.Binding
}

// DefaultKeyMap — биндинги по умолчанию.
var DefaultKeyMap = KeyMap{
	NextPanel: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
	PrevPanel: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
	Send:      key.NewBinding(key.WithKeys("ctrl+enter"), key.WithHelp("ctrl+enter", "send request")),
	Save:      key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save request")),
	New:       key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new request")),
	Search:    key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	Duplicate: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "duplicate")),
	Delete:    key.NewBinding(key.WithKeys("delete"), key.WithHelp("del", "delete")),
	Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

// Bindings для экрана помощи (порядок фиксирован).
func (k KeyMap) Bindings() []key.Binding {
	return []key.Binding{
		k.NextPanel, k.PrevPanel, k.Send, k.Save, k.New, k.Search,
		k.Duplicate, k.Delete, k.Help, k.Quit,
	}
}
