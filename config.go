package main

import (
	"flag"
	"github.com/golang/glog"
)

type Config struct {
	port         int
	domainPrefix string
	servicePrefix    string
	etcdAddress  string
	resolverType string
	templateDir  string
}

func parseConfig() *Config {
	config := &Config{}
	flag.IntVar(&config.port, "port", 7777, "Port to listen")
	flag.StringVar(&config.domainPrefix, "domainDir", "/domains", "etcd prefix to get domains")
	flag.StringVar(&config.servicePrefix, "serviceDir", "/services", "etcd prefix to get services")
	flag.StringVar(&config.etcdAddress, "etcdAddress", "http://127.0.0.1:4001/", "etcd client host")
	flag.StringVar(&config.resolverType, "resolverType", "IoEtcd", "type of resolver (IoEtcd|Env|Dummy)")
	flag.StringVar(&config.templateDir, "templateDir","./templates", "Template directory")
	flag.Parse()

	glog.Infof("Dumping Configuration")
	glog.Infof("  listening port : %d", config.port)
	glog.Infof("  domainPrefix : %s", config.domainPrefix)
	glog.Infof("  servicesPrefix : %s", config.servicePrefix)
	glog.Infof("  etcdAddress : %s", config.etcdAddress)
	glog.Infof("  resolverType : %s", config.resolverType)
	glog.Infof("  templateDir: %s", config.templateDir)

	return config
}
