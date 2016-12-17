package main

import (
	"log"
	"strings"
  "github.com/cenkalti/backoff"
	"github.com/elgutierrez/go-mqtt-rpc"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	instance := "device1"
	log.Println("Starting RPC server for ", instance)

  opts := CreateClientOptions(instance, "iot.eclipse.org:1883", "", "")
  mqtt := NewMqttClient(opts)
  config := &mqttRpc.ServerConfig{
    Instance: instance,
    Prefix: "$rpc",
    QoS: 0,
  }

	errors := make(chan error, 1)
  rpcServer := mqttRpc.NewServer(mqtt, config)
  rpcServer.Register(new(ServerExposed))
  rpcServer.Start(errors)
}


func NewMqttClient(opts *mqtt.ClientOptions) mqtt.Client{
    client := mqtt.NewClient(opts)
    doConn := func() error{
        if token := client.Connect();  token.Wait() && token.Error() != nil{
            //TODO should retry connection with exponential backoff or maxRetry
            log.Println("MQTT Error: ", token.Error())
            return token.Error()
        }
        return nil
    }

    err := backoff.Retry(doConn, backoff.NewExponentialBackOff())
    if err != nil {
        log.Fatal("MQTT Cant connect. Commit Harakiri")
    }

    return client
}

func CreateClientOptions(instance string, broker string, username string, password string) *mqtt.ClientOptions {
    opts := mqtt.NewClientOptions().AddBroker(strings.Join([]string{"tcp://", broker}, "")).SetClientID(instance).
                                SetCleanSession(false).SetUsername(username).SetPassword(password).
                                SetConnectionLostHandler(ConnectionLostHandler)
    return opts
}

func ConnectionLostHandler(client mqtt.Client, reason error){
	log.Println("MQTT Connection lost: ", reason)
}

type PingArgs struct{
}

type PingReply struct{
	Done bool
}


type ServerExposed struct{
}

func (r *ServerExposed) Ping(args *PingArgs, reply *PingReply) error{
	//Empty method only to check if the remote instance is up and answering requests
	log.Println("Received PING request; Sending PONG...")
	reply.Done = true
	return nil
}
