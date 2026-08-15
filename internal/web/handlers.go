package web

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os/exec"
	"time"

	"github.com/632-8nm/remote-wakeup/internal/wol"
)

// loginRequired is middleware that redirects unauthenticated requests to /login.
func (s *Server) loginRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.sess.Valid(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if !s.sess.Valid(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	s.serveTemplate(w, "index.html")
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	type page struct{ Error string }
	if r.Method == http.MethodPost {
		ip := clientIP(r)
		if loginThrottled(ip) {
			s.serveRender(w, "login.html", page{"尝试次数过多，请稍后再试"})
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == "admin" && s.authPass(password) {
			loginSucceeded(ip)
			s.sess.Issue(w, "admin")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		loginFailed(ip)
		s.serveRender(w, "login.html", page{"用户名或密码错误"})
		return
	}
	s.serveRender(w, "login.html", page{})
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	s.sess.Clear(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) wakeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false, "message": "仅支持 POST",
		})
		return
	}
	if err := wol.SendMagicPacket(s.cfg.MAC, s.cfg.Broadcast, s.cfg.WOLPort); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false, "message": "发送失败: " + err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "message": "已发送唤醒包",
	})
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	out, err := exec.Command("ping", "-c", "1", "-W", "2", s.cfg.TargetIP).CombinedOutput()
	_ = out
	online := err == nil
	resp := map[string]any{"online": online}
	if online {
		resp["ping_ms"] = int(time.Since(start).Milliseconds())
	}
	writeJSON(w, http.StatusOK, resp)
}

// authPass compares the submitted password against the single source of
// truth ($ADMIN_PASSWORD) in constant time. The board has no internet access,
// so the stdlib's crypto/subtle is used instead of an external bcrypt lib.
func (s *Server) authPass(submitted string) bool {
	return subtle.ConstantTimeCompare([]byte(submitted), []byte(s.cfg.AdminPass)) == 1
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
