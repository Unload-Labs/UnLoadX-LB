package lbutil

import (
  "log"
)

type ServerHealth struct {
  Address string
  Cpu float64
  Mem float64
}

func ChooseOnHealth(healths []*ServerHealth) int {
  var lowestMemIdx int = 0
  var lowestCpuIdx int = 0
  h0 := *healths[0]
  log.Println("h0: ", h0)
  var lowestMem = float64(h0.Mem)
  var lowestCpu = float64(h0.Cpu)
  var cpuIdxBelowLimit []int

  for i, server := range healths {
    serverVal := *server
    if serverVal.Mem < lowestMem {
      lowestMemIdx = i
      lowestMem = serverVal.Mem
    }

    if serverVal.Cpu < lowestCpu {
      lowestCpuIdx = i
      lowestCpu = serverVal.Cpu
    }

    // ignore machines with high CPU usage
    if serverVal.Cpu < .7 {
      cpuIdxBelowLimit = append(cpuIdxBelowLimit, i)
    }
  }
  // if a server has lowest memory and CPU, choose it
  // else consider all servers with CPU < .7.  Choose based on lowest memory
  // else if all servers have CPU over .7, choose based on lowest memory
  if lowestMemIdx == lowestCpuIdx {
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
