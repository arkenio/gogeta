package main

import (
  "fmt"
  "net/http"
  "net/http/httputil"
  "net/url"
  "log"
  "regexp"
)

type proxy struct {
  config *Config
}


func NewProxy(c *Config) *proxy {
  p := &proxy{}
  p.config = c
  return p
}

func (p *proxy) start() {
  log.Printf("Listening on port %d",p.config.port)
  http.HandleFunc("/", OnRequest)
  http.ListenAndServe(fmt.Sprintf(":%d", p.config.port), nil)

}



func OnRequest(w http.ResponseWriter, r *http.Request) {

  server, found := matchingServerOf(r.Host, r.URL.String())

  if found {
    server.ServeHTTP(w, r)
    return
  }

  http.NotFound(w, r)
}


func ReverseProxyServer(uri string) http.Handler {
  dest, _ := url.Parse(addProtocol(uri))
  return httputil.NewSingleHostReverseProxy(dest)
}

func addProtocol(url string) string {
  if matches, _ := regexp.MatchString("^\\w+://", url); !matches {
    return fmt.Sprintf("http://%s", url)
  }

  return url
}

