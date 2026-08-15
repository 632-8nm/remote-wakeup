// Package config 负责加载与校验运行配置。
package config

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration. Values are resolved in this order:
// explicit Config overrides > environment variable > .env file > default.
type Config struct {
	MAC            string
	Broadcast      string
	WOLPort        int
	TargetIP       string
	AdminPass      string
	SecretKey      string
	SessionTTL     time.Duration
	Port           string
	AllowedOrigins string // 逗号分隔的可信跨站来源（如 https://wol.example.com）
}

// LoadEnv reads a simple KEY=VALUE dotfile (like .env) into the process
// environment, but only for keys not already set. Lines starting with # are
// comments; blank lines are skipped. No quote stripping to keep it simple.
func LoadEnv(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return // .env is optional
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		if key == "" {
			continue
		}
		if _, ok := os.LookupEnv(key); !ok {
			_ = os.Setenv(key, val)
		}
	}
}

// Default returns configuration with all non-sensitive defaults applied.
// Sensitive/required fields (MAC, TargetIP, AdminPass, SecretKey) are left
// empty so the resolver can fail loudly when they are missing.
func Default() *Config {
	return &Config{
		Broadcast:  "255.255.255.255",
		WOLPort:    9,
		SessionTTL: time.Hour,
		Port:       "5000",
	}
}

// Resolve fills the Config from the current environment (which may have been
// seeded from the .env file) and validates required fields.
func (c *Config) Resolve() {
	c.MAC = os.Getenv("WOL_MAC")
	c.TargetIP = os.Getenv("TARGET_IP")
	c.AdminPass = os.Getenv("ADMIN_PASSWORD")
	c.SecretKey = os.Getenv("SECRET_KEY")
	c.AllowedOrigins = os.Getenv("ALLOWED_ORIGINS")

	if v := os.Getenv("WOL_BROADCAST"); v != "" {
		c.Broadcast = v
	}
	if v, err := strconv.Atoi(os.Getenv("WOL_PORT")); err == nil && v > 0 {
		c.WOLPort = v
	}
	if v, err := strconv.ParseFloat(os.Getenv("SESSION_HOURS"), 64); err == nil && v > 0 {
		c.SessionTTL = time.Duration(v * float64(time.Hour))
	}
	if v := os.Getenv("PORT"); v != "" {
		c.Port = v
	}

	for _, v := range []struct{ name, val string }{
		{"WOL_MAC", c.MAC},
		{"TARGET_IP", c.TargetIP},
		{"ADMIN_PASSWORD", c.AdminPass},
	} {
		if v.val == "" {
			log.Fatalf("缺少必要配置: %s (请设置环境变量或 .env 文件中)", v.name)
		}
	}
}
