package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const DirectionalModalHorizontal int = 0
const DirectionalModalVertical int = 1

// DirectionalModal is an extension of Modal that allows for vertically stacked buttons.
type DirectionalModal struct {
	*tview.Box

	// The frame embedded in the modal.
	frame *tview.Frame

	// The forms embedded in the modal's frame.
	forms []*tview.Form

	// The wrapper for the button forms.
	formsFlex *tview.Flex

	// The message text (original, not word-wrapped).
	text string

	// The text color.
	textColor tcell.Color

	// The optional callback for when the user clicked one of the buttons. It
	// receives the index of the clicked button and the button's label.
	done func(buttonIndex int, buttonLabel string)

	// The direction that this modal's buttons should appear (DirectionalModalHorizontal
	// or DirectionalModalVertical)
	direction int

	// The parent application that owns this modal (for focus changes on vertical layouts)
	app *tview.Application

	// The currently selected form (for vertical layouts)
	selected int
}

// NewDirectionalModal returns a new modal message window.
func NewDirectionalModal(direction int, app *tview.Application) *DirectionalModal {
	m := &DirectionalModal{
		Box:       tview.NewBox(),
		textColor: tview.Styles.PrimaryTextColor,
		direction: direction,
		app:       app,
	}
	m.formsFlex = tview.NewFlex()
	if direction == DirectionalModalVertical {
		m.formsFlex.SetDirection(tview.FlexRow)
		m.frame = tview.NewFrame(m.formsFlex).SetBorders(0, 0, 1, 0, 0, 0)
	} else {
		m.formsFlex.SetDirection(tview.FlexColumn)
		form := tview.NewForm().
			SetButtonsAlign(tview.AlignCenter).
			SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
			SetButtonTextColor(tview.Styles.PrimaryTextColor)
		form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
		form.SetCancelFunc(func() {
			if m.done != nil {
				m.done(-1, "")
			}
		})
		m.forms = []*tview.Form{
			form,
		}
		m.frame = tview.NewFrame(form).SetBorders(0, 0, 1, 0, 0, 0)
	}
	m.frame.SetBorder(true).
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)
	return m
}

// SetBackgroundColor sets the color of the modal frame background.
func (m *DirectionalModal) SetBackgroundColor(color tcell.Color) *DirectionalModal {
	for _, form := range m.forms {
		form.SetBackgroundColor(color)
	}
	m.frame.SetBackgroundColor(color)
	return m
}

// SetTextColor sets the color of the message text.
func (m *DirectionalModal) SetTextColor(color tcell.Color) *DirectionalModal {
	m.textColor = color
	return m
}

// SetButtonBackgroundColor sets the background color of the buttons.
func (m *DirectionalModal) SetButtonBackgroundColor(color tcell.Color) *DirectionalModal {
	for _, form := range m.forms {
		form.SetButtonBackgroundColor(color)
	}
	return m
}

// SetButtonTextColor sets the color of the button texts.
func (m *DirectionalModal) SetButtonTextColor(color tcell.Color) *DirectionalModal {
	for _, form := range m.forms {
		form.SetButtonTextColor(color)
	}
	return m
}

// SetDoneFunc sets a handler which is called when one of the buttons was
// pressed. It receives the index of the button as well as its label text. The
// handler is also called when the user presses the Escape key. The index will
// then be negative and the label text an emptry string.
func (m *DirectionalModal) SetDoneFunc(handler func(buttonIndex int, buttonLabel string)) *DirectionalModal {
	m.done = handler
	return m
}

// SetText sets the message text of the window. The text may contain line
// breaks. Note that words are wrapped, too, based on the final size of the
// window.
func (m *DirectionalModal) SetText(text string) *DirectionalModal {
	m.text = text
	return m
}

