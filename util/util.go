package lbutil

import (
  "log"
  "net/url"
  "net/http"
  "encoding/json"
  "strings"
  "time"
  "bytes"
  "os"
)

// data structure for storing server metrics
type ServerHealth struct {
  Address string
  Cpu float64
  Mem float64
  Avail bool
}

type ServerAvailability struct {
  Health bool
  Server bool
  Ip string
}

type Message struct {
  Ip string
  Port string
  Application string
}

type SiegeInput struct {
  Volume float64
  TestId float64
}

type NoSiege struct {
  Servers []ServerAvailability
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

// Poll a collection of servers for health information.  Optionally specify a duration
// to poll over.
// Return a pointer to the location where the collection of healths is stored
// Calls a function to calculate average healths over the duration
func GetHealth(servers[]*url.URL, serverHealths[]*ServerHealth, serverHealthsPtrs[]*ServerHealth, duration int, testId int) []*ServerHealth {
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
  numTicks := 0
  quit := make(chan struct{})
  go func() {
    for numTicks < duration{
     select {
      case <- ticker.C:
        for i, serverHealth := range serverHealths[0:] {
          serverHealth.Avail = true
          r, err := http.Get("http://" + serverHealth.Address + ":5000")
          if err != nil {
            log.Println("Failed to get health from ", serverHealth.Address)
            serverHealth.Avail = false;
            continue
          }
          var jsonBody map[string]interface{}
          dec := json.NewDecoder(r.Body)
          dec.Decode(&jsonBody)
          serverHealth.Cpu = jsonBody["cpu"].(float64)
          serverHealth.Mem = jsonBody["memory"].(float64)

          // update server to unavailable if status code doesn't begin with 2
          // send a request to the server rather than the health service, since
          // the health service may remain up even if the server goes down
          // this is arguably an expensive way of checking for server availability,
          // but better than pings which assume that the administrator has the ping
          // server turned on
          resp, err := http.Get("http://" + serverHealth.Address + ":" + serverPorts[i])
          if resp == nil {
            log.Println("no response to request")
            serverHealth.Avail = false;
          }
          if err != nil {
            log.Println("error in request to client server", err)
            serverHealth.Avail = false;
          }
        }
        numTicks++
      case <- quit:
        ticker.Stop()
        return
      }
    }
  }()
  if duration != 0 {
    CalcAvgHealth(duration, serverHealthsPtrs[0:], testId)
  }
  return serverHealthsPtrs[0:]
}

func CalcAvgHealth(duration int, serverHealthsPtrs[]*ServerHealth, testId int) {
  // for duration seconds, every second add the value of each metric to an analagous
  // struct field that will be used to calculate the average
  numServers := len(serverHealthsPtrs)
  avgHealths := make([]ServerHealth, numServers)

  numTicks := 0
  ticker := time.NewTicker(1 * time.Second)
  for numTicks < duration {
    for i, server := range serverHealthsPtrs {
      if numTicks == 0 {
        avgHealths[i].Address = server.Address
      }
      avgHealths[i].Cpu = avgHealths[i].Cpu + server.Cpu
      avgHealths[i].Mem = avgHealths[i].Mem + server.Mem
      log.Println("avg mem sum: ", avgHealths[i].Mem)
    }
    numTicks++
    <-ticker.C
    if numTicks == duration {
      ticker.Stop()
    }
  }
  for _, server := range avgHealths {
    server.Cpu = server.Cpu / float64(numTicks)
    server.Mem = server.Mem / float64(numTicks)
  }

  type AvgServerHealths struct {
    TestId int
    ServerHealths []ServerHealth
  }

  postData := AvgServerHealths{
    TestId: testId,
    ServerHealths: avgHealths,
  }

  MarshalledData, err := json.Marshal(postData)
  if err != nil {
    log.Println("err marshaling the boy: ", err)
  }

  r := bytes.NewReader(MarshalledData)
  if os.Getenv("ENV") == "dev" {
    resp, _ := http.Post("http://127.0.0.1:3000/api/serverhealth", "application/json", r)
    defer resp.Body.Close()
  } else {
    resp, _ := http.Post("http://52.9.136.53:3000/api/serverhealth", "application/json", r)
    defer resp.Body.Close()
  }
  log.Println("sent post with health")
  return
}

// takes in a server IP and port and returns T/F if it can be reached
func CheckServerAvail(server Message) bool {
  timeout := time.Duration(3 * time.Second)
  client := http.Client{
      Timeout: timeout,
  }
  res, err := client.Get("http://" + server.Ip + ":" + server.Port)
  if err != nil || res == nil {
    log.Println("Server could not be contacted ", server.Ip)
    return false
  }
  log.Println("Server contacted")
  return true
}

// takes in a server IP and returns T/F if the health service can be reached
func CheckServerHealthAvail(server Message) bool {
  timeout := time.Duration(3 * time.Second)
  client := http.Client{
      Timeout: timeout,
  }
  _, err := client.Get("http://" + server.Ip + ":5000")
  if err != nil {
    log.Println("Health not available for ", server.Ip)
    return false
  }
  log.Println("Got health")
  return true
}
