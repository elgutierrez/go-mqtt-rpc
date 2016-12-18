# go-mqtt-rpc
Go MQTT RPC using native golang [jsonrpc](https://golang.org/pkg/net/rpc/jsonrpc/) and [eclipse MQTT](http://github.com/eclipse/paho.mqtt.golang) packages

# Examples

Server (see full example [here](https://github.com/elgutierrez/go-mqtt-rpc/blob/master/examples/server.go))
```go
package main

import (
	"log"
	"github.com/elgutierrez/go-mqtt-rpc"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	//Get mqtt client instance from somewhere
  mqtt := NewMqttClient()
  config := &mqttRpc.ServerConfig{
    Instance: "device1",
    Prefix: "$rpc",
    QoS: 0,
  }

  errors := make(chan error, 1)
  rpcServer := mqttRpc.NewServer(mqtt, config)
  //Register exposed methods
  rpcServer.Register(new(ServerExposed))
  rpcServer.Start(errors)
}


type PingArgs struct{
}

type PingReply struct{
	Done bool
}

//CLass that exposes the Ping method
type ServerExposed struct{
}

func (r *ServerExposed) Ping(args *PingArgs, reply *PingReply) error{
	log.Println("Received PING request; Sending PONG...")
	reply.Done = true
	return nil
}

```

Client (see full example [here](https://github.com/elgutierrez/go-mqtt-rpc/blob/master/examples/client.go))
```go

package main

import(
	"log"
	"github.com/elgutierrez/go-mqtt-rpc"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
  clientInstance := "client1"
  instanceTo := "device1"
  log.Println("Starting MQTT RPC client: instance", clientInstance, " to ", instanceTo)

  err := doPing(clientInstance, instanceTo)
  if err != nil {
    log.Fatal("Ping error", err)
  }
}

type PingArgs struct{
}

type PingReply struct{
		Done bool
}

func doPing(instance string, instanceTo string) error{
	//Get a connected MQTT client instance from somewhere
  mqtt := NewMqttClient()
  config := &mqttRpc.ClientConfig{
    Instance: instance,
    Prefix: "$rpc",
    QoS: 0,
  }
  rpcClient := mqttRpc.NewClient(mqtt, config)
  
  //Connect to the server instance with a 5000ms timeout
  err := rpcClient.Connect(instanceTo, 5000)
  if err != nil {
    log.Fatal("RPC Cant connect", err)
    return err
  }

  var pingReply PingReply
  log.Println("Calling RemoteRpc.Ping to ", instanceTo)
  //Call the method ServerExposed.Ping with 3000ms timeout
  err = rpcClient.CallTimeout("ServerExposed.Ping", &PingArgs{}, &pingReply, 3000)
  if err != nil {
    log.Fatal("RPC Ping failed", err)
    return err
  }
  log.Println("Received pong ", pingReply)

  return nil
}

```
