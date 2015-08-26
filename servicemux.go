package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ServiceMux struct {
	service       *Service
	serveMux      *http.ServeMux
	internalProxy *httputil.ReverseProxy
}

func NewServiceMux(c *Config, s *Service, proxyDest string) *ServiceMux {

	dest, _ := url.Parse(proxyDest)
	r := &ServiceMux{service: s, internalProxy: NewSingleHostReverseProxy(c, dest)}
	r.init()
	return r
}

func (p *ServiceMux) init() {
	p.serveMux = http.NewServeMux()
	if p.service != nil {
		p.serveMux.HandleFunc("/robots.txt", p.robots)
	}
	p.serveMux.Handle("/", p.internalProxy)
}

func (p *ServiceMux) robots(w http.ResponseWriter, r *http.Request) {
	if p.service.config.Robots != "" {
		fmt.Fprint(w, p.service.config.Robots)
	} else {
		fmt.Fprint(w, "User-agent: *\nDisallow: /")
	}

}

func (p *ServiceMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.serveMux.ServeHTTP(rw, req)
}

func NewSingleHostReverseProxy(config *Config, target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}

		// FIXME : Nuxeo hack to be able to add a virtual-host header param.
		// To be removed as soon as possible
		if config.UrlHeaderParam != "" {
			scheme := req.Header.Get("x-forwarded-proto")
			host := req.Host
			port := req.Header.Get("x-forwarded-port")
			url := ""
			if ("https" == scheme && "443" == port) || ("http" == scheme && "80" == port) {
				url = fmt.Sprintf("%s://%s/", scheme, host)
			} else {
				url = fmt.Sprintf("%s://%s:%s/", scheme, host, port)
			}
			req.Header.Add(config.UrlHeaderParam, url)
		}
	}
	return &httputil.ReverseProxy{Director: director}
}
