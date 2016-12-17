package main

import(
	"log"
	"strings"
	"github.com/cenkalti/backoff"
	"github.com/elgutierrez/go-mqtt-rpc"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	clientInstance := "client1123"
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
	opts := CreateClientOptions(instance, "iot.eclipse.org:1883", "", "")
  mqtt := NewMqttClient(opts)
  config := &mqttRpc.ClientConfig{
    Instance: instance,
    Prefix: "$rpc",
    QoS: 0,
  }
  rpcClient := mqttRpc.NewClient(mqtt, config)
  err := rpcClient.Connect(instanceTo, 5000)
	if err != nil {
		log.Fatal("RPC Cant connect", err)
		return err
	}

	var pingReply PingReply
	log.Println("Calling RemoteRpc.Ping to ", instanceTo)
	err = rpcClient.CallTimeout("ServerExposed.Ping", &PingArgs{}, &pingReply, 3000)
	if err != nil {
		log.Fatal("RPC Ping failed", err)
		return err
	}
	log.Println("Received pong ", pingReply)

 	return nil
}


func NewMqttClient(opts *mqtt.ClientOptions) mqtt.Client{
    client := mqtt.NewClient(opts)
    doConn := func() error{
        if token := client.Connect();  token.Wait() && token.Error() != nil{
          //TODO should retry connection with exponential backoff or maxRetry
          log.Println("Mqtt error ", token.Error())
          return token.Error()
        }
        return nil
    }
    err := backoff.Retry(doConn, backoff.NewExponentialBackOff())
    if err != nil {
        log.Fatal("MQTT cant connect. Commit Harakiri", err)
				panic(err)
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
	log.Println("Connection lost", reason)
}
