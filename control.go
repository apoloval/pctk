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
func (s VerbSlot) Draw(app *App, frame *Frame, hover RoomItem) {
	rect := s.Rect()
	color := ControlVerbColor
	if frame.MouseIn(rect) {
		color = ControlVerbHoverColor
	}
	if hover != nil {
		switch s.Verb {
		case VerbOpen:
			if hover.ItemClass().IsOneOf(ObjectClassOpenable) {
				color = ControlVerbHoverColor
			}
		case VerbClose:
			if hover.ItemClass().IsOneOf(ObjectClassCloseable) {
				color = ControlVerbHoverColor
			}
		case VerbTalkTo:
			if hover.ItemClass().IsOneOf(ObjectClassPerson) {
				color = ControlVerbHoverColor
			}
		case VerbLookAt:
			if hover.ItemClass().IsNoneOf(ObjectClassOpenable, ObjectClassCloseable, ObjectClassPerson) {
				color = ControlVerbHoverColor
			}
		}
	}

	DrawDefaultText(string(s.Verb), rect.Pos, AlignLeft, color)
}

// Rect returns the rectangle of the verb slot in the screen.
func (v VerbSlot) Rect() Rectangle {
	x := 2 + v.Col*ScreenWidth/6
	y := (v.Row + 1) * FontDefaultSize
	w := ScreenWidth / 6
	h := FontDefaultSize
	return NewRect(x, y, w, h)
}

// ActionSentence is a sentence that represents the action the player is doing.
type ActionSentence struct {
	app  *App // TODO: replace by an interface to send commands
	verb Verb
	args [2]RoomItem
	fut  Future
}

// Init initializes the action sentence.
func (s *ActionSentence) Init(app *App) {
	s.app = app
	s.Reset(VerbWalkTo)
}

