package infra

import (
	"context"
	"testing"

	"xadrez/ent"
	"xadrez/ent/match"
	"xadrez/internal/domain/chess"
)

var foolsMate = [][4]int{
	{5, 1, 5, 2},
	{4, 6, 4, 4},
	{6, 1, 6, 3},
	{3, 7, 7, 3},
}

var fastestStalemate = [][4]int{
	{4, 1, 4, 2}, {0, 6, 0, 4},
	{3, 0, 7, 4}, {0, 7, 0, 5},
	{7, 4, 0, 4}, {7, 6, 7, 4},
	{7, 1, 7, 3}, {0, 5, 7, 5},
	{0, 4, 2, 6}, {5, 6, 5, 5},
	{2, 6, 3, 6}, {4, 7, 5, 6},
	{3, 6, 1, 6}, {3, 7, 3, 2},
	{1, 6, 1, 7}, {3, 2, 7, 6},
	{1, 7, 2, 7}, {5, 6, 6, 5},
	{2, 7, 4, 5},
}

type matchFixture struct {
	client  *ent.Client
	rooms   *RoomService
	matches *MatchService
}

func newMatchFixture(t *testing.T) matchFixture {
	t.Helper()
	ctx := context.Background()
	client := newTestClient(t)
	createUsers(t, client, "ana", "bob", "carol", "dave")
	roomService := NewRoomService(client)
	_, err := roomService.CreateRoom(ctx, "sala", "ana")
	requireNoError(t, err)
	for _, name := range []string{"bob", "carol"} {
		_, err := roomService.JoinRoom(ctx, "sala", name)
		requireNoError(t, err)
	}
	return matchFixture{client: client, rooms: roomService, matches: NewMatchService(client, roomService)}
}

func (f matchFixture) start(t *testing.T) {
	t.Helper()
	_, err := f.matches.StartMatch(context.Background(), "sala", "ana", "ana", "bob")
	requireNoError(t, err)
}

func (f matchFixture) play(t *testing.T, moves [][4]int) chess.Match {
	t.Helper()
	var snapshot chess.Match
	for i, m := range moves {
		player := "ana"
		if i%2 == 1 {
			player = "bob"
		}
		var err error
		snapshot, err = f.matches.MakeMove(context.Background(), "sala", player, m[0], m[1], m[2], m[3], chess.Queen)
		if err != nil {
			t.Fatalf("lance %d (%v) recusado: %v", i+1, m, err)
		}
	}
	return snapshot
}

func (f matchFixture) record(t *testing.T) *ent.Match {
	t.Helper()
	record, err := f.client.Match.Query().Order(ent.Desc(match.FieldCreatedAt)).First(context.Background())
	requireNoError(t, err)
	return record
}

func requireFinished(t *testing.T, record *ent.Match, result match.Result, termination match.Termination) {
	t.Helper()
	if record.FinishedAt == nil {
		t.Fatal("a partida deveria estar encerrada no banco")
	}
	if record.Result == nil || *record.Result != result {
		t.Fatalf("resultado esperado %s, obtido %v", result, record.Result)
	}
	if record.Termination == nil || *record.Termination != termination {
		t.Fatalf("terminação esperada %s, obtida %v", termination, record.Termination)
	}
}

func TestStartMatchRules(t *testing.T) {
	ctx := context.Background()
	f := newMatchFixture(t)

	_, err := f.matches.StartMatch(ctx, "sala", "bob", "ana", "bob")
	requireDomainError(t, err)
	_, err = f.matches.StartMatch(ctx, "sala", "ana", "ana", "ana")
	requireDomainError(t, err)
	_, err = f.matches.StartMatch(ctx, "sala", "ana", "ana", "dave")
	requireDomainError(t, err)
	_, err = f.matches.StartMatch(ctx, "inexistente", "ana", "ana", "bob")
	requireDomainError(t, err)

	snapshot, err := f.matches.StartMatch(ctx, "sala", "ana", "ana", "bob")
	requireNoError(t, err)
	if snapshot.White != "ana" || snapshot.Black != "bob" || snapshot.Turn != chess.White || snapshot.Outcome != chess.OutcomeNone {
		t.Fatalf("partida iniciada em estado inesperado: %+v", snapshot)
	}

	_, err = f.matches.StartMatch(ctx, "sala", "ana", "bob", "carol")
	requireDomainError(t, err)
}

