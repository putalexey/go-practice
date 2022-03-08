package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type EnvConfig struct {
	Address         string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	ProfileCPUFile  string `env:"PROFILE_CPU"`
}

func Parse() EnvConfig {
	cfg := EnvConfig{
		Address:         ":8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
		DatabaseDSN:     "",
	}
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	cfg.parseFlags()

	return cfg
}

func (cfg *EnvConfig) parseFlags() {
	addressFlag := flag.String("a", "", "Адрес запуска HTTP-сервера")
	baseURLFlag := flag.String("b", "", "Базовый адрес результирующего сокращённого URL")
	fileStoragePathFlag := flag.String("f", "", "Путь до файла с сокращёнными URL")
	databaseDSNFlag := flag.String("d", "", "Адрес подключения к БД")
	profileCPUFlag := flag.String("pcpu", "", "Файл для полайлинга cpu")
	flag.Parse()

	if *addressFlag != "" {
		cfg.Address = *addressFlag
	}
	if *baseURLFlag != "" {
		cfg.BaseURL = *baseURLFlag
	}
	if *fileStoragePathFlag != "" {
		cfg.FileStoragePath = *fileStoragePathFlag
	}
	if *databaseDSNFlag != "" {
		cfg.DatabaseDSN = *databaseDSNFlag
	}
	if *profileCPUFlag != "" {
		cfg.ProfileCPUFile = *profileCPUFlag
	}
}
