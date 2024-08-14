package messages

type Footer struct {
	title string
}

func NewFooter(title string) Footer {
	return Footer{title: title}
}

func (f Footer) View() string {
	return border.Render(f.title)
}

func (f Footer) String() string {
	return f.View()
}
