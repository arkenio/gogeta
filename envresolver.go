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
	config  *Config
	watcher *watcher
	envs    map[string]*EnvironmentCluster
}

func NewEnvResolver(c *Config) *EnvResolver {
	envs := make(map[string]*EnvironmentCluster)
	w := NewEtcdWatcher(c, nil, envs)
	return &EnvResolver{c, w, envs}
}

func (r *EnvResolver) resolve(domain string) (http.Handler, error) {
	envName := strings.Split(domain, ".")[0]

	envTree := r.envs[envName]
	if envTree != nil {

		if env, err := envTree.Next(); err != nil {
			if env.server == nil {
				uri := ""
				uri = fmt.Sprintf("http://%s:%d/", env.location.Host, env.location.Port)
				dest, _ := url.Parse(uri)
				env.server = httputil.NewSingleHostReverseProxy(dest)
			}
			return env.server, nil
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
