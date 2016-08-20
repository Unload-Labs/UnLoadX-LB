package main

import (
  "log"
  "net/http"
  "net/url"
  "encoding/json"
  "github.com/aebrow4/unloadx-lb/loadbalancer"
  "github.com/aebrow4/unloadx-lb/util"
)

func updateIpTables(w http.ResponseWriter, r *http.Request) {

  var jsonBody map[string]interface{}
  dec := json.NewDecoder(r.Body)
  dec.Decode(&jsonBody)

  var serversStructs []lbutil.Message
  siegeInit := lbutil.SiegeInput{
    Volume: jsonBody["volume"].(float64),
    TestId: jsonBody["testId"].(float64),
  }
  duration := int(siegeInit.Volume)
  testId := int(siegeInit.TestId)

  // store unavailable servers in this struct
  siegeVoid := lbutil.NoSiege{}

  // a bunch of type assertions to index into and extract the data
  // which is a series of nested maps and interfaces
  servers := jsonBody["servers"].([]interface{})
  for _, v := range servers {
    ip := v.(map[string]interface{})["ip"].(string)
    port := v.(map[string]interface{})["port"].(string)
    application := v.(map[string]interface{})["application_type"].(string)
    server := lbutil.Message{
      Ip: ip,
      Port: port,
      Application: application,
    }
    serversStructs = append(serversStructs, server)

    // test for server availability before proceeding
    healthAvail := lbutil.CheckServerHealthAvail(server)
    serverAvail := lbutil.CheckServerAvail(server)
    if healthAvail == false || serverAvail == false {
      serverUnavail := lbutil.ServerAvailability{
        Health: healthAvail,
        Server: serverAvail,
        Ip: ip,
      }
      siegeVoid.Servers = append(siegeVoid.Servers, serverUnavail)
    }
  }

  // if servers are unavailable in any way, respond to the client with
  // that information, do not start LB and do not start siege
  if len(siegeVoid.Servers) > 0 {
    log.Println("detected unavail servers", siegeVoid)
    buf, _ := json.Marshal(siegeVoid)
    w.Write(buf)
    return
  }

  serverURLs := make([]url.URL, 0)
  serverPointers := make([]*url.URL, 0)
  for _, element := range serversStructs {
    serverURL := url.URL{
      Scheme: "http",
      Host: element.Ip + ":" + element.Port,
    }
    serverURLs = append(serverURLs, serverURL)
    serverPointers = append(serverPointers, &serverURL)
  }
  buf, _ := json.Marshal(siegeInit)
  // start the load balancer, passing in the array
  log.Println("starting loadbalancer")
  loadbalancer.LoadBalance(loadbalancer.Health, serverPointers, duration, testId)
  w.Write(buf)
}

func respondToPing(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(200)
}

// listens for a POST request of IPs and ports from the API server
func main() {
  http.HandleFunc("/iptables", updateIpTables)
  http.HandleFunc("/", respondToPing)
  http.ListenAndServe(":9000", nil)
}
