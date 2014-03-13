package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Domain struct {
	typ    string
	value  string
	server http.Handler
}

type Environment struct {
	ip     string
	port   string
	server http.Handler
}

type IoEtcdResolver struct {
	config       *Config
	watcher      *watcher
	domains      map[string]*Domain
	environments map[string]*Environment
}

func NewEtcdResolver(c *Config) *IoEtcdResolver {
	domains := make(map[string]*Domain)
	envs := make(map[string]*Environment)
	w := NewEtcdWatcher(c, domains, envs)
	return &IoEtcdResolver{c, w, domains, envs}
}

func (r *IoEtcdResolver) init() {
	r.watcher.init()
}

func (r *IoEtcdResolver) resolve(domainName string) (http.Handler, bool) {
	domain := r.domains[domainName]
	if domain != nil {
		if domain.server == nil {
			switch domain.typ {
			case "iocontainer":
				env := r.environments[domain.value]
				uri := ""
				if env.port != "80" {
					uri = fmt.Sprintf("http://%s:%s/", env.ip, env.port)

				} else {
					uri = fmt.Sprintf("http://%s/", env.ip)
				}
				dest, _ := url.Parse(uri)
				domain.server = httputil.NewSingleHostReverseProxy(dest)

			case "uri":
				dest, _ := url.Parse(domain.value)
				domain.server = httputil.NewSingleHostReverseProxy(dest)
			}
		}

		return domain.server, true
	}
	return nil, false
}
