package main

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type settings struct {
	Host             string `envconfig:"HOST" default:"0.0.0.0"`
	Port             int    `envconfig:"PORT" default:"8080"`
	LogLevel         string `envconfig:"LOG_LEVEL" default:"INFO"`
	UseMemoryStore   bool   `envconfig:"USE_MEMORY_STORE" default:"true"`
	PostgresHost     string `envconfig:"POSTGRES_HOST" default:"localhost"`
	PostgresPort     int    `envconfig:"POSTGRES_PORT" default:"5432"`
	PostgresUser     string `envconfig:"POSTGRES_USER" default:"postgres"`
	PostgresPassword string `envconfig:"POSTGRES_PASSWORD" default:"postgres"`
	PostgresDbName   string `envconfig:"POSTGRES_DB" default:"postgres"`
	AdminPassword    string `envconfig:"ADMIN_PASSWORD" default:"admin"`
}

func (s *settings) PostgresDbUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", s.PostgresUser, s.PostgresPassword, s.PostgresHost, s.PostgresPort, s.PostgresDbName)
}

var Settings settings

func InitSettings() {
	if err := envconfig.Process("", &Settings); err != nil {
		panic("failed to load settings from env: " + err.Error())
	}
}
