package postgres

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

type Config struct {
	Username        string        `yaml:"user" env:"USERNAME"`
	Password        string        `yaml:"password" env:"PASSWORD"`
	Database        string        `yaml:"database" env:"DATABASE"`
	Host            string        `yaml:"host" env:"HOST"`
	MaxOpenConns    int           `yaml:"max_open_conn"`
	MaxIdleConns    int           `yaml:"max_idle_conn"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}
