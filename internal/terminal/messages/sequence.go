package messages

import "github.com/charmbracelet/lipgloss"

var border = lipgloss.NewStyle(). //nolint:gochecknoglobals
					BorderStyle(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("63")).
					Width(72). //nolint:mnd
					Padding(0, 1)

var titleStyle = lipgloss.NewStyle(). //nolint:gochecknoglobals
					Bold(true).
					PaddingBottom(1).
					Width(72). //nolint:mnd
					Align(lipgloss.Center)

type SequenceStart struct {
	title string
}

func NewSequence(title string) SequenceStart {
	return SequenceStart{title: title}
}

func (s SequenceStart) View() string {
	return border.Render(lipgloss.JoinHorizontal(lipgloss.Top, titleStyle.Render(s.title)))
}

func (s SequenceStart) String() string {
	return s.View()
}
