package screens

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"xadrez/cmd/client/components"
	"xadrez/internal/application/messages"
	"xadrez/internal/domain/chess"

	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type matchAction int

const (
	actionNone matchAction = iota
	actionResign
	actionConfirmResign
	actionCancelResign
	actionLeave
	actionPromote
	actionCancelPromotion
	actionAcceptInvite
)

const (
	margin       float32 = 20
	sidebarWidth float32 = 200
	lineHeight   int32   = 22
	textSize     int32   = 18

	sidebarButtonHeight float32 = 40
	sidebarButtonStep   float32 = 50
	maxLastMoveSquares          = 4
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
	session *Session
	room    messages.RoomResponse
	next    components.Screen

	match         messages.MatchResponse
	spectators    []string
	selected      *components.Square
	moves         []components.Square
	lastMove      []components.Square
	promotion     *messages.MoveRequest
	waiting       bool
	confirmResign bool
	errorMessage  string

	refreshRoom   bool
	action        matchAction
	promotionKind chess.PieceKind
}

func NewMatchScreen(session *Session, room messages.RoomResponse, match messages.MatchResponse) *MatchScreen {
	s := &MatchScreen{
		session:     session,
		room:        room,
		match:       match,
		refreshRoom: true,
	}
	s.next = s
	s.updateSpectators()
	return s
}

func (s *MatchScreen) HandleEvent(event components.Event) {
	if message, ok := errorMessage(event); ok {
		s.errorMessage = message
		s.waiting = false
		return
	}
	switch event.Type {
	case "match_updated", "make_move_ok", "resign_ok", "get_match_ok", "start_match_ok":
		if match, ok := event.Payload.(messages.MatchResponse); ok && match.Room == s.match.Room {
			s.applyMatch(match)
			return
		}
		if s.playing() {
			return
		}
	case "list_rooms_ok":
		if response, ok := event.Payload.(messages.ListRoomsResponse); ok {
			for _, room := range response.Rooms {
				if room.Room == s.room.Room {
					s.room = room
					s.updateSpectators()
				}
			}
		}
		return
	}
	if next, ok := s.session.handleCommonEvent(event); ok && next != nil {
		s.next = next
	}
}

func (s *MatchScreen) applyMatch(match messages.MatchResponse) {
	if changed := changedSquares(s.match.Pieces, match.Pieces); len(changed) > 0 {
		s.lastMove = nil
		if len(changed) <= maxLastMoveSquares {
			s.lastMove = changed
		}
	}
	s.match = match
	s.waiting = false
	s.deselect()
	s.promotion = nil
	if s.finished() {
		s.confirmResign = false
	}
	s.updateSpectators()
}

func (s *MatchScreen) updateSpectators() {
	s.spectators = slices.DeleteFunc(slices.Clone(s.room.Participants), func(name string) bool {
		return name == s.match.White || name == s.match.Black
	})
}

func (s *MatchScreen) Update(ctx context.Context) (components.Screen, error) {
	if s.refreshRoom {
		s.refreshRoom = false
		if err := s.session.client.ListRooms(); err != nil {
			s.errorMessage = err.Error()
		}
	}
	s.handleAction()
	s.handleBoardClick()
	return s.next, nil
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
		s.deselect()
		s.promotion = nil
		s.send(func() error {
			return s.session.client.Resign(messages.ResignRequest{Room: s.match.Room})
		})
	case actionLeave:
		s.next = NewRoomScreen(s.session, s.room)
	case actionPromote:
		if s.promotion != nil {
			request := *s.promotion
			request.Promotion = s.promotionKind
			s.promotion = nil
			s.sendMove(request)
		}
	case actionCancelPromotion:
		s.promotion = nil
	case actionAcceptInvite:
		if err := s.session.acceptInvite(); err != nil {
			s.errorMessage = err.Error()
		}
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
			s.deselect()
		} else {
			s.selectSquare(square)
		}
		return
	}
	if s.selected == nil {
		return
	}
	if !slices.Contains(s.moves, square) {
		s.deselect()
		return
	}

	request := messages.MoveRequest{
		Room:  s.match.Room,
		FromX: s.selected.X,
		FromY: s.selected.Y,
		ToX:   square.X,
		ToY:   square.Y,
	}
	s.deselect()
	if s.isPromotion(request) {
		s.promotion = &request
		return
	}
	s.sendMove(request)
}

func (s *MatchScreen) selectSquare(square components.Square) {
	s.selected = &square
	s.moves = chess.LegalMoves(s.position(), square.X, square.Y)
}

func (s *MatchScreen) deselect() {
	s.selected = nil
	s.moves = nil
}

func (s *MatchScreen) position() [8][8]*chess.PlacedPiece {
	var position [8][8]*chess.PlacedPiece
	for x, column := range s.match.Pieces {
		for y, piece := range column {
			if piece != nil {
				position[x][y] = &chess.PlacedPiece{Kind: piece.Kind, Color: piece.Color}
			}
		}
	}
	return position
}

func (s *MatchScreen) sendMove(request messages.MoveRequest) {
	s.send(func() error {
		return s.session.client.MakeMove(request)
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
	switch s.session.user {
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

func (s *MatchScreen) playing() bool {
	_, isPlayer := s.color()
	return isPlayer && !s.finished()
}

func (s *MatchScreen) canMove() bool {
	return s.playing() && !s.waiting && s.isMyTurn()
}

func (s *MatchScreen) isMyTurn() bool {
	color, isPlayer := s.color()
	return isPlayer && s.match.Turn == color
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
	size := s.session.size
	side := min(size.Y-2*margin, size.X-3*margin-sidebarWidth)
	color, _ := s.color()
	return components.BoardLayout{
		Bounds:  rl.Rectangle{X: margin, Y: margin, Width: side, Height: side},
		Flipped: color == chess.Black,
	}
}

func (s *MatchScreen) Render() error {
	layout := s.layout()
	s.session.Board().Render(components.BoardView{
		Layout:   layout,
		Pieces:   s.match.Pieces,
		Selected: s.selected,
		Check:    s.checkSquare(),
		LastMove: s.lastMove,
		Moves:    s.moves,
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
		if s.isMyTurn() {
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

	if !s.playing() {
		bottom := s.session.size.Y - margin - sidebarButtonStep - spacing
		if s.session.renderInvitePanel(x, bottom, sidebarWidth) {
			s.action = actionAcceptInvite
		}
	}
}

func (s *MatchScreen) renderButtons(x float32) {
	button := func(row int, text string) bool {
		bounds := rl.Rectangle{
			X:      x,
			Y:      s.session.size.Y - margin - float32(row)*sidebarButtonStep - sidebarButtonHeight,
			Width:  sidebarWidth,
			Height: sidebarButtonHeight,
		}
		return rg.Button(bounds, text)
	}

	switch {
	case s.playing() && s.confirmResign:
		if button(1, "Confirmar desistência") {
			s.action = actionConfirmResign
		}
		if button(0, "Cancelar") {
			s.action = actionCancelResign
		}
	case s.playing():
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

func (s *MatchScreen) playerLabel(name string) string {
	if name == s.session.user {
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

func changedSquares(before, after [8][8]*messages.Piece) []components.Square {
	var changed []components.Square
	for x := range 8 {
		for y := range 8 {
			if !samePiece(before[x][y], after[x][y]) {
				changed = append(changed, components.Square{X: x, Y: y})
			}
		}
	}
	return changed
}

func samePiece(a, b *messages.Piece) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
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
