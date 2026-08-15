package main

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// ---------- CSRF / Origin 校验 ----------

// originAllowed 校验请求是否来自可信来源，用于防止跨站请求伪造（CSRF）。
// 规则：
//   - 无 Origin 头（非浏览器/同源工具请求）→ 放行（API/curl 场景）
//   - Origin 存在 → 必须与请求 Host 同源，或属于显式配置的可信来源
func originAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // 无 Origin 说明不是浏览器跨站请求（curl/同源 GET 等）
	}

	// 同源：Origin 的 host 部分 == Host 头
	originHost := hostOf(origin)
	if originHost != "" && strings.EqualFold(originHost, r.Host) {
		return true
	}

	// 显式配置的可信来源（如 Cloudflare 域名），逗号分隔
	for _, allowed := range splitCSV(cfg.AllowedOrigins) {
		if strings.EqualFold(hostOf(allowed), originHost) {
			return true
		}
	}
	return false
}

func hostOf(raw string) string {
	// 去掉 scheme 与路径，只留 host[:port]
	s := strings.TrimPrefix(raw, "https://")
	s = strings.TrimPrefix(s, "http://")
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	return s
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// csrfProtect 包装 POST 敏感端点，拒绝跨站来源请求。
func csrfProtect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && !originAllowed(r) {
			writeJSON(w, http.StatusForbidden, map[string]any{
				"success": false, "message": "跨站请求被拒绝",
			})
			return
		}
		next(w, r)
	}
}

// ---------- 登录限速 ----------

type loginAttempt struct {
	fails    int
	lockedAt time.Time
}

var (
	loginMu     sync.Mutex
	loginFails  = map[string]*loginAttempt{}
	lockWindow  = 10 * time.Minute // 锁定时长
	maxFailures = 5                // 阈值：5 次失败锁定
)

// loginThrottled 报告该 IP 是否已被锁定。
func loginThrottled(ip string) bool {
	loginMu.Lock()
	defer loginMu.Unlock()
	a, ok := loginFails[ip]
	if !ok {
		return false
	}
	if a.fails >= maxFailures && time.Since(a.lockedAt) < lockWindow {
		return true
	}
	if time.Since(a.lockedAt) >= lockWindow {
		delete(loginFails, ip)
		return false
	}
	return false
}

// loginFailed 记录一次登录失败。
func loginFailed(ip string) {
	loginMu.Lock()
	defer loginMu.Unlock()
	a, ok := loginFails[ip]
	if !ok {
		a = &loginAttempt{}
		loginFails[ip] = a
	}
	a.fails++
	a.lockedAt = time.Now()
}

// loginSucceeded 登录成功时清零该 IP 记录。
func loginSucceeded(ip string) {
	loginMu.Lock()
	defer loginMu.Unlock()
	delete(loginFails, ip)
}

// clientIP 取请求来源 IP（考虑反代 X-Forwarded-For）。
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host := r.RemoteAddr
	if i := strings.LastIndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	return host
}
