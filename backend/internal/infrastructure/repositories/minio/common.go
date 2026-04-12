package minio

import "time"

type Config struct {
	Endpoint   string        `yaml:"endpoint" env:"ENDPOINT"`
	AccessKey  string        `yaml:"access_key" env:"ACCESS_KEY"`
	SecretKey  string        `yaml:"secret_key" env:"SECRET_KEY"`
	Bucket     string        `yaml:"bucket" env:"BUCKET"`
	Region     string        `yaml:"region" env:"REGION"`
	PublicURL  string        `yaml:"public_url" env:"PUBLIC_URL"`
	PresignTTL time.Duration `yaml:"presign_ttl" env:"PRESIGN_TTL"`
}
