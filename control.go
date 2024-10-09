package pctk

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	ControlActionColor         = Cyan
	ControlActionOngoingColor  = BrigthCyan
	ControlInventoryColor      = Magenta
	ControlInventoryHoverColor = BrigthMagenta
	ControlVerbColor           = Green
	ControlVerbHoverColor      = Yellow
)

// Verb is a type that represents the action verb.
type Verb string

const (
	VerbOpen    Verb = "Open"
	VerbClose   Verb = "Close"
	VerbPush    Verb = "Push"
	VerbPull    Verb = "Pull"
	VerbWalkTo  Verb = "Walk to"
	VerbPickUp  Verb = "Pick up"
	VerbTalkTo  Verb = "Talk to"
	VerbGive    Verb = "Give"
	VerbUse     Verb = "Use"
	VerbLookAt  Verb = "Look at"
	VerbTurnOn  Verb = "Turn on"
	VerbTurnOff Verb = "Turn off"
)

const (
	// SentenceChoiceMagin is the margin of sentence choice in the control pane.
	SentenceChoiceMagin = 2
)

// Action returns the action codename for the verb.
func (v Verb) Action() string {
	action := strings.ToLower(string(v))
	action = strings.ReplaceAll(action, " ", "")
	return action
}

// VerbSlot is a slot in the control panel that holds a verb.
type VerbSlot struct {
	Verb Verb
	Row  int
	Col  int
}

// Draw renders the verb slot in the control pane.
func (s VerbSlot) Draw(app *App, m *MouseCursor) {
	rect := s.Rect()
	color := ControlVerbColor
	if m.IsInto(rect) {
		color = ControlVerbHoverColor
	}
	if room := app.room; room != nil {
		if item := room.ItemAt(m.Position()); item != nil {
			switch s.Verb {
			case VerbOpen:
				if item.ItemClass().IsOneOf(ObjectClassOpenable) {
					color = ControlVerbHoverColor
				}
			case VerbClose:
				if item.ItemClass().IsOneOf(ObjectClassCloseable) {
					color = ControlVerbHoverColor
				}
			case VerbTalkTo:
				if item.ItemClass().IsOneOf(ObjectClassPerson) {
					color = ControlVerbHoverColor
				}
			case VerbLookAt:
				if item.ItemClass().IsNoneOf(ObjectClassOpenable, ObjectClassCloseable, ObjectClassPerson) {
					color = ControlVerbHoverColor
				}
			}
		}
	}

	DrawDefaultText(string(s.Verb), rect.Pos, AlignLeft, color)
}

// Rect returns the rectangle of the verb slot in the screen.
func (v VerbSlot) Rect() Rectangle {
	x := 2 + v.Col*ScreenWidth/6
	y := ViewportHeight + (v.Row+1)*FontDefaultSize
	w := ScreenWidth / 6
	h := FontDefaultSize
	return NewRect(x, y, w, h)
}

// ActionSentence is a sentence that represents the action the player is doing.
type ActionSentence struct {
	verb Verb
	args [2]RoomItem
	fut  Future
}

// Draw renders the action sentence in the control pane.
func (s *ActionSentence) Draw(app *App, hover RoomItem) {
	pos := NewPos(ScreenWidth/2, ViewportHeight)
	action := s.line()
	color := ControlActionColor
	if s.fut != nil {
		// Ongoing action.
		color = ControlActionOngoingColor
		DrawDefaultText(action, pos, AlignCenter, color)
		return
	}

	// Sentence incompleted. Check if hover exists and must be added to the sentence.
	if s.admits(hover) {
		action = action + " " + hover.Caption()
	}
	DrawDefaultText(action, pos, AlignCenter, color)
}

// ProcessInventoryClick processes a click in the inventory.
func (s *ActionSentence) ProcessInventoryClick(app *App, obj *Object) {
	if s.args[0] != nil {
		s.interactWith(app, s.verb, s.args[0], obj)
		return
	}
	switch s.verb {
	case VerbUse, VerbGive:
		if obj.ItemClass().IsOneOf(ObjectClassApplicable) {
			s.args[0] = obj
			return
		}
	}
	s.interactWith(app, s.verb, obj, nil)
}

