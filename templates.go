package main

import (
	"html/template"
	"log"
	"net/http"
)

// Parse the embedded login template once at startup.
var loginTmpl = template.Must(template.ParseFS(templatesFS, "templates/login.html"))

// serveTemplate writes an embedded static template file verbatim with the
// HTTP response content type of text/html. Used for pages without variables
// (currently index.html).
func serveTemplate(w http.ResponseWriter, name string) {
	data, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		log.Printf("无法读取模板 %s: %v", name, err)
		http.Error(w, "模板错误", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

// serveRender renders an embedded template with data using html/template.
// With html/template, `{{if .Error}}...{{end}}` blocks are omitted entirely
// when Error is empty, so no placeholder can ever leak into the page.
func serveRender(w http.ResponseWriter, name string, data any) {
	var tmpl *template.Template
	switch name {
	case "login.html":
		tmpl = loginTmpl
	default:
		log.Printf("未知模板 %s", name)
		http.Error(w, "模板错误", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("渲染模板 %s 失败: %v", name, err)
	}
}
