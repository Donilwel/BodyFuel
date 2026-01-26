package minio

import "time"

type Config struct {
	Endpoint   string        `yaml:"endpoint" env:"MINIO_ENDPOINT"`
	AccessKey  string        `yaml:"access_key" env:"MINIO_ACCESS_KEY"`
	SecretKey  string        `yaml:"secret_key" env:"MINIO_SECRET_KEY"`
	Bucket     string        `yaml:"bucket" env:"MINIO_BUCKET"`
	Region     string        `yaml:"region" env:"MINIO_REGION"`
	PublicURL  string        `yaml:"public_url" env:"MINIO_PUBLIC_URL"`
	PresignTTL time.Duration `yaml:"presign_ttl" env:"MINIO_PRESIGN_TTL"`
}