// ProcessLeftClick processes a left click in the control pane.
func (s *ActionSentence) ProcessLeftClick(app *App, click Position, item RoomItem) {
	if item == nil {
		if s.verb == VerbWalkTo || s.fut != nil {
			s.walkToPos(app, click)
		}
		return
	}
	if s.admits(item) {
		if s.args[0] == nil {
			// Item is candidate to first argument.
			if s.verb == VerbUse && item.ItemClass().IsAllOf(ObjectClassApplicable) {
				// Special case for use verb on a room item that is applicable.
				s.args[0] = item
				return
			}
			s.interactWith(app, s.verb, item, nil)
		} else {
			// Item is candidate to second argument.
			s.interactWith(app, s.verb, s.args[0], item)
		}
	}
}

// ProcessRightClick processes a right click in the control pane.
func (s *ActionSentence) ProcessRightClick(app *App, click Position, item RoomItem) {
	if item != nil {
		// Execute quick action
		if item.ItemClass().IsOneOf(ObjectClassPerson) {
			s.interactWith(app, VerbTalkTo, item, nil)
		} else if item.ItemClass().IsOneOf(ObjectClassOpenable) {
			s.interactWith(app, VerbOpen, item, nil)
		} else if item.ItemClass().IsOneOf(ObjectClassCloseable) {
			s.interactWith(app, VerbClose, item, nil)
		} else {
			s.interactWith(app, VerbLookAt, item, nil)
		}
		return
	}
	// No item there. Only respond if current verb is walk to.
	if s.verb == VerbWalkTo {
		s.walkToPos(app, click)
	}
}

func (s *ActionSentence) admits(item RoomItem) bool {
	if item == nil || s.fut != nil {
		return false
	}
	if s.args[0] == nil {
		// Item is candidate to first argument.
		switch s.verb {
		case VerbTalkTo:
			return item.ItemClass().IsOneOf(ObjectClassPerson)
		case VerbOpen, VerbClose, VerbPickUp, VerbGive, VerbTurnOn, VerbTurnOff:
			return !item.ItemClass().IsOneOf(ObjectClassPerson)
		case VerbUse:
			return !item.ItemClass().IsOneOf(ObjectClassPerson)
		default:
			return true
		}
	}

	// Item is candidate to second argument.
	if !s.args[0].ItemClass().IsOneOf(ObjectClassApplicable) || item == s.args[0] {
		// First argument is not applicable or is the same item.
		return false
	}
	switch s.verb {
	case VerbGive:
		return item.ItemClass().IsOneOf(ObjectClassPerson)
	default:
		return true
	}
}

func (s *ActionSentence) line() string {
	line := string(s.verb)
	if s.args[0] != nil {
		line += " " + s.args[0].Caption()
		switch s.verb {
		case VerbUse:
			if s.args[0].ItemClass().IsOneOf(ObjectClassApplicable) {
				line += " with"
			}
		case VerbGive:
			line += " to"
		}
	}
	if s.args[1] != nil {
		line += " " + s.args[1].Caption()
	}
	return line
}

func (s *ActionSentence) interactWith(app *App, verb Verb, item, other RoomItem) {
	s.verb = verb
	s.args[0] = item
	s.args[1] = other

	var cmd Command
	if verb == VerbWalkTo {
		cmd = ActorWalkToItem{
			Actor: app.ego,
			Item:  item,
		}
	} else {
		cmd = ActorInteractWith{
			Actor:   app.ego,
			Targets: [2]RoomItem{item, other},
			Verb:    verb,
		}
	}
	s.fut = app.RunCommandSequence(
		cmd,
		CommandFunc(func(app *App) (any, error) {
			s.Reset(VerbWalkTo)
			return nil, nil
		}),
	)
}

func (s *ActionSentence) walkToPos(app *App, click Position) {
	app.RunCommand(ActorWalkToPosition{
		Actor:    app.ego,
		Position: click,
	})
	s.Reset(VerbWalkTo)
}

// Reset resets the action sentence to the given verb.
func (s *ActionSentence) Reset(verb Verb) {
	s.verb = verb
	s.args[0] = nil
	s.args[1] = nil
	s.fut = nil
}

