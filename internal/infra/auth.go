package infra

import (
	"context"
	"fmt"
	"uuid"

	"xadrez/ent"
	"xadrez/ent/user"

	"charm.land/log/v2"
)

type AuthService struct {
	client *ent.Client
}

func NewAuthService(client *ent.Client) *AuthService {
	return &AuthService{
		client: client,
	}
}

func (s *AuthService) Login(ctx context.Context, name, password string) (string, error) {
	existing, err := s.client.User.Query().
		Where(user.Name(name)).
		WithPassword().
		Only(ctx)
	if ent.IsNotFound(err) {
		return s.register(ctx, name, password)
	}
	if err != nil {
		return "", fmt.Errorf("falha ao buscar o usuário: %w", err)
	}

	if existing.Edges.Password == nil {
		return "", fmt.Errorf("o usuário não possui senha cadastrada")
	}
	ok, err := verifyPassword(password, existing.Edges.Password.Hash)
	if err != nil {
		return "", fmt.Errorf("falha ao verificar a senha: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("a senha do usuário está incorreta")
	}

	log.Info("usuário autenticado com sucesso", "usuário", name)
	return generateToken(), nil
}

func (s *AuthService) register(ctx context.Context, name, password string) (string, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return "", err
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("falha ao iniciar a transação: %w", err)
	}

	created, err := tx.User.Create().SetName(name).Save(ctx)
	if err != nil {
		return "", rollback(tx, fmt.Errorf("falha ao criar o usuário: %w", err))
	}
	if _, err := tx.UserPassword.Create().SetUser(created).SetHash(hash).Save(ctx); err != nil {
		return "", rollback(tx, fmt.Errorf("falha ao salvar a senha do usuário: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("falha ao confirmar a transação: %w", err)
	}

	log.Info("novo usuário registrado", "usuário", name)
	return generateToken(), nil
}

func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		return fmt.Errorf("%w: falha no rollback: %v", err, rerr)
	}
	return err
}

func generateToken() string {
	return uuid.NewV7().String()
}
