package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var page = `<html>
  <body>
    {{template "content" .Content}}
  </body>
</html>`

var content = `{{define "content"}}
<div>
   <p>{{.Title}}</p>
   <p>{{.Content}}</p>
</div>
{{end}}`

type Content struct {
	Title   string
	Content string
}

type Page struct {
	Content *Content
}

type domainResolver interface {
	resolve(domain string) (http.Handler, error)
	init()
}

type proxy struct {
	config         *Config
	domainResolver domainResolver
}

func NewProxy(c *Config, resolver domainResolver) *proxy {
	return &proxy{c, resolver}
}


type proxyHandler func(http.ResponseWriter, *http.Request) error

func (ph proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := ph(w, r); err != nil {
		ph.OnError(w,r,err)
	}
}

func (ph proxyHandler) OnError(w http.ResponseWriter, r *http.Request, error error) {
	if stError, ok := error.(StatusError); ok {
		//TODO: refactor with templates on filesystem
		pagedata := &Page{Content: &Content{Title: "Status", Content: stError.computedStatus}}
		tmpl, err := template.New("page").Parse(page)
		tmpl, err = tmpl.Parse(content)
		if err == nil {
			tmpl.Execute(w, pagedata)
		}
	} else {
		http.NotFound(w, r)
	}
}



func (p *proxy) start() {
	log.Printf("Listening on port %d", p.config.port)
	http.Handle("/", proxyHandler(p.proxy))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", p.config.port), nil))
}


func (p *proxy) proxy(w http.ResponseWriter, r *http.Request) error {
	host := hostnameOf(r.Host)
	if server, err := p.domainResolver.resolve(host); err != nil {
		return err
	} else {
		server.ServeHTTP(w, r)
		return nil
	}
}



func hostnameOf(host string) string {
	hostname := strings.Split(host, ":")[0]

	if len(hostname) > 4 && hostname[0:4] == "www." {
		hostname = hostname[4:]
	}

	return hostname
}
