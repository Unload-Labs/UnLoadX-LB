package lbutil

import (
  "log"
  "net/url"
  "net/http"
  "encoding/json"
  "strings"
  "time"
)

// data structure for storing server metrics
type ServerHealth struct {
  Address string
  Cpu float64
  Mem float64
  Avail bool
}

// a linear search method for unsorted int slices and a target value
func contains(collection []int, target int) bool {
    for _, val := range collection {
        if val == target {
            return true
        }
    }
    return false
}

// Given a slice of structs with server health and status, return the index
// of the healthiest server (lowest mem and CPU usage). Ignore servers that
// are unavailable (did not return 2XX status code on last ping)
func ChooseOnHealth(healths []*ServerHealth) int {
  var lowestMemIdx int = 0
  var lowestCpuIdx int = 0
  h0 := *healths[0]
  var lowestMem = float64(h0.Mem)
  var lowestCpu = float64(h0.Cpu)
  var cpuIdxBelowLimit []int

  // filter out any unavailable servers
  var availServerIdx []int
  for i, server := range healths {
    serverVal := *server
    if (serverVal.Avail == true) {
      availServerIdx = append(availServerIdx, i)
    }
  }

  // find server with lowest usage
  for i, server := range healths {
    serverVal := *server
    if serverVal.Mem < lowestMem && serverVal.Avail {
      lowestMemIdx = i
      lowestMem = serverVal.Mem
    }
    if serverVal.Cpu < lowestCpu && serverVal.Avail {
      lowestCpuIdx = i
      lowestCpu = serverVal.Cpu
    }

    // ignore machines with high CPU usage
    if serverVal.Cpu < .7 && serverVal.Avail {
      cpuIdxBelowLimit = append(cpuIdxBelowLimit, i)
    }
  }

  // if a server has lowest memory and CPU, choose it
  // else consider all servers with CPU < .7.  Choose based on lowest memory
  // else if all servers have CPU over .7, choose based on lowest memory
  isLowestAvail := contains(availServerIdx, lowestMemIdx)
  if lowestMemIdx == lowestCpuIdx && isLowestAvail {
    log.Println("avail indices: ", availServerIdx)
    log.Println("choosing idx: ", lowestMemIdx)
    return lowestMemIdx
  } else if len(cpuIdxBelowLimit) < len(healths) {
    lowestMem = 1000
    for _, idx := range cpuIdxBelowLimit {
      healthVal := *healths[idx]
      if healthVal.Mem < lowestMem {
        lowestMemIdx = idx
        lowestMem = healthVal.Mem
        }
    }
  }
  return lowestMemIdx
}

func GetHealth(servers[]*url.URL, serverHealths[]*ServerHealth, serverHealthsPtrs[]*ServerHealth) []*ServerHealth {
  var serverPorts []string

  for _, server := range servers {
    currServerHealth := &ServerHealth{
      Address: strings.Split(server.Host, ":")[0],
      Cpu: 0,
      Mem: 0,
    }
    serverHealths = append(serverHealths, currServerHealth)
    serverPorts = append(serverPorts, strings.Split(server.Host, ":")[1])
  }

  for _, val := range serverHealths {
    serverHealthsPtrs = append(serverHealthsPtrs, val)
  }

  // send an HTTP request for each server in serverHealths
  // updating the serverHealths structs with the response info
  ticker := time.NewTicker(1 * time.Second)
  quit := make(chan struct{})
  go func() {
      for {
         select {
          case <- ticker.C:
            for i, serverHealth := range serverHealths[0:] {
              r, _ := http.Get("http://" + serverHealth.Address + ":5000")
              var jsonBody map[string]interface{}
              dec := json.NewDecoder(r.Body)
              dec.Decode(&jsonBody)
              serverHealth.Cpu = jsonBody["cpu"].(float64)
              serverHealth.Mem = jsonBody["memory"].(float64)
              serverHealth.Avail = true

              // update server to unavailable if status code doesn't begin with 2
              // send a request to the server rather than the health service, since
              // the health service may remain up even if the server goes down
              // this is arguably an expensive way of checking for server availability,
              // but better than pings which assume that the administrator has the ping
              // server turned on
              resp, err := http.Get("http://" + serverHealth.Address + ":" + serverPorts[i])
              if resp == nil || err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
                log.Println("error ocurred in get request")
                serverHealth.Avail = false
              }
            }
          case <- quit:
            ticker.Stop()
            return
          }
      }
   }()
  return serverHealthsPtrs[0:]
}
