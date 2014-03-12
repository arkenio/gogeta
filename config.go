package main

import (
  "flag"
)

type Config struct {
  port int
  domainPrefix string
  envPrefix string
  etcdAdress string
}


func parseConfig() *Config {
  config := &Config{}
  flag.IntVar(&config.port, "port", 7777, "Port to listen")
  flag.StringVar(&config.domainPrefix, "domainDir", "domains", "etcd prefix to get domains")
  flag.StringVar(&config.envPrefix, "envDir", "environments", "etcd prefix to get environments")
  flag.StringVar(&config.etcdAdress, "etcdAdress", "http://127.0.0.1:4001/", "etcd client host")
  flag.Parse()
  return config
}
