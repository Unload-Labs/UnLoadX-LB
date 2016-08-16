package lbutil

import (
  "log"
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
  h1 := *healths[1]
  log.Println("h0: ", h0)
  log.Println("h1: ", h1)
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
  log.Println("got here: ", availServerIdx)

  // if a server has lowest memory and CPU, choose it
  // else consider all servers with CPU < .7.  Choose based on lowest memory
  // else if all servers have CPU over .7, choose based on lowest memory
  isLowestAvail := contains(availServerIdx, lowestMemIdx)
  log.Println("lowestmemidx: ", lowestMemIdx)
  log.Println(isLowestAvail)
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
  log.Println("avail indices: ", availServerIdx)
  log.Println("choosing idx: ", lowestMemIdx)
  return lowestMemIdx
}
