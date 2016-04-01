package main

import (
	goarken "github.com/arkenio/goarken/model"
	"html/template"
	"net/http"
)

type StatusPage struct {
	config *Config
	error  goarken.StatusError
}

type StatusData struct {
	status string
}

func (sp *StatusPage) serve(w http.ResponseWriter, r *http.Request) {

	var code int
	switch sp.error.ComputedStatus {
	case "notfound":
		code = http.StatusNotFound
	case goarken.STARTING_STATUS, goarken.PASSIVATED_STATUS:
		code = http.StatusServiceUnavailable
	default:
		code = http.StatusInternalServerError

	}

	templateDir := sp.config.templateDir
	tmpl, err := template.ParseFiles(templateDir+"/main.tpl", templateDir+"/body_"+sp.error.ComputedStatus+".tpl")
	if err != nil {
		http.Error(w, "Unable to serve page : "+sp.error.ComputedStatus, code)

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	err = tmpl.ExecuteTemplate(w, "main", &StatusData{sp.error.ComputedStatus})
	if err != nil {
		http.Error(w, "Failed to execute templates : "+err.Error(), 500)
	}
}
