package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"github.com/golang/glog"
)

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


type proxyHandler func(http.ResponseWriter, *http.Request) (*Config, error)

func (ph proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if r:=recover();r!=nil {
			http.Error(w,"An error occured serving request",500)
			log.Panicln("Recovered from error", r)
		}
	}()

	if c,err := ph(w, r); err != nil {
		ph.OnError(w,r,err,c)
	}
}

func (ph proxyHandler) OnError(w http.ResponseWriter, r *http.Request, error error, c *Config) {
	if stError, ok := error.(StatusError); ok {
		sp := &StatusPage{c,stError}
		sp.serve(w,r)
	} else {
		sp := &StatusPage{c,StatusError{"notfound", nil}}
		sp.serve(w,r)
	}
}



func (p *proxy) start() {
	log.Printf("Listening on port %d", p.config.port)
	http.Handle("/__static__/", http.FileServer(http.Dir(p.config.templateDir)))
	http.Handle("/", proxyHandler(p.proxy))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", p.config.port), nil))
}


func (p *proxy) proxy(w http.ResponseWriter, r *http.Request) (*Config, error) {
	host := hostnameOf(r.Host)
	if server, err := p.domainResolver.resolve(host); err != nil {
		return p.config, err
	} else {
		server.ServeHTTP(w, r)
		return p.config, nil
	}
}



func hostnameOf(host string) string {
	hostname := strings.Split(host, ":")[0]

	if len(hostname) > 4 && hostname[0:4] == "www." {
		hostname = hostname[4:]
	}

	return hostname
}
