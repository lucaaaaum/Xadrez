package infra

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    uint32 = 2
	argon2Memory  uint32 = 19 * 1024
	argon2Threads uint8  = 1
	argon2KeyLen  uint32 = 32
	argon2SaltLen        = 16
)

var argon2Encoding = base64.RawStdEncoding

func hashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("falha ao gerar o salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory,
		argon2Time,
		argon2Threads,
		argon2Encoding.EncodeToString(salt),
		argon2Encoding.EncodeToString(key),
	), nil
}

func verifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("formato de hash inválido")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("versão do hash inválida: %w", err)
	}
	if version != argon2.Version {
		return false, fmt.Errorf("versão do argon2 incompatível: %d", version)
	}

	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, fmt.Errorf("parâmetros do hash inválidos: %w", err)
	}

	salt, err := argon2Encoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("salt do hash inválido: %w", err)
	}
	expected, err := argon2Encoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("chave do hash inválida: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expected)))
	return subtle.ConstantTimeCompare(key, expected) == 1, nil
}