func TestMakeMoveRules(t *testing.T) {
	ctx := context.Background()
	f := newMatchFixture(t)

	_, err := f.matches.MakeMove(ctx, "sala", "ana", 4, 1, 4, 3, chess.Queen)
	requireDomainError(t, err)

	f.start(t)
	_, err = f.matches.MakeMove(ctx, "sala", "carol", 4, 1, 4, 3, chess.Queen)
	requireDomainError(t, err)
	_, err = f.matches.MakeMove(ctx, "sala", "bob", 4, 6, 4, 4, chess.Queen)
	requireDomainError(t, err)
	_, err = f.matches.MakeMove(ctx, "sala", "ana", 4, 1, 4, 5, chess.Queen)
	requireDomainError(t, err)

	snapshot, err := f.matches.MakeMove(ctx, "sala", "ana", 4, 1, 4, 3, chess.Queen)
	requireNoError(t, err)
	if snapshot.Turn != chess.Black {
		t.Fatalf("a vez deveria passar às pretas, está com %v", snapshot.Turn)
	}
	if p := snapshot.Board[4][3]; p == nil || p.GetKind() != chess.Pawn || p.GetColor() != chess.White {
		t.Fatal("o peão branco deveria estar em e4")
	}
}

func TestMatchReloadedFromHistory(t *testing.T) {
	ctx := context.Background()
	f := newMatchFixture(t)
	f.start(t)
	played := f.play(t, foolsMate[:3])

	reloaded, err := NewMatchService(f.client, f.rooms).GetMatch(ctx, "sala")
	requireNoError(t, err)
	if !sameBoard(reloaded.Board, played.Board) {
		t.Fatal("o tabuleiro reconstruído do histórico difere do jogado")
	}
	if reloaded.Turn != chess.Black {
		t.Fatalf("a vez deveria ser das pretas, está com %v", reloaded.Turn)
	}

	moves, err := f.record(t).QueryMoves().Count(ctx)
	requireNoError(t, err)
	if moves != 3 {
		t.Fatalf("esperados 3 lances gravados, há %d", moves)
	}
}

type boardPiece interface {
	comparable
	GetKind() chess.PieceKind
	GetColor() chess.PieceColor
}

func sameBoard[P boardPiece](a, b [8][8]P) bool {
	var empty P
	for x := range a {
		for y := range a[x] {
			if (a[x][y] == empty) != (b[x][y] == empty) {
				return false
			}
			if a[x][y] != empty && (a[x][y].GetKind() != b[x][y].GetKind() || a[x][y].GetColor() != b[x][y].GetColor()) {
				return false
			}
		}
	}
	return true
}

func TestCheckmateFinishesMatch(t *testing.T) {
	ctx := context.Background()
	f := newMatchFixture(t)
	f.start(t)

	snapshot := f.play(t, foolsMate)
	if snapshot.Outcome != chess.OutcomeCheckmate || snapshot.Winner != "bob" {
		t.Fatalf("esperado xeque-mate das pretas, obtido %s vencedor %q", snapshot.Outcome, snapshot.Winner)
	}
	requireFinished(t, f.record(t), match.ResultBlack, match.TerminationCheckmate)

	_, err := f.matches.GetMatch(ctx, "sala")
	requireDomainError(t, err)
	_, err = f.matches.MakeMove(ctx, "sala", "ana", 0, 1, 0, 2, chess.Queen)
	requireDomainError(t, err)

	_, err = f.matches.StartMatch(ctx, "sala", "ana", "bob", "carol")
	requireNoError(t, err)
}

func TestStalemateIsDraw(t *testing.T) {
	f := newMatchFixture(t)
	f.start(t)

	snapshot := f.play(t, fastestStalemate)
	if snapshot.Outcome != chess.OutcomeStalemate || snapshot.Winner != "" {
		t.Fatalf("esperado afogamento sem vencedor, obtido %s vencedor %q", snapshot.Outcome, snapshot.Winner)
	}
	requireFinished(t, f.record(t), match.ResultDraw, match.TerminationStalemate)
}

