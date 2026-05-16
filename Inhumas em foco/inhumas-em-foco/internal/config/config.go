package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                  string
	DBDriver              string
	DatabaseURL           string
	MigrationsDir         string
	SessionSecret         string
	PreviousSessionSecret string
	AdminPathPrefix       string
	MaintenanceMode       bool
	UploadDir             string
	StaticDir             string
	SiteURL               string
	ProjectRoot           string
	MetricsToken          string
	BackupDir             string
	BackupScript          string
	CSPReportOnly         bool
	CSPReportURI          string
	SMTPHost              string
	SMTPPort              string
	SMTPUsername          string
	SMTPPassword          string
	SMTPFrom              string
	SMTPFromName          string
	MaxUploadSize         int64
	DefaultBcryptCost     int
	OriginalRetentionDays int
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	maxUploadSize, _ := strconv.ParseInt(os.Getenv("MAX_UPLOAD_SIZE"), 10, 64)
	if maxUploadSize == 0 {
		maxUploadSize = 2 * 1024 * 1024 // 2MB
	}
	bcryptCost, _ := strconv.Atoi(os.Getenv("BCRYPT_COST"))
	if bcryptCost == 0 {
		bcryptCost = 12
	}
	originalRetentionDays, _ := strconv.Atoi(os.Getenv("ORIGINAL_RETENTION_DAYS"))
	if originalRetentionDays == 0 {
		originalRetentionDays = 7
	}
	maintMode, _ := strconv.ParseBool(os.Getenv("MAINTENANCE_MODE"))
	cspReportOnly, _ := strconv.ParseBool(os.Getenv("CSP_REPORT_ONLY"))
	return &Config{
		Port:                  port,
		DBDriver:              strings.ToLower(strings.TrimSpace(getEnv("DB_DRIVER", ""))),
		DatabaseURL:           getEnv("DATABASE_URL", "./inhumas.db"),
		MigrationsDir:         getEnv("MIGRATIONS_DIR", "./migrations"),
		SessionSecret:         getEnv("SESSION_SECRET", "change-me-in-production-32-bytes-minimum-length-here"),
		PreviousSessionSecret: os.Getenv("PREVIOUS_SESSION_SECRET"),
		AdminPathPrefix:       getEnv("ADMIN_PATH_PREFIX", "/painel/7x9k2m"),
		MaintenanceMode:       maintMode,
		UploadDir:             getEnv("UPLOAD_DIR", "./uploads"),
		StaticDir:             getEnv("STATIC_DIR", "./static"),
		SiteURL:               strings.TrimRight(getEnv("SITE_URL", "https://inhumasemfoco.com.br"), "/"),
		ProjectRoot:           getEnv("PROJECT_ROOT", "."),
		MetricsToken:          os.Getenv("METRICS_TOKEN"),
		BackupDir:             os.Getenv("BACKUP_DIR"),
		BackupScript:          os.Getenv("BACKUP_SCRIPT"),
		CSPReportOnly:         cspReportOnly,
		CSPReportURI:          os.Getenv("CSP_REPORT_URI"),
		SMTPHost:              os.Getenv("SMTP_HOST"),
		SMTPPort:              getEnv("SMTP_PORT", "587"),
		SMTPUsername:          os.Getenv("SMTP_USERNAME"),
		SMTPPassword:          os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:              os.Getenv("SMTP_FROM"),
		SMTPFromName:          getEnv("SMTP_FROM_NAME", "Inhumas em Foco"),
		MaxUploadSize:         maxUploadSize,
		DefaultBcryptCost:     bcryptCost,
		OriginalRetentionDays: originalRetentionDays,
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func (c *Config) IsValidSessionSecret() bool {
	return len(c.SessionSecret) >= 32
}

func (c *Config) IsValidPreviousSessionSecret() bool {
	return c.PreviousSessionSecret == "" || len(c.PreviousSessionSecret) >= 32
}

func (c *Config) IsLocal() bool {
	if strings.EqualFold(os.Getenv("FORCE_SECURE_COOKIES"), "true") {
		return false
	}
	env := strings.ToLower(os.Getenv("APP_ENV"))
	if env == "production" || env == "prod" {
		return false
	}
	if env == "local" || env == "development" || env == "dev" || env == "test" {
		return true
	}
	return strings.Contains(c.DatabaseURL, ".db")
}

func (c *Config) SMTPEnabled() bool {
	return c.SMTPHost != "" && c.SMTPFrom != ""
}
