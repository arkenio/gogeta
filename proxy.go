package main

import (
  "fmt"
  "log"
  "net/http"
)

type domainResolver interface {
  resolve(domain string) (http.Handler, bool)
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

  server, found := p.domainResolver.resolve(r.Host)

  if found {
    server.ServeHTTP(w, r)
    return
  }

  http.NotFound(w, r)
}

