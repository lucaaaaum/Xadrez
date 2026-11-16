package config

import (
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsDefault(t *testing.T) {
	config, err := Load(filepath.Join(t.TempDir(), "inexistente.toml"))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if config != Default() {
		t.Fatalf("esperava a configuração padrão, veio %+v", config)
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "xadrez", "client.toml")
	want := Config{ServerAddress: "192.168.0.10:9001"}
	if err := Save(path, want); err != nil {
		t.Fatalf("erro ao salvar: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("erro ao carregar: %v", err)
	}
	if got != want {
		t.Fatalf("esperava %+v, veio %+v", want, got)
	}
}

func TestSaveRejectsInvalidAddress(t *testing.T) {
	for _, address := range []string{"", "localhost", ":9000"} {
		if err := Save(filepath.Join(t.TempDir(), "client.toml"), Config{ServerAddress: address}); err == nil {
			t.Fatalf("endereço %q deveria ser recusado", address)
		}
	}
}
