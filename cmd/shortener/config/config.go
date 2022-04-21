package config

import (
	"encoding/json"
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
	"os"
)

type EnvConfig struct {
	Address         string `env:"SERVER_ADDRESS" json:"server_address"`
	BaseURL         string `env:"BASE_URL" json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN     string `env:"DATABASE_DSN" json:"database_dsn"`
	ProfileCPUFile  string `env:"PROFILE_CPU" json:"profile_cpu_file"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	CertFile        string `env:"CERT" json:"cert_file"`
	CertKeyFile     string `env:"CERT_KEY" json:"cert_key_file"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
}

type ConfigFile struct {
	File string `env:"CONFIG"`
}

func Parse() EnvConfig {
	var err error
	cfg := EnvConfig{
		Address:         ":8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
		DatabaseDSN:     "",
		EnableHTTPS:     false,
		CertFile:        "./cert/certificate.crt",
		CertKeyFile:     "./cert/certificate.key",
	}

	argFlags := parseFlags()
	configFile := os.Getenv("CONFIG")
	if file, ok := argFlags["Config"]; ok {
		configFile = file
	}
	if configFile != "" {
		err = parseConfigFile(&cfg, configFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	applyArgsToConfig(&cfg, argFlags)

	return cfg
}

func parseConfigFile(cfg *EnvConfig, configFile string) error {
	f, err := os.Open(configFile)
	if err != nil {
		return err
	}

	d := json.NewDecoder(f)
	err = d.Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}

func parseFlags() map[string]string {
	configFileFlag := flag.String("c", "", "Файл конфига в формате JSON")
	addressFlag := flag.String("a", "", "Адрес запуска HTTP-сервера")
	baseURLFlag := flag.String("b", "", "Базовый адрес результирующего сокращённого URL")
	fileStoragePathFlag := flag.String("f", "", "Путь до файла с сокращёнными URL")
	databaseDSNFlag := flag.String("d", "", "Адрес подключения к БД")
	profileCPUFlag := flag.String("pcpu", "", "Файл для полайлинга cpu")
	enableHTTPSFlag := flag.Bool("s", false, "Включить HTTPS")
	certFile := flag.String("crypto-key", "", "Путь к файлу сертификата")
	certKeyFile := flag.String("k", "", "Путь к ключу сертификата")
	trustedSubnet := flag.String("t", "", "Доверенная подсеть")
	flag.Parse()

	cfg := make(map[string]string)
	if *configFileFlag != "" {
		cfg["Config"] = *configFileFlag
	}
	if *addressFlag != "" {
		cfg["Address"] = *addressFlag
	}
	if *baseURLFlag != "" {
		cfg["BaseURL"] = *baseURLFlag
	}
	if *fileStoragePathFlag != "" {
		cfg["FileStoragePath"] = *fileStoragePathFlag
	}
	if *databaseDSNFlag != "" {
		cfg["DatabaseDSN"] = *databaseDSNFlag
	}
	if *profileCPUFlag != "" {
		cfg["ProfileCPUFile"] = *profileCPUFlag
	}
	if *enableHTTPSFlag {
		cfg["EnableHTTPS"] = "on"
	}
	if *certFile != "" {
		cfg["CertFile"] = *certFile
	}
	if *certKeyFile != "" {
		cfg["CertKeyFile"] = *certKeyFile
	}
	if *trustedSubnet != "" {
		cfg["TrustedSubnet"] = *trustedSubnet
	}
	return cfg
}

func applyArgsToConfig(config *EnvConfig, args map[string]string) {
	if value, ok := args["Address"]; ok {
		config.Address = value
	}
	if value, ok := args["BaseURL"]; ok {
		config.BaseURL = value
	}
	if value, ok := args["FileStoragePath"]; ok {
		config.FileStoragePath = value
	}
	if value, ok := args["DatabaseDSN"]; ok {
		config.DatabaseDSN = value
	}
	if value, ok := args["ProfileCPUFile"]; ok {
		config.ProfileCPUFile = value
	}
	if value, ok := args["EnableHTTPS"]; ok {
		config.EnableHTTPS = value == "on"
	}
	if value, ok := args["CertFile"]; ok {
		config.CertFile = value
	}
	if value, ok := args["CertKeyFile"]; ok {
		config.CertKeyFile = value
	}
	if value, ok := args["TrustedSubnet"]; ok {
		config.TrustedSubnet = value
	}
}
