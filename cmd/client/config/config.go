package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const fileName = "client.toml"

type Config struct {
	ServerAddress string `toml:"server_address"`
}

func Default() Config {
	return Config{ServerAddress: "localhost:9000"}
}

func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return fileName
	}
	return filepath.Join(dir, "xadrez", fileName)
}

func Load(path string) (Config, error) {
	config := Default()
	if _, err := toml.DecodeFile(path, &config); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Default(), nil
		}
		return Default(), fmt.Errorf("erro ao ler a configuração %s: %w", path, err)
	}
	if err := config.Validate(); err != nil {
		return Default(), fmt.Errorf("configuração inválida em %s: %w", path, err)
	}
	return config, nil
}

func (c Config) Validate() error {
	host, port, err := net.SplitHostPort(c.ServerAddress)
	if err != nil || host == "" || port == "" {
		return errors.New("o endereço do servidor deve estar no formato host:porta")
	}
	return nil
}

func Save(path string, config Config) error {
	if err := config.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("erro ao criar o diretório da configuração: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("erro ao gravar a configuração: %w", err)
	}
	defer file.Close()
	if err := toml.NewEncoder(file).Encode(config); err != nil {
		return fmt.Errorf("erro ao gravar a configuração: %w", err)
	}
	return file.Close()
}
