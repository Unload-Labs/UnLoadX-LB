package main

import (
  "net/http"
  "net/url"
  "encoding/json"
  "github.com/aebrow4/unloadx-lb/loadbalancer"
  "log"
)

func updateIpTables(w http.ResponseWriter, r *http.Request) {
  // declare some type to parse the POSTed JSON:
  // jsonBody holds the entire JSON object
  // message holds the objects of ip, port, and application info
  // siegeInput holds the extracted volume and testId to be relayed
  // to the siege service

  var jsonBody map[string]interface{}

  type message struct {
    Ip, Port, Application string
  }
  var serversStructs []message
  dec := json.NewDecoder(r.Body)
  dec.Decode(&jsonBody)
  type siegeInput struct {
    Volume float64
    TestId float64
  }

  siegeInit := siegeInput{
    Volume: jsonBody["volume"].(float64),
    TestId: jsonBody["testId"].(float64),
  }

  // a bunch of type assertions to index into and extract the data
  // which is a series of nested maps and interfaces
  servers := jsonBody["servers"].([]interface{})
  for _, v := range servers {
    ip := v.(map[string]interface{})["ip"].(string)
    port := v.(map[string]interface{})["port"].(string)
    application := v.(map[string]interface{})["application_type"].(string)
    server := message{
      Ip: ip,
      Port: port,
      Application: application,
    }
    serversStructs = append(serversStructs, server)
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
  log.Println(serversStructs)
  b, err := json.Marshal(siegeInit)
  if err != nil {
    log.Println("error marshaling json: ", err)
  }
  // start the load balancer, passing in the array
  log.Println("starting loadbalancer")
  loadbalancer.LoadBalance(loadbalancer.Health, serverPointers)
  log.Println("responding to post")
  w.Write(b)
}

// listens for a POST request of IPs and ports from the API server
func main() {
  http.HandleFunc("/iptables", updateIpTables)
  http.ListenAndServe(":9000", nil)
}
