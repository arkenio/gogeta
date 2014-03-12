package main

import (
  "flag"
)

type Config struct {
  port int
  etcdPrefix string
}


func parseConfig() *Config {
  config := &Config{}
  flag.IntVar(&config.port, "port", 7777, "Port to listen")
  flag.StringVar(&config.etcdPrefix, "prefix", "applications", "etcd prefix to get applications")
  flag.Parse()
  return config
}
