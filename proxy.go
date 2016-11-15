package main

import (
	"fmt"
	goarken "github.com/arkenio/arken/goarken/model"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

type domainResolver interface {
	resolve(domain string) (http.Handler, error)
	init()
}

type proxy struct {
	config         *Config
	domainResolver domainResolver
}

func NewProxy(c *Config, resolver domainResolver) *proxy {
	return &proxy{c, resolver}
}

type proxyHandler func(http.ResponseWriter, *http.Request) (*Config, error)

func (ph proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if r := recover(); r != nil {
			http.Error(w, "An error occured serving request", 500)
			glog.Errorf("Recovered from error : %s", r)
		}
	}()

	if c, err := ph(w, r); err != nil {
		ph.OnError(w, r, err, c)
	}
}

func (ph proxyHandler) OnError(w http.ResponseWriter, r *http.Request, error error, c *Config) {
	if stError, ok := error.(goarken.StatusError); ok {
		sp := &StatusPage{c, stError}
		// Check if status is passivated -> setting expected state = started
		if sp.error.ComputedStatus == goarken.PASSIVATED_STATUS {
			reactivate(sp, c)
		}
		sp.serve(w, r)
	} else {
		sp := &StatusPage{c, goarken.StatusError{"notfound", nil}}
		sp.serve(w, r)
	}
}

func (p *proxy) start() {
	glog.Infof("Listening on port %d", p.config.port)
	if p.config.templateDir != "" {
		http.Handle("/__static__/", http.FileServer(http.Dir(p.config.templateDir)))
	} else {
		http.Handle("/__static__/", http.FileServer(FS(false)))
	}
	http.Handle("/", proxyHandler(p.proxy))
	glog.Fatalf("%s", http.ListenAndServe(fmt.Sprintf(":%d", p.config.port), nil))

}

func (p *proxy) proxy(w http.ResponseWriter, r *http.Request) (*Config, error) {

	if p.config.forceFwSsl && "https" != r.Header.Get("x-forwarded-proto") {

		http.Redirect(w, r, fmt.Sprintf("https://%s%s", hostnameOf(r.Host), r.URL.String()), http.StatusMovedPermanently)
		return p.config, nil
	}

	host := hostnameOf(r.Host)
	if server, err := p.domainResolver.resolve(host); err != nil {
		return p.config, err
	} else {
		server.ServeHTTP(w, r)
		return p.config, nil
	}
}

func hostnameOf(host string) string {
	hostname := strings.Split(host, ":")[0]

	if len(hostname) > 4 && hostname[0:4] == "www." {
		hostname = hostname[4:]
	}

	return hostname
}

func reactivate(sp *StatusPage, c *Config) {
	client, _ := c.getEtcdClient()
	//TODO use model
	_, error := client.Set(context.Background(), c.servicePrefix+"/"+sp.error.Status.Service.Name+"/"+sp.error.Status.Service.Index+"/status/expected", goarken.STARTED_STATUS, nil)
	if error != nil {
		glog.Errorf("Fail: setting expected state = 'started' for instance %s. Error:%s", sp.error.Status.Service.Name, error)
	}
	glog.Infof("Instance %s is ready for re-activation", sp.error.Status.Service.Name)
}
