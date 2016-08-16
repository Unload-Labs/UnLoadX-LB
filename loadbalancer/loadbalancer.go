package loadbalancer

import (
  "net/http"
  "net/http/httputil"
  "net/url"
  "log"
  "reflect"
  "runtime"
  "time"
  "strings"
  "encoding/json"
  "github.com/aebrow4/unloadx-lb/util"
)

// TODO
// implement the other strategy functions

// define a type that all strategy functions will implement
// type strategy func([]*url.URL, []lbutil.ServerHealth) *httputil.ReverseProxy
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
func LoadBalance(fn strategy, servers[]*url.URL) {
  serverHealths := make([]*lbutil.ServerHealth, 0)
  var serverHealthsPtrs []*lbutil.ServerHealth
  // if strategy is health, poll servers for their health
  if runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name() == "github.com/aebrow4/unloadx-lb/loadbalancer.Health" {
    // initialize array of health structs to store health info in
    for _, server := range servers {
      currServerHealth := &lbutil.ServerHealth{
        Address: strings.Split(server.Host, ":")[0],
        Cpu: 0,
        Mem: 0,
      }
      serverHealths = append(serverHealths, currServerHealth)
    }


    for _, val := range serverHealths {
      serverHealthsPtrs = append(serverHealthsPtrs, val)
    }

    // send an HTTP request for each server in serverHealths
    // updating the serverHealths structs with the response info
    ticker := time.NewTicker(5 * time.Second)
    quit := make(chan struct{})
    go func() {
        for {
           select {
            case <- ticker.C:
                for _, serverHealth := range serverHealths[0:] {
                  r, _ := http.Get("http://" + serverHealth.Address + ":5000")
                  var jsonBody map[string]interface{}
                  dec := json.NewDecoder(r.Body)
                  dec.Decode(&jsonBody)
                  serverHealth.Cpu = jsonBody["cpu"].(float64)
                  serverHealth.Mem = jsonBody["memory"].(float64)
                }
            case <- quit:
                ticker.Stop()
                return
            }
        }
     }()
  }


  proxy := fn(servers, serverHealthsPtrs)
  http.ListenAndServe(":9090", proxy)
}