// Draw renders the action sentence in the control pane.
func (s *ActionSentence) Draw(hover RoomItem) {
	pos := NewPos(ScreenWidth/2, 0)
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
func (s *ActionSentence) ProcessInventoryClick(obj *Object) {
	if s.args[0] != nil {
		s.interactWith(s.verb, s.args[0], obj)
		return
	}
	switch s.verb {
	case VerbUse, VerbGive:
		if obj.ItemClass().IsOneOf(ObjectClassApplicable) {
			s.args[0] = obj
			return
		}
	}
	s.interactWith(s.verb, obj, nil)
}

// ProcessLeftClick processes a left click in the control pane.
func (s *ActionSentence) ProcessLeftClick(click Position, item RoomItem) {
	if item == nil {
		if s.verb == VerbWalkTo || s.fut != nil {
			s.walkToPos(s.app, click)
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
			s.interactWith(s.verb, item, nil)
		} else {
			// Item is candidate to second argument.
			s.interactWith(s.verb, s.args[0], item)
		}
	}
}

// ProcessRightClick processes a right click in the control pane.
func (s *ActionSentence) ProcessRightClick(app *App, click Position, item RoomItem) {
	if item != nil {
		// Execute quick action
		if item.ItemClass().IsOneOf(ObjectClassPerson) {
			s.interactWith(VerbTalkTo, item, nil)
		} else if item.ItemClass().IsOneOf(ObjectClassOpenable) {
			s.interactWith(VerbOpen, item, nil)
		} else if item.ItemClass().IsOneOf(ObjectClassCloseable) {
			s.interactWith(VerbClose, item, nil)
		} else {
			s.interactWith(VerbLookAt, item, nil)
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
		case VerbWalkTo:
			return item.ItemOwner() == nil
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

func (s *ActionSentence) interactWith(verb Verb, item, other RoomItem) {
	s.verb = verb
	s.args[0] = item
	s.args[1] = other

	var cmd Command
	if verb == VerbWalkTo {
		if item.ItemOwner() != nil {
			// Cannot walk to an object in the inventory
			s.Reset(VerbWalkTo)
			return
		}
		cmd = ActorWalkToItem{
			Actor: s.app.ego,
			Item:  item,
		}
	} else {
		cmd = ActorInteractWith{
			Actor:   s.app.ego,
			Targets: [2]RoomItem{item, other},
			Verb:    verb,
		}
	}
	s.fut = s.app.RunCommandSequence(
		cmd,
		CommandFunc(func(app *App) (any, error) {
			s.Reset(VerbWalkTo)
			return nil, nil
		}),
	)
}

func (s *ActionSentence) walkToPos(app *App, pos Position) {
	app.RunCommand(ActorWalkToPosition{
		Actor:    app.ego,
		Position: pos,
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
func (c *ControlInventory) Draw(app *App, frame *Frame) {
	if app.ego == nil {
		return
	}
	for i, item := range app.ego.Inventory() {
		rect := c.slotsRect[i]
		color := ControlInventoryColor
		if frame.MouseIn(rect) {
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
			FontDefaultSize*(i+1),
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
func (c *ControlSentenceChoice) Draw(frame *Frame) {
	for i, sentence := range c.Sentences {
		rect := NewRect(
			SentenceChoiceMagin,
			SentenceChoiceMagin+FontDefaultSize*(i),
			ScreenWidth-2*SentenceChoiceMagin,
			FontDefaultSize,
		)
		color := Green
		if frame.MouseIn(rect) {
			color = Yellow
		}

		// Hack: remove EOL characters from the sentence so programmers must include line breaks
		// manually.
		sentence = strings.ReplaceAll(sentence, "\n", " ")

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
			SentenceChoiceMagin+FontDefaultSize*(i),
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

	action    ActionSentence
	camera    Camera
	choice    *ControlSentenceChoice
	hover     RoomItem
	inventory ControlInventory
	verbs     []VerbSlot
}

// Init initializes the control pane.
func (p *ControlPane) Init(app *App, cam Camera, vp *Viewport) {
	p.camera = cam.WithTarget(NewPos(0, -ViewportHeight))

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
	p.action.Init(app)
	p.inventory.Init()
	vp.SubscribeEventHandler(func(e ViewportEvent) {
		if p.Mode != ControlPaneNormal {
			return
		}
		switch e.Type {
		case ViewportEventMouseEnter:
			p.hover = e.Item
		case ViewportEventMouseLeave:
			p.hover = nil
		case ViewportEventLeftClick:
			p.action.ProcessLeftClick(e.Pos, e.Item)
		case ViewportEventRightClick:
			p.action.ProcessRightClick(app, e.Pos, e.Item)
		}
	})
}

func (p *ControlPane) ProcessFrame(app *App, frame *Frame) {
	frame.WithCamera(&p.camera, func(f *Frame) {
		p.draw(app, f)
		p.processInputs(app, f)
	})
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

func (p *ControlPane) draw(app *App, frame *Frame) {
	switch p.Mode {
	case ControlPaneDisabled:
		return
	case ControlPaneNormal:
		for _, v := range p.verbs {
			v.Draw(app, frame, p.hover)
		}
		p.action.Draw(p.hover)
		p.inventory.Draw(app, frame)
	case ControlPaneDialog:
		if p.choice != nil {
			p.choice.Draw(frame)
		}
	}
}

func (p *ControlPane) processMouseOver(app *App, frame *Frame) {
	if frame.MouseIn(ControlPaneRect) {
		if obj := p.inventory.ObjectAt(app, frame.MouseRelativePos()); obj != nil {
			p.hover = obj
			return
		}
		p.hover = nil
	}
}

func (p *ControlPane) processInputs(app *App, frame *Frame) {
	if !frame.Mouse.Enabled || app.ego == nil {
		return
	}

	p.processMouseOver(app, frame)

	mpos := frame.MouseRelativePos()
	switch p.Mode {
	case ControlPaneNormal:
		if frame.Mouse.LeftClick() && frame.MouseIn(ControlPaneRect) {
			p.normalProcessLeftClick(app, mpos)
		}
		if frame.Mouse.RightClick() && p.hover != nil {
			p.action.ProcessRightClick(app, mpos, p.hover)
		}
	case ControlPaneDialog:
		if frame.Mouse.LeftClick() && p.choice.ProcessLeftClick(mpos) {
			p.Mode = ControlPaneDisabled
			p.choice = nil
		}
	}
	if app.debugMode && rl.IsKeyPressed(rl.KeyD) { // TODO: have a keyboard input in the frame
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
	if obj := p.inventory.ObjectAt(app, click); obj != nil {
		p.action.ProcessInventoryClick(obj)
	}
}
