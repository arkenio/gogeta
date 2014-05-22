package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type EnvResolver struct {
	config          *Config
	watcher         *watcher
	envs            map[string]*EnvironmentCluster
	dest2ProxyCache map[string]http.Handler
}

func NewEnvResolver(c *Config) *EnvResolver {
	envs := make(map[string]*EnvironmentCluster)
	w := NewEtcdWatcher(c, nil, envs)
	return &EnvResolver{c, w, envs, make(map[string]http.Handler)}
}

func (r *EnvResolver) resolve(domain string) (http.Handler, error) {
	envName := strings.Split(domain, ".")[0]

	envTree := r.envs[envName]
	if envTree != nil {

		if env, err := envTree.Next(); err != nil {
			uri := fmt.Sprintf("http://%s:%d/", env.location.Host, env.location.Port)
			return r.getOrCreateProxyFor(uri), nil
		}
	}

	return nil, errors.New("Unable to resolve")

}

func (r *EnvResolver) init() {
	r.watcher.loadAndWatch(r.config.envPrefix, r.watcher.registerEnvironment)
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
