package main

import (
	"github.com/golang/glog"
)

const (
	progname = "gogeta"
)

func getResolver(c *Config) domainResolver {
	switch c.resolverType {
	case "Dummy":
		return &DummyResolver{}
	case "Env":
		return NewEnvResolver(c)
	default:
		return NewEtcdResolver(c)
	}
}

func main() {

	glog.Infof("%s starting", progname)

	c := parseConfig()

	resolver := getResolver(c)
	resolver.init()

	p := NewProxy(c, resolver)
	p.start()

}
