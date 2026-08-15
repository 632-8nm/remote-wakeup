package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestStore(ttl time.Duration) SessionStore {
	return NewSessionStore([]byte("test-secret-0123456789"), ttl)
}

// issueToken 签发一个会话 cookie 并返回其值。
func issueToken(t *testing.T, s SessionStore) string {
	t.Helper()
	rec := httptest.NewRecorder()
	s.Issue(rec, "admin")
	cookies := rec.Result().Cookies()
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			return c.Value
		}
	}
	t.Fatal("未找到会话 cookie")
	return ""
}

func TestSessionIssueAndValid(t *testing.T) {
	s := newTestStore(time.Hour)
	token := issueToken(t, s)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})

	if !s.Valid(req) {
		t.Fatal("有效会话应通过校验")
	}
}

func TestSessionMissingCookie(t *testing.T) {
	s := newTestStore(time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if s.Valid(req) {
		t.Fatal("无 cookie 应判定无效")
	}
}

func TestSessionTamperedSignature(t *testing.T) {
	s := newTestStore(time.Hour)
	token := issueToken(t, s)

	// 篡改 payload 中的用户名
	parts := strings.SplitN(token, ".", 2)
	payload := strings.Replace(parts[0], "admin", "evil", 1)
	tampered := payload + "." + parts[1]

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: tampered})
	if s.Valid(req) {
		t.Fatal("篡改签名的会话应判定无效")
	}
}

func TestSessionWrongSecret(t *testing.T) {
	s1 := newTestStore(time.Hour)
	s2 := NewSessionStore([]byte("different-secret"), time.Hour)

	token := issueToken(t, s1)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})

	if s2.Valid(req) {
		t.Fatal("用不同密钥校验应判定无效")
	}
}

func TestSessionExpired(t *testing.T) {
	// 过期时间设为负值，立即过期
	s := newTestStore(-time.Minute)
	token := issueToken(t, s)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	if s.Valid(req) {
		t.Fatal("过期会话应判定无效")
	}
}

func TestSessionClear(t *testing.T) {
	s := newTestStore(time.Hour)
	rec := httptest.NewRecorder()
	s.Clear(rec)
	// Clear 应设置过期 cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName && c.MaxAge < 0 {
			return // OK
		}
	}
	t.Fatal("Clear 应写入 MaxAge<0 的过期 cookie")
}
