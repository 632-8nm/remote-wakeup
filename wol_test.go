package main

import (
	"bytes"
	"testing"
)

func TestBuildMagicPacket(t *testing.T) {
	// 标准冒号分隔 MAC
	packet, err := buildMagicPacket("00:11:22:33:44:55")
	if err != nil {
		t.Fatalf("buildMagicPacket 出错: %v", err)
	}

	// 长度：6 前缀 + 16 次重复 × 6 字节 = 102
	if len(packet) != 102 {
		t.Fatalf("包长度应为 102, 实际 %d", len(packet))
	}

	// 前 6 字节全为 0xFF
	wantFF := bytes.Repeat([]byte{0xFF}, 6)
	if !bytes.Equal(packet[:6], wantFF) {
		t.Fatalf("前 6 字节应为 0xFF*6, 实际 %x", packet[:6])
	}

	// 后续应为 MAC 重复 16 次
	mac := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	for i := 0; i < 16; i++ {
		start := 6 + i*6
		if !bytes.Equal(packet[start:start+6], mac) {
			t.Fatalf("第 %d 次 MAC 重复不正确: %x", i, packet[start:start+6])
		}
	}
}

func TestBuildMagicPacketHyphenMAC(t *testing.T) {
	// 连字符分隔的 MAC 也能正确构造
	packet, err := buildMagicPacket("00-11-22-33-44-55")
	if err != nil {
		t.Fatalf("连字符 MAC 构造出错: %v", err)
	}
	if len(packet) != 102 {
		t.Fatalf("包长度应为 102, 实际 %d", len(packet))
	}
	mac := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	if !bytes.Equal(packet[6:12], mac) {
		t.Fatalf("连字符 MAC 解析错误: %x", packet[6:12])
	}
}

func TestBuildMagicPacketInvalidMAC(t *testing.T) {
	// 非法 MAC 应返回错误
	if _, err := buildMagicPacket("not-a-mac"); err == nil {
		t.Fatal("非法 MAC 应返回错误")
	}
	if _, err := buildMagicPacket(""); err == nil {
		t.Fatal("空 MAC 应返回错误")
	}
}

func TestNormalizeMAC(t *testing.T) {
	cases := []struct{ in, want string }{
		{"00:11:22:33:44:55", "00:11:22:33:44:55"}, // 已是冒号形式
		{"00-11-22-33-44-55", "00:11:22:33:44:55"}, // 连字符→冒号
	}
	for _, c := range cases {
		if got := normalizeMAC(c.in); got != c.want {
			t.Errorf("normalizeMAC(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
