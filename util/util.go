package lbutil

import (
  "net/http"
)

type struct HealthInfo {
  cpu, mem float64
}

func GetHealth(server) HealthInfo {
  response, err := http.Get(server)
  if err != nil {
    log.Fatal(err)
  } else {
    defer response.Body.Close()

  }
}
