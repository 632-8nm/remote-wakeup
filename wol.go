package main

import (
	"fmt"
	"net"
	"strings"
)

// SendMagicPacket builds and sends a Wake-on-LAN magic packet:
// 6 bytes of 0xFF followed by the target MAC repeated 16 times.
// addr may be a unicast IP or a broadcast address such as 255.255.255.255.
func SendMagicPacket(macStr, addr string, port int) error {
	packet, err := buildMagicPacket(macStr)
	if err != nil {
		return err
	}

	dst := &net.UDPAddr{IP: net.ParseIP(addr), Port: port}
	// ListenUDP with a wildcard source, then WriteToUDP so the socket can
	// send to a broadcast destination (SO_BROADCAST is implied for UDP
	// broadcast in Go when the destination is a broadcast address).
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return fmt.Errorf("无法创建 UDP socket: %w", err)
	}
	defer conn.Close()

	_, err = conn.WriteToUDP(packet, dst)
	if err != nil {
		return fmt.Errorf("发送 WOL 包失败: %w", err)
	}
	return nil
}

// buildMagicPacket 构造 WOL 魔术包：6 字节 0xFF + MAC 重复 16 次。
// 纯函数，便于单元测试。
func buildMagicPacket(macStr string) ([]byte, error) {
	hw, err := net.ParseMAC(normalizeMAC(macStr))
	if err != nil {
		return nil, fmt.Errorf("无效 MAC 地址: %w", err)
	}

	packet := make([]byte, 0, 6+16*6)
	for i := 0; i < 6; i++ {
		packet = append(packet, 0xFF)
	}
	for i := 0; i < 16; i++ {
		packet = append(packet, hw...)
	}
	return packet, nil
}

// normalizeMAC maps hyphen- or no-separator MAC forms to the colon form that
// net.ParseMAC expects (net.ParseMAC already accepts colons and some other
// separators, but not all). Returns the input unchanged if it cannot be
// normalised safely (letting net.ParseMAC produce the error).
func normalizeMAC(s string) string {
	if strings.Contains(s, ":") {
		return s
	}
	// 去掉连字符，仅保留 12 个十六进制字符
	digits := strings.ReplaceAll(s, "-", "")
	if len(digits) != 12 {
		return s // 无法安全规范化，交给 net.ParseMAC 报错
	}
	return strings.Join([]string{
		digits[0:2], digits[2:4], digits[4:6], digits[6:8], digits[8:10], digits[10:12],
	}, ":")
}
