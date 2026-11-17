package infra

import (
	"strings"
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := hashPassword("segredo")
	requireNoError(t, err)

	ok, err := verifyPassword("segredo", hash)
	requireNoError(t, err)
	if !ok {
		t.Fatal("a senha correta deveria ser aceita")
	}

	ok, err = verifyPassword("outra", hash)
	requireNoError(t, err)
	if ok {
		t.Fatal("a senha errada deveria ser recusada")
	}
}

func TestHashUsesRandomSaltAndHidesPassword(t *testing.T) {
	first, err := hashPassword("segredo")
	requireNoError(t, err)
	second, err := hashPassword("segredo")
	requireNoError(t, err)

	if first == second {
		t.Fatal("dois hashes da mesma senha deveriam ser diferentes")
	}
	if !strings.HasPrefix(first, "$argon2id$") {
		t.Fatalf("hash fora do formato argon2id: %s", first)
	}
	if strings.Contains(first, "segredo") {
		t.Fatal("o hash não pode conter a senha")
	}
}

func TestVerifyRejectsMalformedHash(t *testing.T) {
	valid, err := hashPassword("segredo")
	requireNoError(t, err)
	parts := strings.Split(valid, "$")

	cases := map[string]string{
		"vazio":            "",
		"sem separadores":  "texto",
		"outro algoritmo":  strings.Replace(valid, "argon2id", "bcrypt", 1),
		"versão diferente": strings.Join([]string{"", parts[1], "v=18", parts[3], parts[4], parts[5]}, "$"),
		"parâmetros":       strings.Join([]string{"", parts[1], parts[2], "m=x", parts[4], parts[5]}, "$"),
		"salt inválido":    strings.Join([]string{"", parts[1], parts[2], parts[3], "!!!", parts[5]}, "$"),
		"chave inválida":   strings.Join([]string{"", parts[1], parts[2], parts[3], parts[4], "!!!"}, "$"),
	}
	for name, encoded := range cases {
		t.Run(name, func(t *testing.T) {
			if ok, err := verifyPassword("segredo", encoded); err == nil || ok {
				t.Fatalf("hash malformado deveria dar erro, obtido ok=%v err=%v", ok, err)
			}
		})
	}
}
