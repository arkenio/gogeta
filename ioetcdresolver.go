package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/golang/glog"
)

const (
	SERVICE_DOMAINTYTPE = "service"
	URI_DOMAINTYPE      = "uri"
)

type Domain struct {
	typ    string
	value  string
}

type service struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Environment struct {
	key      string
	location service
	domain   string
	name     string
	status   *Status
}

type IoEtcdResolver struct {
	config          *Config
	watcher         *watcher
	domains         map[string]*Domain
	environments    map[string]*EnvironmentCluster
	dest2ProxyCache map[string]http.Handler
	watchIndex      uint64
}

func NewEtcdResolver(c *Config) *IoEtcdResolver {
	domains := make(map[string]*Domain)
	envs := make(map[string]*EnvironmentCluster)
	dest2ProxyCache := make(map[string]http.Handler)
	w := NewEtcdWatcher(c, domains, envs)
	return &IoEtcdResolver{c, w, domains, envs, dest2ProxyCache, 0}
}

func (r *IoEtcdResolver) init() {
	r.watcher.init()
}

func (env *Environment) Dump() {
	log.Printf("Dumping environment %s : ", env.name)
	log.Printf("   domain : %s", env.domain)
	log.Printf("   location : %s:%d", env.location.Host, env.location.Port)
}

func (domain *Domain) equals(other *Domain) bool {
	if domain == nil && other == nil {
		return true
	}

	return domain != nil && other != nil &&
		domain.typ == other.typ && domain.value == other.value
}

func (env *Environment) equals(other *Environment) bool {

	return env != nil && other != nil &&
		env.location.Host == other.location.Host &&
		env.location.Port == other.location.Port &&
		env.status.equals(other.status)
}

func (r *IoEtcdResolver) resolve(domainName string) (http.Handler, error) {
	glog.V(5).Infof("Looking for domain : %s ", domainName)
	domain := r.domains[domainName]
	if domain != nil {
		switch domain.typ {

		case SERVICE_DOMAINTYTPE:
			if env, err := r.environments[domain.value].Next(); err == nil {
				addr := net.JoinHostPort(env.location.Host, strconv.Itoa(env.location.Port))
				uri := fmt.Sprintf("http://%s/", addr)

				return r.getOrCreateProxyFor(uri), nil

			} else {
				return nil, err
			}
		case URI_DOMAINTYPE:
			return r.getOrCreateProxyFor(domain.value), nil
		}

	}
	glog.V(5).Infof("Domain %s not found", domainName)
	return nil, errors.New("Domain not found")
}

func (r *IoEtcdResolver) getOrCreateProxyFor(uri string) http.Handler {
	if _, ok := r.dest2ProxyCache[uri]; !ok {
		dest, _ := url.Parse(uri)
		r.dest2ProxyCache[uri] = httputil.NewSingleHostReverseProxy(dest)
	}
	return r.dest2ProxyCache[uri]
}
