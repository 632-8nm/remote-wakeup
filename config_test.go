package main

import (
	"os"
	"testing"
	"time"
)

// 每个用例前清理相关环境变量。
func clearEnv() {
	for _, k := range []string{
		"WOL_MAC", "TARGET_IP", "ADMIN_PASSWORD", "SECRET_KEY",
		"WOL_BROADCAST", "WOL_PORT", "SESSION_HOURS", "PORT", "ALLOWED_ORIGINS",
	} {
		os.Unsetenv(k)
	}
}

func TestDefaultConfig(t *testing.T) {
	c := defaultConfig()
	if c.Broadcast != "255.255.255.255" {
		t.Errorf("默认广播应为 255.255.255.255, 实际 %s", c.Broadcast)
	}
	if c.WOLPort != 9 {
		t.Errorf("默认 WOL 端口应为 9, 实际 %d", c.WOLPort)
	}
	if c.SessionTTL != time.Hour {
		t.Errorf("默认会话时长应为 1h, 实际 %v", c.SessionTTL)
	}
	if c.Port != "5000" {
		t.Errorf("默认端口应为 5000, 实际 %s", c.Port)
	}
}

func TestConfigResolveWithEnv(t *testing.T) {
	clearEnv()
	defer clearEnv()

	os.Setenv("WOL_MAC", "00:11:22:33:44:55")
	os.Setenv("TARGET_IP", "192.168.1.100")
	os.Setenv("ADMIN_PASSWORD", "secret")
	os.Setenv("WOL_BROADCAST", "192.168.1.255")
	os.Setenv("WOL_PORT", "7")
	os.Setenv("SESSION_HOURS", "0.5")
	os.Setenv("PORT", "8080")
	os.Setenv("ALLOWED_ORIGINS", "https://wol.example.com, https://other.example.com")

	c := defaultConfig()
	c.resolve()

	if c.MAC != "00:11:22:33:44:55" {
		t.Errorf("MAC 解析错误: %s", c.MAC)
	}
	if c.TargetIP != "192.168.1.100" {
		t.Errorf("TargetIP 解析错误: %s", c.TargetIP)
	}
	if c.AdminPass != "secret" {
		t.Errorf("AdminPass 解析错误")
	}
	if c.Broadcast != "192.168.1.255" {
		t.Errorf("广播覆盖错误: %s", c.Broadcast)
	}
	if c.WOLPort != 7 {
		t.Errorf("WOL 端口覆盖错误: %d", c.WOLPort)
	}
	if c.SessionTTL != 30*time.Minute {
		t.Errorf("SESSION_HOURS=0.5 应解析为 30 分钟, 实际 %v", c.SessionTTL)
	}
	if c.Port != "8080" {
		t.Errorf("端口覆盖错误: %s", c.Port)
	}
	if c.AllowedOrigins != "https://wol.example.com, https://other.example.com" {
		t.Errorf("AllowedOrigins 解析错误: %q", c.AllowedOrigins)
	}
}

func TestConfigResolveDefaults(t *testing.T) {
	clearEnv()
	defer clearEnv()

	os.Setenv("WOL_MAC", "00:11:22:33:44:55")
	os.Setenv("TARGET_IP", "192.168.1.100")
	os.Setenv("ADMIN_PASSWORD", "secret")

	c := defaultConfig()
	c.resolve()

	// 未设置的项应保持默认
	if c.Broadcast != "255.255.255.255" {
		t.Errorf("未设置广播应保持默认")
	}
	if c.WOLPort != 9 {
		t.Errorf("未设置 WOL 端口应保持默认")
	}
	if c.SessionTTL != time.Hour {
		t.Errorf("未设置 SESSION_HOURS 应保持默认 1h")
	}
	if c.Port != "5000" {
		t.Errorf("未设置 PORT 应保持默认 5000")
	}
}
