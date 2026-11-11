package screens

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"xadrez/cmd/frontend/components"
	"xadrez/internal/application/messages"
	"xadrez/internal/domain/chess"

	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type MatchClient interface {
	MakeMove(request messages.MoveRequest) error
	Resign(request messages.ResignRequest) error
}

type matchInbox struct {
	mu     sync.Mutex
	match  *messages.MatchResponse
	room   *messages.RoomResponse
	errors []string
}

type matchAction int

const (
	actionNone matchAction = iota
	actionResign
	actionConfirmResign
	actionCancelResign
	actionLeave
	actionPromote
	actionCancelPromotion
)

const (
	margin       float32 = 20
	sidebarWidth float32 = 200
	lineHeight   int32   = 22
	textSize     int32   = 18
)

var promotionChoices = []struct {
	label string
	kind  chess.PieceKind
}{
	{"Rainha", chess.Queen},
	{"Torre", chess.Rook},
	{"Bispo", chess.Bishop},
	{"Cavalo", chess.Knight},
}

type MatchScreen struct {
	size   rl.Vector2
	user   string
	client MatchClient
	board  *components.BoardRenderer
	back   func() components.Screen
	next   components.Screen
	inbox  matchInbox

	match         messages.MatchResponse
	spectators    []string
	selected      *components.Square
	promotion     *messages.MoveRequest
	waiting       bool
	confirmResign bool
	errorMessage  string

	action        matchAction
	promotionKind chess.PieceKind
}

func NewMatchScreen(size rl.Vector2, user string, match messages.MatchResponse, client MatchClient, board *components.BoardRenderer, back func() components.Screen) *MatchScreen {
	s := &MatchScreen{
		size:   size,
		user:   user,
		client: client,
		board:  board,
		back:   back,
		match:  match,
	}
	s.next = s
	return s
}

func (s *MatchScreen) PushMatch(match messages.MatchResponse) {
	s.inbox.mu.Lock()
	defer s.inbox.mu.Unlock()
	s.inbox.match = &match
}

func (s *MatchScreen) PushRoom(room messages.RoomResponse) {
	s.inbox.mu.Lock()
	defer s.inbox.mu.Unlock()
	s.inbox.room = &room
}

func (s *MatchScreen) PushError(message string) {
	s.inbox.mu.Lock()
	defer s.inbox.mu.Unlock()
	s.inbox.errors = append(s.inbox.errors, message)
}

func (s *MatchScreen) HandleEvent(event components.Event) {
	if message, ok := errorMessage(event); ok {
		s.PushError(message)
		return
	}
	switch event.Type {
	case "match_updated", "make_move_ok", "resign_ok", "get_match_ok":
		if match, ok := event.Payload.(messages.MatchResponse); ok && match.Room == s.match.Room {
			s.PushMatch(match)
		}
	case "list_rooms_ok":
		if response, ok := event.Payload.(messages.ListRoomsResponse); ok {
			for _, room := range response.Rooms {
				if room.Room == s.match.Room {
					s.PushRoom(room)
				}
			}
		}
	}
}

func (s *MatchScreen) Update(ctx context.Context) (components.Screen, error) {
	s.drainInbox()
	s.handleAction()
	s.handleBoardClick()
	return s.next, nil
}

func (s *MatchScreen) drainInbox() {
	s.inbox.mu.Lock()
	match, room, errors := s.inbox.match, s.inbox.room, s.inbox.errors
	s.inbox.match, s.inbox.room, s.inbox.errors = nil, nil, nil
	s.inbox.mu.Unlock()

	if match != nil {
		s.match = *match
		s.waiting = false
		s.selected = nil
		s.promotion = nil
		if s.finished() {
			s.confirmResign = false
		}
	}
	if room != nil {
		s.spectators = slices.DeleteFunc(slices.Clone(room.Participants), func(name string) bool {
			return name == s.match.White || name == s.match.Black
		})
	}
	if len(errors) > 0 {
		s.errorMessage = errors[len(errors)-1]
		s.waiting = false
	}
}

