package loadbalancer

import (
  "net/http"
  "net/http/httputil"
  "net/url"
  // "log"
  "reflect"
  "runtime"
  "github.com/aebrow4/unloadx-lb/util"
)

type strategy func([]*url.URL, []*lbutil.ServerHealth) *httputil.ReverseProxy

func RoundRobin(servers []*url.URL, _ []*lbutil.ServerHealth) *httputil.ReverseProxy {
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

func Health(servers []*url.URL, healths []*lbutil.ServerHealth) *httputil.ReverseProxy {
  var currServer int = 0
  director := func(req *http.Request) {
    // make sure healths are getting updated even though we passed them in
    // update currserver based on health
    currServer = lbutil.ChooseOnHealth(healths)
    server := servers[currServer]
    req.URL.Scheme = server.Scheme
    req.URL.Host = server.Host
    req.URL.Path = server.Path
  }

  return &httputil.ReverseProxy{Director: director}
}

// the LoadBalance function takes a loadbalancing strategy function,
// and an array of servers which it will pass to the strategy function
func LoadBalance(fn strategy, servers[]*url.URL, duration int) {
  serverHealths := make([]*lbutil.ServerHealth, 0)
  var serverHealthsPtrs []*lbutil.ServerHealth

  // if strategy is health, poll servers for their health
  if runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name() == "github.com/aebrow4/unloadx-lb/loadbalancer.Health" {
    serverHealthsPtrs = lbutil.GetHealth(servers, serverHealths[0:], serverHealthsPtrs[0:], duration);
  }

  proxy := fn(servers, serverHealthsPtrs)
  http.ListenAndServe(":9090", proxy)
}
