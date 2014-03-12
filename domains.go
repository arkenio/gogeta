package main



import (
  "fmt"
  "net/http"
  "strings"
)

type Domains map[string]EndPoints

type EndPoints map[string]EndPoint

type EndPoint struct {
  url string
  server http.Handler
}

var (

  domains Domains
)

func matchingHandlerOf(url, hostname string, handlers EndPoints) (result http.Handler, found bool) {

  if handlers == nil {
    return nil, false
  }

  for pattern, handler := range handlers {

    if pattern == "*" {
      continue
    }

    if len(url) >= len(pattern) && url[0:len(pattern)] == pattern {
      found = true
      result = http.StripPrefix(pattern, handler.server)
    }
  }

  if handler, hasDefaultHandler := handlers["*"]; !found && hasDefaultHandler {
    result = handler.server
    found = true
  }

  return result, found
}

func matchingServerOf(host, url string) (result http.Handler, found bool) {

  hostname := hostnameOf(host)
  wildcard := wildcardOf(hostname)

  result, found = matchingHandlerOf(url, hostname, domains[hostname])

  if !found {
    if _, hasWildcard := domains[wildcard]; hasWildcard {
      result, found = matchingHandlerOf(url, hostname, domains[wildcard])
    } else {
    }
  }

  if wildcardSite, hasWildcardSite := domains["*"]; !found && hasWildcardSite {
    result, found = matchingHandlerOf(url, hostname, wildcardSite)
  } else if !found {
  } else {
  }

  return result, found
}

func hostnameOf(host string) string {
  hostname := strings.Split(host, ":")[0]

  if len(hostname) > 4 && hostname[0:4] == "www." {
    hostname = hostname[4:]
  }

  return hostname
}

func wildcardOf(hostname string) string {
  parts := strings.Split(hostname, ".")

  if len(parts) < 3 {
    return fmt.Sprintf("*.%s", hostname)
  }

  parts[0] = "*"
  return strings.Join(parts, ".")

}