func (s *MatchScreen) handleAction() {
	action := s.action
	s.action = actionNone

	switch action {
	case actionResign:
		s.confirmResign = true
	case actionCancelResign:
		s.confirmResign = false
	case actionConfirmResign:
		s.confirmResign = false
		s.selected = nil
		s.promotion = nil
		s.send(func() error {
			return s.client.Resign(messages.ResignRequest{Room: s.match.Room})
		})
	case actionLeave:
		s.next = s.back()
	case actionPromote:
		if s.promotion != nil {
			request := *s.promotion
			request.Promotion = s.promotionKind
			s.promotion = nil
			s.sendMove(request)
		}
	case actionCancelPromotion:
		s.promotion = nil
	}
}

func (s *MatchScreen) handleBoardClick() {
	if !rl.IsMouseButtonPressed(rl.MouseButtonLeft) || !s.canMove() || s.promotion != nil || s.confirmResign {
		return
	}
	square, ok := s.layout().SquareAt(rl.GetMousePosition())
	if !ok {
		return
	}

	color, _ := s.color()
	piece := s.match.Pieces[square.X][square.Y]
	if piece != nil && piece.Color == color {
		if s.selected != nil && *s.selected == square {
			s.selected = nil
		} else {
			s.selected = &square
		}
		return
	}
	if s.selected == nil {
		return
	}

	request := messages.MoveRequest{
		Room:  s.match.Room,
		FromX: s.selected.X,
		FromY: s.selected.Y,
		ToX:   square.X,
		ToY:   square.Y,
	}
	s.selected = nil
	if s.isPromotion(request) {
		s.promotion = &request
		return
	}
	s.sendMove(request)
}

func (s *MatchScreen) sendMove(request messages.MoveRequest) {
	s.send(func() error {
		return s.client.MakeMove(request)
	})
}

func (s *MatchScreen) send(request func() error) {
	s.errorMessage = ""
	if err := request(); err != nil {
		s.errorMessage = err.Error()
		return
	}
	s.waiting = true
}

func (s *MatchScreen) color() (chess.PieceColor, bool) {
	switch s.user {
	case s.match.White:
		return chess.White, true
	case s.match.Black:
		return chess.Black, true
	}
	return chess.White, false
}

func (s *MatchScreen) finished() bool {
	return s.match.Outcome != chess.OutcomeNone
}

func (s *MatchScreen) canMove() bool {
	color, isPlayer := s.color()
	return isPlayer && !s.finished() && !s.waiting && s.match.Turn == color
}

func (s *MatchScreen) isPromotion(request messages.MoveRequest) bool {
	piece := s.match.Pieces[request.FromX][request.FromY]
	if piece == nil || piece.Kind != chess.Pawn {
		return false
	}
	return (piece.Color == chess.White && request.ToY == 7) || (piece.Color == chess.Black && request.ToY == 0)
}

func (s *MatchScreen) checkSquare() *components.Square {
	if !s.match.InCheck {
		return nil
	}
	for x, column := range s.match.Pieces {
		for y, piece := range column {
			if piece != nil && piece.Kind == chess.King && piece.Color == s.match.Turn {
				return &components.Square{X: x, Y: y}
			}
		}
	}
	return nil
}

func (s *MatchScreen) layout() components.BoardLayout {
	side := min(s.size.Y-2*margin, s.size.X-3*margin-sidebarWidth)
	color, _ := s.color()
	return components.BoardLayout{
		Bounds:  rl.Rectangle{X: margin, Y: margin, Width: side, Height: side},
		Flipped: color == chess.Black,
	}
}

func (s *MatchScreen) Render() error {
	layout := s.layout()
	s.board.Render(components.BoardView{
		Layout:   layout,
		Pieces:   s.match.Pieces,
		Selected: s.selected,
		Check:    s.checkSquare(),
	})

	s.renderSidebar(layout.Bounds.X + layout.Bounds.Width + margin)

	if s.promotion != nil {
		s.renderPromotion(layout.Bounds)
	}
	return nil
}

