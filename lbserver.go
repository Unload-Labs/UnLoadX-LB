package main

import (
  "net/http"
  "net/url"
  "encoding/json"
  "github.com/aebrow4/unloadx-lb/loadbalancer"
  "bytes"
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
  // send siegeInit to the siege service
  b, err := json.Marshal(siegeInit)
  if err != nil {
    log.Println("error marshaling json: ", err)
  }
  // convert the []byte to a buffer that http.POST can use
  var buf bytes.Buffer
  // log.Println("posting to siege")
  buf.Write(b)
  // _, er := http.Post("http://52.9.136.53:4000/siege", "application/json; charset=utf-8", &buf)
  //
  // if er != nil {
  //   log.Println(err)
  // }

  // start the load balancer, passing in the array
  // this works, but for some reason it causes the above call to
  // WriteHeader to be ignored
  log.Println("starting loadbalancer")
  loadbalancer.LoadBalance(loadbalancer.RoundRobin, serverPointers)
  // i think these lines of code are not being reached because starting the
  // server locks up the thread?...
  log.Println("responding to post")
  // w.WriteHeader(http.StatusOK)
  w.Write(b)
  w.WriteHeader(http.StatusOK)
}


// listens for a POST request of IPs and ports from the API server
func main() {
  http.HandleFunc("/iptables", updateIpTables)
  http.ListenAndServe(":9000", nil)
}
