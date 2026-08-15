// Package web 实现 HTTP 服务：路由、处理函数、会话、安全防护与页面渲染。
package web

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/632-8nm/remote-wakeup/internal/config"
	"github.com/632-8nm/remote-wakeup/internal/wol"
)

//go:embed templates/index.html templates/login.html
var templatesFS embed.FS

// Server 持有 HTTP 服务运行所需的全部状态。
type Server struct {
	cfg  *config.Config
	sess SessionStore
	tmpl *template.Template
}

// New 创建并初始化 Server。cfg 必须已 Resolve。
func New(cfg *config.Config) *Server {
	return &Server{
		cfg:  cfg,
		sess: NewSessionStore([]byte(cfg.SecretKey), cfg.SessionTTL),
		tmpl: template.Must(template.ParseFS(templatesFS, "templates/login.html")),
	}
}

// Handler 返回已组装好全部路由的 http.Handler。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", s.loginHandler)
	mux.HandleFunc("/logout", s.logoutHandler)
	mux.HandleFunc("/wake", s.loginRequired(s.csrfProtect(s.wakeHandler)))
	mux.HandleFunc("/status", s.loginRequired(s.csrfProtect(s.statusHandler)))
	mux.HandleFunc("/", s.indexHandler)
	return mux
}

// Start 启动监听并阻塞运行。
func (s *Server) Start() error {
	addr := "0.0.0.0:" + s.cfg.Port
	log.Printf("WOL Web 服务启动，监听 %s，目标 %s (%s)，广播 %s:%d",
		addr, s.cfg.TargetIP, s.cfg.MAC, s.cfg.Broadcast, s.cfg.WOLPort)
	return http.ListenAndServe(addr, s.Handler())
}

// 供 handlers.go / security.go 使用
var _ = wol.SendMagicPacket