// ControlInventory is a screen control that shows the inventory.
type ControlInventory struct {
	slotsRect [6]Rectangle
}

// Draw renders the inventory in the control pane.
func (c *ControlInventory) Draw(app *App, m *MouseCursor) {
	if app.ego == nil {
		return
	}
	mpos := m.Position()
	for i, item := range app.ego.Inventory() {
		rect := c.slotsRect[i]
		color := ControlInventoryColor
		if rect.Contains(mpos) {
			color = ControlInventoryHoverColor
		}
		DrawDefaultText(item.Caption(), rect.Pos, AlignLeft, color)
	}
}

// Init initializes the control inventory.
func (c *ControlInventory) Init() {
	arrowsWidth := 32
	for i := range c.slotsRect {
		c.slotsRect[i] = NewRect(
			2+3*ScreenWidth/6+arrowsWidth,
			ViewportHeight+FontDefaultSize*(i+1),
			2*ScreenWidth/6,
			FontDefaultSize,
		)
	}
}

// ObjectAt returns the object at the given position in the inventory box.
func (c *ControlInventory) ObjectAt(app *App, pos Position) *Object {
	if app.ego == nil {
		return nil
	}
	inv := app.ego.Inventory()
	for i, rect := range c.slotsRect {
		if rect.Contains(pos) {
			if i < len(inv) {
				return inv[i]
			}
			return nil
		}
	}
	return nil
}

// ControlPaneMode is the mode of the control pane.
type ControlPaneMode int

const (
	ControlPaneDisabled ControlPaneMode = iota
	ControlPaneNormal
	ControlPaneDialog
)

// IndexedSentence is a sentence with an index. It is the result of a ControlSentenceChoice when
// the player chooses a sentence.
type IndexedSentence struct {
	Index    int
	Sentence string
}

// ControlSentenceChoice is a choice of sentences to show in the control pane.
type ControlSentenceChoice struct {
	Sentences []string
	done      *Promise
}

func NewControlSentenceChoice() *ControlSentenceChoice {
	return &ControlSentenceChoice{
		done: NewPromise(),
	}
}

// Abort the sentence choice.
func (c *ControlSentenceChoice) Abort() {
	c.done.Break()
}

// Add a new sentence to the choice.
func (c *ControlSentenceChoice) Add(sentence string) {
	c.Sentences = append(c.Sentences, sentence)
}

// Draw the sentence choice in the control pane.
func (c *ControlSentenceChoice) Draw(mouse Position) {
	for i, sentence := range c.Sentences {
		rect := NewRect(
			SentenceChoiceMagin,
			ViewportHeight+SentenceChoiceMagin+FontDefaultSize*(i),
			ScreenWidth-2*SentenceChoiceMagin,
			FontDefaultSize,
		)
		color := Green
		if rect.Contains(mouse) {
			color = Yellow
		}
		DrawDefaultText(sentence, rect.Pos, AlignLeft, color)
	}
}

// Done returns a future that will be completed when the player chooses a sentence.
func (c *ControlSentenceChoice) Done() Future {
	if c.done == nil {
		c.done = new(Promise)
	}
	return c.done
}

// ProcessLeftClick processes a left click in the sentence choice. Returns true if the click has
// selected a choice.
func (c *ControlSentenceChoice) ProcessLeftClick(pos Position) bool {
	for i, sent := range c.Sentences {
		rect := NewRect(
			SentenceChoiceMagin,
			ViewportHeight+SentenceChoiceMagin+FontDefaultSize*(i),
			ScreenWidth-2*SentenceChoiceMagin,
			FontDefaultSize,
		)
		if rect.Contains(pos) {
			c.done.CompleteWithValue(IndexedSentence{
				Index:    i,
				Sentence: sent,
			})
			return true
		}
	}
	return false
}

// ControlPane is the screen control pane that shows the action, verbs and inventory.
type ControlPane struct {
	Mode ControlPaneMode // Mode is the mode of the control pane (default, dialog...)

	action ActionSentence
	cursor *MouseCursor
	choice *ControlSentenceChoice
	inv    ControlInventory
	verbs  []VerbSlot
}