func TestResignGivesVictoryToOpponent(t *testing.T) {
	ctx := context.Background()
	f := newMatchFixture(t)
	f.start(t)

	_, err := f.matches.Resign(ctx, "sala", "carol")
	requireDomainError(t, err)

	snapshot, err := f.matches.Resign(ctx, "sala", "bob")
	requireNoError(t, err)
	if snapshot.Outcome != chess.OutcomeResignation || snapshot.Winner != "ana" {
		t.Fatalf("esperada desistência com vitória da ana, obtido %s vencedor %q", snapshot.Outcome, snapshot.Winner)
	}
	requireFinished(t, f.record(t), match.ResultWhite, match.TerminationResignation)

	_, err = f.matches.Resign(ctx, "sala", "ana")
	requireDomainError(t, err)
}

func TestAbandonGivesVictoryToOpponent(t *testing.T) {
	ctx := context.Background()
	f := newMatchFixture(t)

	finished, err := f.matches.Abandon(ctx, "ana")
	requireNoError(t, err)
	if len(finished) != 0 {
		t.Fatalf("sem partida em andamento nada deveria ser encerrado: %v", finished)
	}

	f.start(t)
	f.play(t, foolsMate[:1])
	finished, err = f.matches.Abandon(ctx, "ana")
	requireNoError(t, err)
	if len(finished) != 1 || finished[0].Outcome != chess.OutcomeDisconnection || finished[0].Winner != "bob" {
		t.Fatalf("esperada vitória do bob por desconexão, obtido %+v", finished)
	}
	requireFinished(t, f.record(t), match.ResultBlack, match.TerminationDisconnection)

	finished, err = f.matches.Abandon(ctx, "carol")
	requireNoError(t, err)
	if len(finished) != 0 {
		t.Fatalf("espectador não tem partida para abandonar: %v", finished)
	}
}

func TestLeaveWithoutMatch(t *testing.T) {
	f := newMatchFixture(t)
	_, ended, err := f.matches.Leave(context.Background(), "sala", "bob", false)
	requireNoError(t, err)
	if ended {
		t.Fatal("sem partida em andamento nada deveria ser encerrado")
	}
}

func TestPlayerLeavingResigns(t *testing.T) {
	f := newMatchFixture(t)
	f.start(t)

	snapshot, ended, err := f.matches.Leave(context.Background(), "sala", "bob", false)
	requireNoError(t, err)
	if !ended || snapshot.Outcome != chess.OutcomeResignation || snapshot.Winner != "ana" {
		t.Fatalf("esperada vitória da ana por desistência, obtido ended=%v %+v", ended, snapshot)
	}
	requireFinished(t, f.record(t), match.ResultWhite, match.TerminationResignation)
}

func TestSpectatorLeavingKeepsMatch(t *testing.T) {
	ctx := context.Background()
	f := newMatchFixture(t)
	f.start(t)

	_, ended, err := f.matches.Leave(ctx, "sala", "carol", false)
	requireNoError(t, err)
	if ended {
		t.Fatal("a saída de um espectador não deveria encerrar a partida")
	}
	_, err = f.matches.GetMatch(ctx, "sala")
	requireNoError(t, err)
}

func TestClosingRoomEndsOthersMatch(t *testing.T) {
	ctx := context.Background()
	f := newMatchFixture(t)
	_, err := f.matches.StartMatch(ctx, "sala", "ana", "bob", "carol")
	requireNoError(t, err)

	snapshot, ended, err := f.matches.Leave(ctx, "sala", "ana", true)
	requireNoError(t, err)
	if !ended || snapshot.Outcome != chess.OutcomeRoomClosed || snapshot.Winner != "" {
		t.Fatalf("esperado encerramento sem vencedor, obtido ended=%v %+v", ended, snapshot)
	}
	record := f.record(t)
	if record.FinishedAt == nil || record.Result != nil || record.Termination == nil || *record.Termination != match.TerminationRoomClosed {
		t.Fatalf("registro inesperado: fim %v resultado %v terminação %v", record.FinishedAt, record.Result, record.Termination)
	}
}
