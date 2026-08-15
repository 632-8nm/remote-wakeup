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
	hw, err := net.ParseMAC(normalizeMAC(macStr))
	if err != nil {
		return fmt.Errorf("无效 MAC 地址: %w", err)
	}

	packet := make([]byte, 0, 6+16*6)
	for i := 0; i < 6; i++ {
		packet = append(packet, 0xFF)
	}
	for i := 0; i < 16; i++ {
		packet = append(packet, hw...)
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

// normalizeMAC maps hyphen- or no-separator MAC forms to the colon form that
// net.ParseMAC expects (net.ParseMAC already accepts colons and some other
// separators, but not all).
func normalizeMAC(s string) string {
	if !strings.Contains(s, ":") {
		s = strings.Join([]string{
			s[0:2], s[2:4], s[4:6], s[6:8], s[8:10], s[10:12],
		}, ":")
	}
	return s
}
