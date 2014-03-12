package main

import (
	"log"
)

const (
	progname = "etcd-reverse-proxy"
)

func main() {
	log.Printf("%s starting", progname)
	c := parseConfig()


	resolver := NewEtcdResolver(c)
	resolver.init()


	p := NewProxy(c, resolver)
	p.start()

}





