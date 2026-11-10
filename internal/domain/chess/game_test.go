package chess

import "testing"

type placement struct {
	kind  PieceKind
	color PieceColor
	x, y  int
}

func boardWith(placements ...placement) *board {
	b := &board{}
	for _, p := range placements {
		b.addPiece(newPiece(p.kind, p.color), p.x, p.y)
	}
	return b
}

func gameWith(turn PieceColor, placements ...placement) *Game {
	g := &Game{board: boardWith(placements...), currentTurn: turn}
	g.updateStatus()
	return g
}

func mustMove(t *testing.T, g *Game, fromX, fromY, toX, toY int) {
	t.Helper()
	if err := g.MakeMove(g.CurrentTurn(), fromX, fromY, toX, toY, Pawn); err != nil {
		t.Fatalf("lance (%d,%d)->(%d,%d) recusado: %v", fromX, fromY, toX, toY, err)
	}
}

func mustReject(t *testing.T, g *Game, fromX, fromY, toX, toY int) {
	t.Helper()
	if err := g.MakeMove(g.CurrentTurn(), fromX, fromY, toX, toY, Pawn); err == nil {
		t.Fatalf("lance (%d,%d)->(%d,%d) deveria ser recusado", fromX, fromY, toX, toY)
	}
}

func TestSlidingPiecesNeedClearPath(t *testing.T) {
	g := NewGame()
	mustReject(t, g, 0, 0, 0, 3)
	mustReject(t, g, 2, 0, 4, 2)
	mustReject(t, g, 3, 0, 3, 4)
	mustMove(t, g, 1, 0, 2, 2)
}

func TestPawnRules(t *testing.T) {
	g := NewGame()
	mustReject(t, g, 4, 1, 5, 2)
	mustMove(t, g, 4, 1, 4, 3)
	mustMove(t, g, 4, 6, 4, 4)
	mustReject(t, g, 4, 3, 4, 4)
	mustMove(t, g, 3, 1, 3, 3)
	mustMove(t, g, 4, 4, 3, 3)
}

func TestPawnDoubleStepBlocked(t *testing.T) {
	g := NewGame()
	mustMove(t, g, 6, 0, 5, 2)
	mustMove(t, g, 0, 6, 0, 5)
	mustReject(t, g, 5, 1, 5, 3)
}

func TestCannotMoveOpponentPieceOrOutOfTurn(t *testing.T) {
	g := NewGame()
	if err := g.MakeMove(White, 4, 6, 4, 4, Pawn); err == nil {
		t.Fatal("brancas moveram uma peça preta")
	}
	if err := g.MakeMove(Black, 4, 6, 4, 4, Pawn); err == nil {
		t.Fatal("pretas jogaram fora da vez")
	}
}

func TestPinnedPieceCannotExposeKing(t *testing.T) {
	g := gameWith(White,
		placement{King, White, 4, 0},
		placement{Rook, White, 4, 1},
		placement{Rook, Black, 4, 7},
		placement{King, Black, 0, 7},
	)
	mustReject(t, g, 4, 1, 0, 1)
	mustMove(t, g, 4, 1, 4, 5)
}

func TestKingCannotMoveIntoCheck(t *testing.T) {
	g := gameWith(White,
		placement{King, White, 4, 0},
		placement{Rook, Black, 3, 7},
		placement{King, Black, 7, 7},
	)
	mustReject(t, g, 4, 0, 3, 0)
	mustMove(t, g, 4, 0, 5, 0)
}

func TestFoolsMate(t *testing.T) {
	g := NewGame()
	mustMove(t, g, 5, 1, 5, 2)
	mustMove(t, g, 4, 6, 4, 4)
	mustMove(t, g, 6, 1, 6, 3)
	mustMove(t, g, 3, 7, 7, 3)
	if g.Status() != Checkmate {
		t.Fatalf("esperado xeque-mate, obtido %v", g.Status())
	}
	if !g.InCheck() {
		t.Fatal("as brancas deveriam estar em xeque")
	}
	mustReject(t, g, 0, 1, 0, 2)
}

func TestStalemate(t *testing.T) {
	g := gameWith(Black,
		placement{King, Black, 7, 7},
		placement{Queen, White, 6, 5},
		placement{King, White, 5, 6},
	)
	if g.Status() != Stalemate {
		t.Fatalf("esperado afogamento, obtido %v", g.Status())
	}
}

func TestCheckIsNotCheckmateWhenKingCanEscape(t *testing.T) {
	g := gameWith(White,
		placement{King, White, 4, 0},
		placement{Rook, White, 0, 1},
		placement{King, Black, 4, 7},
	)
	mustMove(t, g, 0, 1, 0, 7)
	if g.Status() != Ongoing || !g.InCheck() {
		t.Fatalf("esperado xeque sem mate, obtido status %v e xeque %v", g.Status(), g.InCheck())
	}
}

func TestPromotion(t *testing.T) {
	g := gameWith(White,
		placement{Pawn, White, 0, 6},
		placement{Pawn, White, 1, 6},
		placement{King, White, 4, 0},
		placement{King, Black, 7, 4},
	)
	if !g.IsPromotion(0, 6, 7) {
		t.Fatal("lance deveria ser promoção")
	}
	mustMove(t, g, 0, 6, 0, 7)
	if kind := g.GetBoard()[0][7].GetKind(); kind != Queen {
		t.Fatalf("promoção padrão deveria ser dama, obtido %v", kind)
	}
	mustMove(t, g, 7, 4, 7, 3)
	if err := g.MakeMove(White, 1, 6, 1, 7, Knight); err != nil {
		t.Fatal(err)
	}
	if kind := g.GetBoard()[1][7].GetKind(); kind != Knight {
		t.Fatalf("promoção deveria ser cavalo, obtido %v", kind)
	}
}

func TestPromotionToKingRejected(t *testing.T) {
	g := gameWith(White,
		placement{Pawn, White, 0, 6},
		placement{King, White, 4, 0},
		placement{King, Black, 7, 4},
	)
	if err := g.MakeMove(White, 0, 6, 0, 7, King); err == nil {
		t.Fatal("promoção a rei deveria ser recusada")
	}
}
