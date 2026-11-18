package infra

import (
	"context"
	"errors"
	"testing"

	"xadrez/ent"
	"xadrez/internal/domain"
)

func newTestClient(t *testing.T) *ent.Client {
	t.Helper()
	client, err := OpenDatabase(context.Background(), "file::memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("falha ao abrir o banco em memória: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

func createUsers(t *testing.T, client *ent.Client, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, err := client.User.Create().SetName(name).Save(context.Background()); err != nil {
			t.Fatalf("falha ao criar o usuário %s: %v", name, err)
		}
	}
}

func requireDomainError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("esperado erro de domínio, obtido nil")
	}
	if _, ok := errors.AsType[*domain.Error](err); !ok {
		t.Fatalf("esperado erro de domínio, obtido %T: %v", err, err)
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
}
