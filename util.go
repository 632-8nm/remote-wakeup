package main

import (
	"crypto/rand"
	"encoding/hex"
)

// randomHex returns n random bytes hex-encoded. Used for the fallback
// session secret and as a per-request CSRF-ish nonce. Fatal-free: if the
// entropy source fails it logs nothing meaningful, so panic is acceptable
// here (it can only fail if the OS RNG is broken).
func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("无法读取系统随机源: " + err.Error())
	}
	return hex.EncodeToString(b)
}
