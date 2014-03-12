package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type domainResolver interface {
	resolve(domain string) (http.Handler, bool)
	init()
}

type proxy struct {
	config *Config
	domainResolver domainResolver
}

func NewProxy(c *Config, resolver domainResolver) *proxy {
	return &proxy{c, resolver}
}

func (p *proxy) start() {
	log.Printf("Listening on port %d", p.config.port)
	http.HandleFunc("/", p.OnRequest)
	http.ListenAndServe(fmt.Sprintf(":%d", p.config.port), nil)

}

func (p *proxy) OnRequest(w http.ResponseWriter, r *http.Request) {

	server, found := p.domainResolver.resolve(hostnameOf(r.Host))

	if found {
		server.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}


func hostnameOf(host string) string {
	hostname := strings.Split(host, ":")[0]

	if len(hostname) > 4 && hostname[0:4] == "www." {
		hostname = hostname[4:]
	}

	return hostname
}

