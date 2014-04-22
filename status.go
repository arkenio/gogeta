package main

import (
	"html/template"
	"net/http"
)

type StatusData struct {
	status string
}

type StatusPage struct {
	config *Config
	error  StatusError
}

func (sp *StatusPage) serve(w http.ResponseWriter, r *http.Request) {

	var code int
	switch sp.error.computedStatus {
	case "notfound":
		code = 404
		break
	case "starting":
		code = 503
		break
	default:
		code = 500

	}


	templateDir := sp.config.templateDir
	tmpl, err := template.ParseFiles(templateDir+"/main.tpl", templateDir+"/body_"+sp.error.computedStatus+".tpl")
	if err != nil {
		http.Error(w, "Unable to serve page : " + sp.error.computedStatus, code)
		return
	}


	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	err = tmpl.ExecuteTemplate(w, "main", &StatusData{sp.error.computedStatus})
	if err != nil {
		http.Error(w, "Failed to execute templates : "+err.Error(), 500)
	}
}
