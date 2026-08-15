package web

import (
	"log"
	"net/http"
)

// serveTemplate writes an embedded static template file verbatim with the
// HTTP response content type of text/html. Used for pages without variables
// (currently index.html).
func (s *Server) serveTemplate(w http.ResponseWriter, name string) {
	data, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		log.Printf("无法读取模板 %s: %v", name, err)
		http.Error(w, "模板错误", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

// serveRender renders the login template with data using html/template.
// With html/template, `{{if .Error}}...{{end}}` blocks are omitted entirely
// when Error is empty, so no placeholder can ever leak into the page.
func (s *Server) serveRender(w http.ResponseWriter, name string, data any) {
	if s.tmpl == nil {
		http.Error(w, "模板未初始化", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.Execute(w, data); err != nil {
		log.Printf("渲染模板 %s 失败: %v", name, err)
	}
}
