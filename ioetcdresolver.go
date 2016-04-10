package main

import (
	"errors"
	"fmt"
	goarken "github.com/arkenio/goarken/model"
	"github.com/arkenio/goarken/storage"
	"github.com/golang/glog"
	"golang.org/x/net/context"
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
	arkenModel      *goarken.Model
	dest2ProxyCache map[string]*ServiceMux
	watchIndex      uint64
}

func NewEtcdResolver(c *Config) (*IoEtcdResolver, error) {

	dest2ProxyCache := make(map[string]*ServiceMux)

	client, err := c.getEtcdClient()

	if err != nil {
		return nil, err
	}

	persistenceDriver := storage.NewWatcher(client, c.servicePrefix, c.domainPrefix)

	arkenModel, err := goarken.NewArkenModel(nil, persistenceDriver)
	if err != nil {
		return nil, err
	}

	return &IoEtcdResolver{c, arkenModel, dest2ProxyCache, 0}, nil
}

func (r *IoEtcdResolver) init() {

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
	domain := r.arkenModel.Domains[domainName]
	glog.V(5).Infof("Services:%s", r.arkenModel.Services)
	if domain != nil {
		service := r.arkenModel.Services[domain.Value]
		if service == nil {
			glog.Errorf("The services map doesn't contain service with the domain value: %s", domain.Value)
		}
		switch domain.Typ {

		case SERVICE_DOMAINTYTPE:
			service, err := r.arkenModel.Services[domain.Value].Next()
			if err == nil && service.Location.IsFullyDefined() {
				addr := net.JoinHostPort(service.Location.Host, strconv.Itoa(service.Location.Port))
				uri := fmt.Sprintf("http://%s/", addr)
				r.setLastAccessTime(service)
				return r.getOrCreateProxyFor(service, uri), nil

			} else {
				return nil, err
			}
		case URI_DOMAINTYPE:
			return r.getOrCreateProxyFor(nil, domain.Value), nil
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
		//TODO: use model
		_, error = client.Set(context.Background(), lastAccessKey, t, nil)

		glog.V(5).Infof("Settign lastAccessKey to :%s", t)
		if error != nil {
			glog.Errorf("error :%s", error)
		}
	}

}

func (r *IoEtcdResolver) getOrCreateProxyFor(s *goarken.Service, uri string) http.Handler {
	if serviceMux, ok := r.dest2ProxyCache[uri]; !ok || !serviceMux.service.Equals(s) {
		glog.Infof("Creating a new muxer for : %s", s.Name)
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
