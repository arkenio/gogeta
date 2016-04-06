package main

import (
	"errors"
	"fmt"
	"github.com/arkenio/goarken/model"
	"github.com/arkenio/goarken/storage"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type EnvResolver struct {
	config          *Config
	arkenModel      *model.Model
	dest2ProxyCache map[string]http.Handler
}

func NewEnvResolver(c *Config) *EnvResolver {

	client, err := c.getEtcdClient()

	if err != nil {
		panic(err)
	}

	persistenceDriver := storage.NewWatcher(client, "/services", "/domains")

	arkenModel, err := model.NewArkenModel(nil, persistenceDriver)
	if err != nil {
		return nil
	}

	return &EnvResolver{c, arkenModel, make(map[string]http.Handler)}
}

func (r *EnvResolver) resolve(domain string) (http.Handler, error) {
	serviceName := strings.Split(domain, ".")[0]

	serviceTree := r.arkenModel.Services[serviceName]
	if serviceTree != nil {

		if service, err := serviceTree.Next(); err != nil {
			uri := fmt.Sprintf("http://%s:%d/", service.Location.Host, service.Location.Port)
			return r.getOrCreateProxyFor(uri), nil
		}
	}

	return nil, errors.New("Unable to resolve")

}

func (r *EnvResolver) init() {

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
