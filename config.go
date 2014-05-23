package main

import (
	"flag"
	"github.com/golang/glog"
)

type Config struct {
	port         int
	domainPrefix string
	envPrefix    string
	etcdAddress  string
	resolverType string
	templateDir  string
}

func parseConfig() *Config {
	config := &Config{}
	flag.IntVar(&config.port, "port", 7777, "Port to listen")
	flag.StringVar(&config.domainPrefix, "domainDir", "/nuxeo.io/domains", "etcd prefix to get domains")
	flag.StringVar(&config.envPrefix, "envDir", "/nuxeo.io/envs", "etcd prefix to get environments")
	flag.StringVar(&config.etcdAddress, "etcdAddress", "http://127.0.0.1:4001/", "etcd client host")
	flag.StringVar(&config.resolverType, "resolverType", "IoEtcd", "type of resolver (IoEtcd|Env|Dummy)")
	flag.StringVar(&config.templateDir, "templateDir","/var/www/gogeta", "Template directory")
	flag.Parse()

	glog.Infof("Dumping Configuration")
	glog.Infof("  listening port : %d", config.port)
	glog.Infof("  domainPrefix : %s", config.domainPrefix)
	glog.Infof("  envsPrefix : %s", config.envPrefix)
	glog.Infof("  etcdAddress : %s", config.etcdAddress)
	glog.Infof("  resolverType : %s", config.resolverType)
	glog.Infof("  templateDir: %s", config.templateDir)

	return config
}
