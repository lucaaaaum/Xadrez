package infra

import (
	"context"
	"testing"

	"xadrez/ent/user"
)

func TestLoginRegistersNewUser(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)
	service := NewAuthService(client)

	token, err := service.Login(ctx, "ana", "segredo")
	requireNoError(t, err)
	if token == "" {
		t.Fatal("o login deveria devolver um token")
	}

	created, err := client.User.Query().Where(user.Name("ana")).WithPassword().Only(ctx)
	requireNoError(t, err)
	stored := created.Edges.Password.Hash
	if stored == "segredo" {
		t.Fatal("a senha não pode ser gravada em texto puro")
	}
	ok, err := verifyPassword("segredo", stored)
	requireNoError(t, err)
	if !ok {
		t.Fatal("o hash gravado deveria corresponder à senha")
	}
}

func TestLoginExistingUser(t *testing.T) {
	ctx := context.Background()
	service := NewAuthService(newTestClient(t))

	first, err := service.Login(ctx, "ana", "segredo")
	requireNoError(t, err)
	second, err := service.Login(ctx, "ana", "segredo")
	requireNoError(t, err)
	if first == second {
		t.Fatal("cada login deveria gerar um token novo")
	}

	if _, err := service.Login(ctx, "ana", "errada"); err == nil {
		t.Fatal("login com senha errada deveria falhar")
	}
}

func TestLoginRejectsInvalidName(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)
	service := NewAuthService(client)

	if _, err := service.Login(ctx, "Ana Maria", "segredo"); err == nil {
		t.Fatal("nome inválido deveria ser recusado")
	}
	count, err := client.User.Query().Count(ctx)
	requireNoError(t, err)
	if count != 0 {
		t.Fatalf("nenhum usuário deveria ter sido criado, há %d", count)
	}
}
