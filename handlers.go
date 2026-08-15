package main

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os/exec"
	"time"
)

// loginRequired is middleware that redirects unauthenticated requests to /login.
func loginRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !sess.Valid(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if !sess.Valid(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	serveTemplate(w, "index.html")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	type page struct{ Error string }
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == "admin" && authPass(password) {
			sess.Issue(w, "admin")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		serveRender(w, "login.html", page{"用户名或密码错误"})
		return
	}
	serveRender(w, "login.html", page{})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	sess.Clear(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func wakeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false, "message": "仅支持 POST",
		})
		return
	}
	if err := SendMagicPacket(cfg.MAC, cfg.Broadcast, cfg.WOLPort); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false, "message": "发送失败: " + err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "message": "已发送唤醒包",
	})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	out, err := exec.Command("ping", "-c", "1", "-W", "2", cfg.TargetIP).CombinedOutput()
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
func authPass(submitted string) bool {
	return subtle.ConstantTimeCompare([]byte(submitted), []byte(cfg.AdminPass)) == 1
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
