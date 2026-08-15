package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const sessionCookieName = "wol_session"

// SessionStore issues and verifies stateless HMAC-signed session cookies.
// No DB, no in-memory store: the cookie itself carries username + expiry,
// signed with a server-secret so it cannot be forged or tampered with.
type SessionStore struct {
	secret []byte
	ttl    time.Duration
}

func NewSessionStore(secret []byte, ttl time.Duration) SessionStore {
	if len(secret) == 0 {
		secret = []byte(randomHex(32))
	}
	return SessionStore{secret: secret, ttl: ttl}
}

// Issue writes a signed cookie for the given username.
func (s SessionStore) Issue(w http.ResponseWriter, username string) {
	exp := time.Now().Add(s.ttl).Unix()
	payload := username + "|" + strconv.FormatInt(exp, 10)
	sig := s.sign(payload)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    payload + "." + sig,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(s.ttl.Seconds()),
	})
}

// Valid reports whether the request carries a valid, unexpired session.
func (s SessionStore) Valid(r *http.Request) bool {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}
	return s.check(c.Value)
}

// Clear removes the session cookie.
func (s SessionStore) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (s SessionStore) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s SessionStore) check(token string) bool {
	i := strings.LastIndexByte(token, '.')
	if i < 0 {
		return false
	}
	payload, sigHex := token[:i], token[i+1:]
	expected := s.sign(payload)

	got, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}
	want, err := hex.DecodeString(expected)
	if err != nil {
		return false
	}
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return false
	}

	// parse expiry
	parts := strings.SplitN(payload, "|", 2)
	if len(parts) != 2 {
		return false
	}
	exp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}
	return time.Now().Unix() < exp
}
