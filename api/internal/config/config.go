package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL        string
	Port               string
	JWTSecret          string
	AllowedOrigins     []string
	LiquidsoapSocket   string
	EncryptionKey      string
	StorageDriver      string // "local" or "s3"
	StorageLocalRoot   string
	S3Endpoint         string
	S3Bucket           string
	S3Region           string
	S3AccessKey        string
	S3SecretKey        string
	IcecastURL         string
	IcecastAdminUser   string
	IcecastAdminPass   string
	IcecastMounts      []string
	GeoIPDatabasePath  string
	SMTPHost           string
	SMTPPort           int
	SMTPUsername       string
	SMTPPassword       string
	SMTPFrom           string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:             "8080",
		AllowedOrigins:   []string{"http://localhost:5173"},
		LiquidsoapSocket: "/var/run/liquidsoap/radio.sock",
	}
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if p := os.Getenv("PORT"); p != "" {
		cfg.Port = p
	}
	if o := os.Getenv("ALLOWED_ORIGINS"); o != "" {
		cfg.AllowedOrigins = strings.Split(o, ",")
	}
	if s := os.Getenv("LIQUIDSOAP_SOCKET"); s != "" {
		cfg.LiquidsoapSocket = s
	}
	cfg.EncryptionKey = os.Getenv("ENCRYPTION_KEY")

	cfg.StorageDriver = os.Getenv("STORAGE_DRIVER")
	if cfg.StorageDriver == "" {
		cfg.StorageDriver = "local"
	}
	cfg.StorageLocalRoot = os.Getenv("STORAGE_LOCAL_ROOT")
	if cfg.StorageLocalRoot == "" {
		cfg.StorageLocalRoot = "/data/media"
	}
	cfg.IcecastURL = os.Getenv("ICECAST_URL")
	if cfg.IcecastURL == "" {
		cfg.IcecastURL = "http://icecast:8000"
	}
	cfg.IcecastAdminUser = os.Getenv("ICECAST_ADMIN_USER")
	if cfg.IcecastAdminUser == "" {
		cfg.IcecastAdminUser = "admin"
	}
	cfg.IcecastAdminPass = os.Getenv("ICECAST_ADMIN_PASS")
	if cfg.IcecastAdminPass == "" {
		cfg.IcecastAdminPass = "hackme"
	}
	if m := os.Getenv("ICECAST_MOUNTS"); m != "" {
		cfg.IcecastMounts = strings.Split(m, ",")
	} else {
		cfg.IcecastMounts = []string{"/main"}
	}
	cfg.GeoIPDatabasePath = os.Getenv("GEOIP_DB_PATH")

	cfg.S3Endpoint = os.Getenv("S3_ENDPOINT")
	cfg.S3Bucket = os.Getenv("S3_BUCKET")
	cfg.S3Region = os.Getenv("S3_REGION")
	cfg.S3AccessKey = os.Getenv("S3_ACCESS_KEY")
	cfg.S3SecretKey = os.Getenv("S3_SECRET_KEY")

	cfg.SMTPHost = os.Getenv("SMTP_HOST")
	cfg.SMTPPort = 587
	if p := os.Getenv("SMTP_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &cfg.SMTPPort)
	}
	cfg.SMTPUsername = os.Getenv("SMTP_USERNAME")
	cfg.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	cfg.SMTPFrom = os.Getenv("SMTP_FROM")

	return cfg, nil
}
