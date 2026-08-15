package main

import (
	"embed"
	"log"
	"net/http"
)

//go:embed templates/index.html templates/login.html
var templatesFS embed.FS

var (
	cfg  *Config
	sess SessionStore
)

func main() {
	// Optionally load a dotenv file (secrets stay out of the repo).
	loadEnv(".env")

	cfg = defaultConfig()
	cfg.resolve()
	sess = NewSessionStore([]byte(cfg.SecretKey), cfg.SessionTTL)

	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/wake", loginRequired(csrfProtect(wakeHandler)))
	mux.HandleFunc("/status", loginRequired(csrfProtect(statusHandler)))
	mux.HandleFunc("/", indexHandler)

	addr := "0.0.0.0:" + cfg.Port
	log.Printf("WOL Web 服务启动，监听 %s，目标 %s (%s)，广播 %s:%d",
		addr, cfg.TargetIP, cfg.MAC, cfg.Broadcast, cfg.WOLPort)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