// AddButtons adds buttons to the window. There must be at least one button and
// a "done" handler so the window can be closed again.
func (m *DirectionalModal) AddButtons(labels []string) *DirectionalModal {
	for index, label := range labels {
		func(i int, l string) {
			if m.direction == DirectionalModalHorizontal {
				m.forms[0].AddButton(label, func() {
					if m.done != nil {
						m.done(i, l)
					}
				})
				button := m.forms[0].GetButton(m.forms[0].GetButtonCount() - 1)
				button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					switch event.Key() {
					case tcell.KeyDown, tcell.KeyRight:
						return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
					case tcell.KeyUp, tcell.KeyLeft:
						return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
					}
					return event
				})
			} else {
				form := tview.NewForm().
					SetButtonsAlign(tview.AlignCenter).
					SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
					SetButtonTextColor(tview.Styles.PrimaryTextColor)
				form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
				form.SetCancelFunc(func() {
					if m.done != nil {
						m.done(-1, "")
					}
				})
				form.AddButton(label, func() {
					if m.done != nil {
						m.done(i, l)
					}
				})
				button := form.GetButton(form.GetButtonCount() - 1)
				button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					switch event.Key() {
					case tcell.KeyDown, tcell.KeyRight, tcell.KeyTab:
						var nextSelection int
						if m.selected == len(m.forms)-1 {
							nextSelection = 0
						} else {
							nextSelection = i + 1
						}
						m.app.SetFocus(m.forms[nextSelection])
						m.selected = nextSelection
					case tcell.KeyUp, tcell.KeyLeft, tcell.KeyBacktab:
						var nextSelection int
						if m.selected == 0 {
							nextSelection = len(m.forms) - 1
						} else {
							nextSelection = i - 1
						}
						m.app.SetFocus(m.forms[nextSelection])
						m.selected = nextSelection
					}
					return event
				})
				m.forms = append(m.forms, form)
				m.formsFlex.AddItem(form, 1, 1, true)
				m.formsFlex.AddItem(nil, 1, 1, false)
			}
		}(index, label)
	}
	return m
}

// ClearButtons removes all buttons from the window.
func (m *DirectionalModal) ClearButtons() *DirectionalModal {
	if m.direction == DirectionalModalHorizontal {
		m.forms[0].ClearButtons()
	} else {
		m.formsFlex.Clear()
	}
	return m
}

// SetFocus shifts the focus to the button with the given index.
func (m *DirectionalModal) SetFocus(index int) *DirectionalModal {
	if m.direction == DirectionalModalHorizontal {
		m.forms[0].SetFocus(index)
	} else {
		m.forms[index].SetFocus(0)
	}
	return m
}

// Focus is called when this primitive receives focus.
func (m *DirectionalModal) Focus(delegate func(p tview.Primitive)) {
	if m.direction == DirectionalModalHorizontal {
		delegate(m.forms[0])
	} else {
		delegate(m.forms[0])
	}
}

// HasFocus returns whether or not this primitive has focus.
func (m *DirectionalModal) HasFocus() bool {
	if m.direction == DirectionalModalHorizontal {
		return m.forms[0].HasFocus()
	}

	for _, form := range m.forms {
		if form.HasFocus() {
			return true
		}
	}
	return false
}

// Draw draws this primitive onto the screen.
func (m *DirectionalModal) Draw(screen tcell.Screen) {
	// Calculate the width of this modal.
	buttonsWidth := 0
	if m.direction == DirectionalModalHorizontal {
		for i := 0; i < m.forms[0].GetButtonCount(); i++ {
			button := m.forms[0].GetButton(i)
			buttonWidth := tview.TaggedStringWidth(button.GetLabel()) + 4 + 2
			buttonsWidth += buttonWidth
		}
	} else {
		for _, form := range m.forms {
			button := form.GetButton(0)
			buttonWidth := tview.TaggedStringWidth(button.GetLabel()) + 4 + 2
			if buttonWidth > buttonsWidth {
				buttonsWidth = buttonWidth
			}
		}
	}
	buttonsWidth -= 2
	screenWidth, screenHeight := screen.Size()
	width := screenWidth / 3
	if width < buttonsWidth {
		width = buttonsWidth
	}
	// width is now without the box border.

	// Reset the text and find out how wide it is.
	m.frame.Clear()
	lines := tview.WordWrap(m.text, width)
	for _, line := range lines {
		m.frame.AddText(line, true, tview.AlignCenter, m.textColor)
	}

	// Set the modal's position and size.
	height := len(lines) + 6
	if m.direction == DirectionalModalVertical {
		height += (len(m.forms) - 1) * 2
	}
	width += 4
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2
	m.SetRect(x, y, width, height)

	// Draw the frame.
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}

// MouseHandler returns the mouse handler for this primitive.
/* TODO
func (m *DirectionalModal) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// Pass mouse events on to the form.
		consumed, capture = m.form.MouseHandler()(action, event, setFocus)
		if !consumed && action == tview.MouseLeftClick && m.InRect(event.Position()) {
			setFocus(m)
			consumed = true
		}
		return
	})
}
*/

// InputHandler returns the handler for this primitive.
func (m *DirectionalModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if m.frame.HasFocus() {
			if handler := m.frame.InputHandler(); handler != nil {
				handler(event, setFocus)
				return
			}
		}
	})
}
