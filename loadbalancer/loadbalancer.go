package loadbalancer

import (
  "net/http"
  "net/http/httputil"
  "net/url"
  "log"
)

// TODO
// implement the other strategy functions

// define a type that all strategy functions will implement
type strategy func([]*url.URL) *httputil.ReverseProxy

func RoundRobin(servers []*url.URL) *httputil.ReverseProxy {
  var currServer int = 0
  director := func(req *http.Request) {
    server := servers[currServer]
    req.URL.Scheme = server.Scheme
    req.URL.Host = server.Host
    req.URL.Path = server.Path

    currServer++
    if currServer > len(servers) - 1 {
      currServer = 0
    }
  }
  return &httputil.ReverseProxy{Director: director}
}


func Health(servers []*url.URL) *httputil.ReverseProxy {

}

// the LoadBalance function takes a loadbalancing strategy function,
// and an array of servers which it will pass to the strategy function

func LoadBalance(fn strategy, servers[]*url.URL) {
  proxy := fn(servers)
  http.ListenAndServe(":9090", proxy)
}
