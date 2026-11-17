package infra

import (
	"context"
	"slices"
	"testing"

	"xadrez/internal/domain/rooms"
)

func newRoomFixture(t *testing.T) *RoomService {
	t.Helper()
	client := newTestClient(t)
	createUsers(t, client, "ana", "bob", "carol")
	service := NewRoomService(client)
	_, err := service.CreateRoom(context.Background(), "sala", "ana")
	requireNoError(t, err)
	return service
}

func TestCreateRoom(t *testing.T) {
	service := newRoomFixture(t)

	room, err := service.GetRoom(context.Background(), "sala")
	requireNoError(t, err)
	if room.Admin != "ana" || !slices.Equal(room.Participants, []string{"ana"}) {
		t.Fatalf("o criador deveria ser administrador e único participante: %+v", room)
	}
}

func TestCreateRoomValidation(t *testing.T) {
	ctx := context.Background()
	service := newRoomFixture(t)

	_, err := service.CreateRoom(ctx, "sala", "bob")
	requireDomainError(t, err)
	_, err = service.CreateRoom(ctx, "Sala Grande", "bob")
	requireDomainError(t, err)
	_, err = service.CreateRoom(ctx, "outra", "ninguem")
	requireDomainError(t, err)
}

func TestJoinRoom(t *testing.T) {
	ctx := context.Background()
	service := newRoomFixture(t)

	room, err := service.JoinRoom(ctx, "sala", "bob")
	requireNoError(t, err)
	if !slices.Equal(room.Participants, []string{"ana", "bob"}) {
		t.Fatalf("participantes inesperados: %v", room.Participants)
	}

	_, err = service.JoinRoom(ctx, "sala", "bob")
	requireDomainError(t, err)
	_, err = service.JoinRoom(ctx, "inexistente", "bob")
	requireDomainError(t, err)
}

func TestListRoomsSortedByName(t *testing.T) {
	ctx := context.Background()
	service := newRoomFixture(t)
	_, err := service.CreateRoom(ctx, "alfa", "bob")
	requireNoError(t, err)

	list, err := service.ListRooms(ctx)
	requireNoError(t, err)
	names := make([]string, 0, len(list))
	for _, room := range list {
		names = append(names, room.Name)
	}
	if !slices.Equal(names, []string{"alfa", "sala"}) {
		t.Fatalf("salas fora de ordem: %v", names)
	}
}

func TestInviteToRoomRules(t *testing.T) {
	ctx := context.Background()
	service := newRoomFixture(t)

	_, err := service.InviteToRoom(ctx, "sala", "bob", "carol")
	requireDomainError(t, err)
	_, err = service.InviteToRoom(ctx, "sala", "ana", "ana")
	requireDomainError(t, err)
	_, err = service.InviteToRoom(ctx, "sala", "ana", "ninguem")
	requireDomainError(t, err)
	_, err = service.InviteToRoom(ctx, "inexistente", "ana", "bob")
	requireDomainError(t, err)

	invite, err := service.InviteToRoom(ctx, "sala", "ana", "bob")
	requireNoError(t, err)
	if invite != (rooms.Invite{RoomName: "sala", From: "ana", To: "bob"}) {
		t.Fatalf("convite inesperado: %+v", invite)
	}
	_, err = service.InviteToRoom(ctx, "sala", "ana", "bob")
	requireDomainError(t, err)
}

func TestDeclineInvite(t *testing.T) {
	ctx := context.Background()
	service := newRoomFixture(t)

	_, err := service.DeclineInvite(ctx, "sala", "bob")
	requireDomainError(t, err)

	_, err = service.InviteToRoom(ctx, "sala", "ana", "bob")
	requireNoError(t, err)
	invite, err := service.DeclineInvite(ctx, "sala", "bob")
	requireNoError(t, err)
	if invite.From != "ana" {
		t.Fatalf("a recusa deveria identificar o administrador: %+v", invite)
	}

	_, err = service.DeclineInvite(ctx, "sala", "bob")
	requireDomainError(t, err)
	_, err = service.InviteToRoom(ctx, "sala", "ana", "bob")
	requireNoError(t, err)
}

func TestJoinRoomConsumesInvite(t *testing.T) {
	ctx := context.Background()
	service := newRoomFixture(t)

	_, err := service.InviteToRoom(ctx, "sala", "ana", "bob")
	requireNoError(t, err)
	_, err = service.JoinRoom(ctx, "sala", "bob")
	requireNoError(t, err)

	_, err = service.DeclineInvite(ctx, "sala", "bob")
	requireDomainError(t, err)
	_, err = service.InviteToRoom(ctx, "sala", "ana", "bob")
	requireDomainError(t, err)
}