// Init initializes the control pane.
func (p *ControlPane) Init(cam *rl.Camera2D) {
	p.Mode = ControlPaneDisabled
	p.verbs = []VerbSlot{
		{Verb: VerbOpen, Row: 0, Col: 0},
		{Verb: VerbClose, Row: 1, Col: 0},
		{Verb: VerbPush, Row: 2, Col: 0},
		{Verb: VerbPull, Row: 3, Col: 0},

		{Verb: VerbWalkTo, Row: 0, Col: 1},
		{Verb: VerbPickUp, Row: 1, Col: 1},
		{Verb: VerbTalkTo, Row: 2, Col: 1},
		{Verb: VerbGive, Row: 3, Col: 1},

		{Verb: VerbUse, Row: 0, Col: 2},
		{Verb: VerbLookAt, Row: 1, Col: 2},
		{Verb: VerbTurnOn, Row: 2, Col: 2},
		{Verb: VerbTurnOff, Row: 3, Col: 2},
	}
	p.action.Reset(VerbWalkTo)
	p.inv.Init()
	p.cursor = NewMouseCursor(cam)
}

// Draw renders the control panel in the viewport.
func (p *ControlPane) Draw(app *App) {
	switch p.Mode {
	case ControlPaneDisabled:
		return
	case ControlPaneNormal:
		for _, v := range p.verbs {
			v.Draw(app, p.cursor)
		}
		hover := p.hover(app, p.cursor.Position())
		p.action.Draw(app, hover)
		p.inv.Draw(app, p.cursor)
		p.cursor.Draw()
	case ControlPaneDialog:
		if p.choice != nil {
			p.choice.Draw(p.cursor.Position())
		}
		p.cursor.Draw()
	}
}

// Disable the control pane.
func (p *ControlPane) Disable() {
	if p.choice != nil {
		p.choice.Abort()
		p.choice = nil
	}
	p.Mode = ControlPaneDisabled
}

// Enable the control pane.
func (p *ControlPane) Enable() {
	if p.choice != nil {
		p.choice.Abort()
		p.choice = nil
	}
	p.Mode = ControlPaneNormal
}

// NewSentenceChoice creates a new sentence choice and sets the control pane to dialog mode.
func (p *ControlPane) NewSentenceChoice() *ControlSentenceChoice {
	p.Mode = ControlPaneDialog
	p.choice = NewControlSentenceChoice()
	return p.choice
}

func (p *ControlPane) hover(app *App, pos Position) RoomItem {
	var item RoomItem
	if ViewportRect.Contains(pos) && app.room != nil {
		item = app.room.ItemAt(pos)
	} else if ControlPaneRect.Contains(pos) {
		if obj := p.inv.ObjectAt(app, pos); obj != nil {
			item = obj
		}
	}
	return item
}

func (p *ControlPane) processControlInputs(app *App) {
	if !p.cursor.Enabled || app.ego == nil {
		return
	}
	pos := p.cursor.Position()
	hover := p.hover(app, pos)
	switch p.Mode {
	case ControlPaneNormal:
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			if ViewportRect.Contains(pos) {
				p.action.ProcessLeftClick(app, pos, hover)
			}
			if ControlPaneRect.Contains(pos) {
				p.normalProcessLeftClick(app, pos)
			}
		} else if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
			if ViewportRect.Contains(pos) {
				p.action.ProcessRightClick(app, pos, hover)
			}
		}
	case ControlPaneDialog:
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			if p.choice.ProcessLeftClick(pos) {
				p.Mode = ControlPaneDisabled
				p.choice = nil
			}
		}
	}
	if app.debugMode && rl.IsKeyPressed(rl.KeyD) {
		app.debugEnabled = !app.debugEnabled
	}
}

func (p *ControlPane) normalProcessLeftClick(app *App, click Position) {
	for _, v := range p.verbs {
		if v.Rect().Contains(click) {
			p.action.Reset(v.Verb)
			return
		}
	}
	if obj := p.inv.ObjectAt(app, click); obj != nil {
		p.action.ProcessInventoryClick(app, obj)
	}
}
