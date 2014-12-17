package main

import (
	"errors"
	"fmt"
	"github.com/arkenio/goarken"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type EnvResolver struct {
	config          *Config
	watcher         *goarken.Watcher
	services        map[string]*goarken.ServiceCluster
	dest2ProxyCache map[string]http.Handler
}

func NewEnvResolver(c *Config) *EnvResolver {
	services := make(map[string]*goarken.ServiceCluster)

	client, err := c.getEtcdClient()

	if err != nil {
		panic(err)
	}

	w := &goarken.Watcher{
		Client:        client,
		DomainPrefix:  "/domains",
		ServicePrefix: "/services",
		Domains:       nil,
		Services:      services,
	}

	return &EnvResolver{c, w, services, make(map[string]http.Handler)}
}

func (r *EnvResolver) resolve(domain string) (http.Handler, error) {
	serviceName := strings.Split(domain, ".")[0]

	serviceTree := r.services[serviceName]
	if serviceTree != nil {

		if service, err := serviceTree.Next(); err != nil {
			uri := fmt.Sprintf("http://%s:%d/", service.Location.Host, service.Location.Port)
			return r.getOrCreateProxyFor(uri), nil
		}
	}

	return nil, errors.New("Unable to resolve")

}

func (r *EnvResolver) init() {
	r.watcher.Init()
}

func (r *EnvResolver) redirectToStatusPage(domainName string) string {
	return ""
}

func (r *EnvResolver) getOrCreateProxyFor(uri string) http.Handler {
	if _, ok := r.dest2ProxyCache[uri]; !ok {
		dest, _ := url.Parse(uri)
		r.dest2ProxyCache[uri] = httputil.NewSingleHostReverseProxy(dest)
	}
	return r.dest2ProxyCache[uri]
}
