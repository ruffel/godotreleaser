package messages

import "github.com/charmbracelet/lipgloss"

type StageStart struct {
	title string
}

func NewStage(title string) StageStart {
	return StageStart{title: title}
}

func (s StageStart) View() string {
	border := lipgloss.NewStyle().
		BorderBottom(true).
		BorderForeground(lipgloss.Color("63")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderTop(true).
		Padding(0, 1)

	return border.Render(s.title)
}

func (s StageStart) String() string {
	return s.View()
}
