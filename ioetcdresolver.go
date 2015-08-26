package main

import (
	"errors"
	"fmt"
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

type Domain struct {
	typ   string
	value string
}

type location struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (s *location) equals(other *location) bool {
	if s == nil && other == nil {
		return true
	}

	return s != nil && other != nil &&
		s.Host == other.Host &&
		s.Port == other.Port
}

func (s *location) isFullyDefined() bool {
	return s.Host != "" && s.Port != 0
}

type ServiceConfig struct {
	Robots string `json:"robots"`
}

type Service struct {
	index      string
	nodeKey    string
	location   *location
	domain     string
	name       string
	status     *Status
	config     *ServiceConfig
	lastAccess *time.Time
}

type IoEtcdResolver struct {
	config          *Config
	watcher         *watcher
	domains         map[string]*Domain
	services        map[string]*ServiceCluster
	dest2ProxyCache map[string]*ServiceMux
	watchIndex      uint64
}

func NewEtcdResolver(c *Config) (*IoEtcdResolver, error) {
	domains := make(map[string]*Domain)
	services := make(map[string]*ServiceCluster)
	dest2ProxyCache := make(map[string]*ServiceMux)
	w, error := NewEtcdWatcher(c, domains, services)

	if error != nil {
		return nil, error
	}

	return &IoEtcdResolver{c, w, domains, services, dest2ProxyCache, 0}, nil
}

func (r *IoEtcdResolver) init() {
	r.watcher.init()
}

func (domain *Domain) equals(other *Domain) bool {
	if domain == nil && other == nil {
		return true
	}

	return domain != nil && other != nil &&
		domain.typ == other.typ && domain.value == other.value
}

func (config *ServiceConfig) equals(other *ServiceConfig) bool {
	if config == nil && other == nil {
		return true
	}

	return config != nil && other != nil &&
		config.Robots == other.Robots
}

func (service *Service) equals(other *Service) bool {
	if service == nil && other == nil {
		return true
	}

	return service != nil && other != nil &&
		service.location.equals(other.location) &&
		service.status.equals(other.status) &&
		service.config.equals(other.config)

}

func (r *IoEtcdResolver) resolve(domainName string) (http.Handler, error) {
	glog.V(5).Infof("Looking for domain : %s ", domainName)
	domain := r.domains[domainName]
	glog.V(5).Infof("Services:%s", r.services)
	if domain != nil {
		service := r.services[domain.value]
		if service == nil {
			glog.Errorf("The services map doesn't contain service with the domain value: %s", domain.value)
		}
		switch domain.typ {

		case SERVICE_DOMAINTYTPE:
			service, err := r.services[domain.value].Next()
			if err == nil && service.location.isFullyDefined() {
				addr := net.JoinHostPort(service.location.Host, strconv.Itoa(service.location.Port))
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

func (r *IoEtcdResolver) setLastAccessTime(service *Service) {

	interval := time.Duration(r.config.lastAccessInterval) * time.Second
	if service.lastAccess == nil || service.lastAccess.Add(interval).Before(time.Now()) {
		lastAccessKey := fmt.Sprintf("%s/lastAccess", service.nodeKey)

		client, error := r.config.getEtcdClient()
		if error != nil {
			glog.Errorf("Unable to get etcd client : %s ", error)
			return
		}

		now := time.Now()
		service.lastAccess = &now

		t := service.lastAccess.Format(TIME_FORMAT)
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
