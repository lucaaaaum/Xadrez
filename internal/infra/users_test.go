package infra

import (
	"context"
	"errors"
	"slices"
	"testing"

	"xadrez/internal/domain/users"
)

func onlineNames(t *testing.T, service *UserService) []string {
	t.Helper()
	online, err := service.GetOnlineUsers(context.Background())
	requireNoError(t, err)
	names := make([]string, 0, len(online))
	for _, u := range online {
		names = append(names, u.Name)
	}
	return names
}

func TestRegisterLoginRequiresExistingUser(t *testing.T) {
	service := NewUserService(newTestClient(t))
	if err := service.RegisterUserLogin(context.Background(), users.User{Name: "ana"}); err == nil {
		t.Fatal("usuário inexistente não deveria ficar online")
	}
}

func TestRegisterLoginTwiceFails(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)
	createUsers(t, client, "ana")
	service := NewUserService(client)

	requireNoError(t, service.RegisterUserLogin(ctx, users.User{Name: "ana"}))
	err := service.RegisterUserLogin(ctx, users.User{Name: "ana"})
	if !errors.Is(err, users.ErrAlreadyOnline) {
		t.Fatalf("esperado ErrAlreadyOnline, obtido %v", err)
	}
}

func TestLogoutRemovesUser(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)
	createUsers(t, client, "ana")
	service := NewUserService(client)

	requireNoError(t, service.RegisterUserLogin(ctx, users.User{Name: "ana"}))
	requireNoError(t, service.RegisterUserLogout(ctx, users.User{Name: "ana"}))
	if names := onlineNames(t, service); len(names) != 0 {
		t.Fatalf("ninguém deveria estar online, há %v", names)
	}
	if err := service.RegisterUserLogout(ctx, users.User{Name: "ana"}); err == nil {
		t.Fatal("logout de quem não está online deveria falhar")
	}
	requireNoError(t, service.RegisterUserLogin(ctx, users.User{Name: "ana"}))
}

func TestOnlineUsersSortedByName(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)
	createUsers(t, client, "carol", "ana", "bob")
	service := NewUserService(client)

	for _, name := range []string{"carol", "ana", "bob"} {
		requireNoError(t, service.RegisterUserLogin(ctx, users.User{Name: name}))
	}
	if names := onlineNames(t, service); !slices.Equal(names, []string{"ana", "bob", "carol"}) {
		t.Fatalf("lista fora de ordem: %v", names)
	}
}