func (s *MatchScreen) renderSidebar(x float32) {
	y := int32(margin)
	line := func(text string, color rl.Color) {
		for _, wrapped := range wrapText(text, textSize, int32(sidebarWidth)) {
			rl.DrawText(wrapped, int32(x), y, textSize, color)
			y += lineHeight
		}
	}

	line("Sala: "+s.match.Room, rl.DarkGray)
	y += lineHeight / 2
	line("Brancas: "+s.playerLabel(s.match.White), rl.DarkGray)
	line("Pretas: "+s.playerLabel(s.match.Black), rl.DarkGray)
	y += lineHeight / 2

	if s.finished() {
		line(s.outcomeText(), rl.DarkBlue)
	} else {
		turn := "Vez das " + s.match.Turn.String()
		if s.canMove() || (s.waiting && s.isMyTurn()) {
			turn += " (sua)"
		}
		line(turn, rl.DarkGray)
		if s.match.InCheck {
			line("Xeque!", rl.Red)
		}
	}
	if s.waiting {
		line("Aguardando o servidor...", rl.Gray)
	}
	if s.errorMessage != "" {
		line(s.errorMessage, rl.Red)
	}

	if len(s.spectators) > 0 {
		y += lineHeight / 2
		line("Espectadores:", rl.DarkGray)
		for _, name := range s.spectators {
			line("  "+name, rl.Gray)
		}
	}

	s.renderButtons(x)
}

func (s *MatchScreen) renderButtons(x float32) {
	button := func(row int, text string) bool {
		bounds := rl.Rectangle{
			X:      x,
			Y:      s.size.Y - margin - float32(row+1)*50,
			Width:  sidebarWidth,
			Height: 40,
		}
		return rg.Button(bounds, text)
	}

	_, isPlayer := s.color()
	switch {
	case isPlayer && !s.finished() && s.confirmResign:
		if button(1, "Confirmar desistência") {
			s.action = actionConfirmResign
		}
		if button(0, "Cancelar") {
			s.action = actionCancelResign
		}
	case isPlayer && !s.finished():
		if button(0, "Desistir") && !s.waiting {
			s.action = actionResign
		}
	default:
		if button(0, "Sair") {
			s.action = actionLeave
		}
	}
}

func (s *MatchScreen) renderPromotion(board rl.Rectangle) {
	rl.DrawRectangleRec(board, rl.NewColor(0, 0, 0, 140))

	width, height := float32(160), float32(40)
	x := board.X + (board.Width-width)/2
	y := board.Y + (board.Height-float32(len(promotionChoices)+2)*(height+10))/2

	title := "Promover para:"
	rl.DrawText(title, int32(board.X+(board.Width-float32(rl.MeasureText(title, textSize)))/2), int32(y), textSize, rl.White)
	y += height

	for _, choice := range promotionChoices {
		if rg.Button(rl.Rectangle{X: x, Y: y, Width: width, Height: height}, choice.label) {
			s.action = actionPromote
			s.promotionKind = choice.kind
		}
		y += height + 10
	}
	if rg.Button(rl.Rectangle{X: x, Y: y, Width: width, Height: height}, "Cancelar") {
		s.action = actionCancelPromotion
	}
}

func (s *MatchScreen) isMyTurn() bool {
	color, isPlayer := s.color()
	return isPlayer && s.match.Turn == color
}

func (s *MatchScreen) playerLabel(name string) string {
	if name == s.user {
		return name + " (você)"
	}
	return name
}

func (s *MatchScreen) outcomeText() string {
	switch s.match.Outcome {
	case chess.OutcomeCheckmate:
		return fmt.Sprintf("Xeque-mate! Vencedor: %s", s.match.Winner)
	case chess.OutcomeStalemate:
		return "Afogamento. Empate."
	case chess.OutcomeResignation:
		return fmt.Sprintf("Desistência. Vencedor: %s", s.match.Winner)
	}
	return string(s.match.Outcome)
}

func wrapText(text string, size, width int32) []string {
	var lines []string
	current := ""
	for word := range strings.FieldsSeq(text) {
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}
		if current != "" && rl.MeasureText(candidate, size) > width {
			lines = append(lines, current)
			current = word
			continue
		}
		current = candidate
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
