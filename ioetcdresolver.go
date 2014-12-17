package main

import (
	"errors"
	"fmt"
	"github.com/arkenio/goarken"
	"github.com/golang/glog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	SERVICE_DOMAINTYTPE = "service"
	URI_DOMAINTYPE      = "uri"
	TIME_FORMAT         = "2006-01-02 15:04:05"
)

type IoEtcdResolver struct {
	config          *Config
	watcher         *goarken.watcher
	domains         map[string]*goarken.Domain
	services        map[string]*goarken.ServiceCluster
	dest2ProxyCache map[string]*ServiceMux
	watchIndex      uint64
}

func NewEtcdResolver(c *Config) (*IoEtcdResolver, error) {

	domains := make(map[string]*Domain)
	services := make(map[string]*ServiceCluster)
	dest2ProxyCache := make(map[string]*ServiceMux)

	client, err := c.getEtcdClient()

	if err != nil {
		panic(err)
	}
	w := &goarken.Watcher{
		Client:        client,
		DomainPrefix:  "/domains",
		ServicePrefix: "/services",
		Domains:       domains,
		Services:      services,
	}

	return &IoEtcdResolver{c, w, domains, services, dest2ProxyCache, 0}, nil
}

func (r *IoEtcdResolver) init() {
	r.watcher.init()
}

type ServiceConfig struct {
	Robots string `json:"robots"`
}

func (config *ServiceConfig) equals(other *ServiceConfig) bool {
	if config == nil && other == nil {
		return true
	}

	return config != nil && other != nil &&
		config.Robots == other.Robots
}

func (r *IoEtcdResolver) resolve(domainName string) (http.Handler, error) {
	glog.V(5).Infof("Looking for domain : %s ", domainName)
	domain := r.domains[domainName]
	glog.V(5).Infof("Services:%s", r.services)
	if domain != nil {
		service := r.services[domain.Value]
		if service == nil {
			glog.Errorf("The services map doesn't contain service with the domain value: %s", domain.Value)
		}
		switch domain.Typ {

		case SERVICE_DOMAINTYTPE:
			service, err := r.services[domain.Value].Next()
			if err == nil && service.Location.IsFullyDefined() {
				addr := net.JoinHostPort(service.Location.Host, strconv.Itoa(service.Location.Port))
				uri := fmt.Sprintf("http://%s/", addr)
				r.setLastAccessTime(service)
				return r.getOrCreateProxyFor(service, uri), nil

			} else {
				return nil, err
			}
		case URI_DOMAINTYPE:
			return r.getOrCreateProxyFor(nil, domain.value), nil
		}

	}
	glog.V(5).Infof("Domain %s not found", domainName)
	return nil, errors.New("Domain not found")
}

func (r *IoEtcdResolver) setLastAccessTime(service *goarken.Service) {

	interval := time.Duration(r.config.lastAccessInterval) * time.Second
	if service.LastAccess == nil || service.LastAccess.Add(interval).Before(time.Now()) {
		lastAccessKey := fmt.Sprintf("%s/lastAccess", service.NodeKey)

		client, error := r.config.getEtcdClient()
		if error != nil {
			glog.Errorf("Unable to get etcd client : %s ", error)
			return
		}

		now := time.Now()
		service.LastAccess = &now

		t := service.LastAccess.Format(TIME_FORMAT)
		_, error = client.Set(lastAccessKey, t, 0)

		glog.V(5).Infof("Settign lastAccessKey to :%s", t)
		if error != nil {
			glog.Errorf("error :%s", error)
		}
	}

}

func (r *IoEtcdResolver) getOrCreateProxyFor(s *Service, uri string) http.Handler {
	if serviceMux, ok := r.dest2ProxyCache[uri]; !ok || !serviceMux.service.equals(s) {
		glog.Infof("Creating a new muxer for : %s", s.name)
		r.dest2ProxyCache[uri] = NewServiceMux(r.config, s, uri)
	}
	return r.dest2ProxyCache[uri]
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
