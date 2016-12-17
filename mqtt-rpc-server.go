package mqttRpc

import (
	"strings"
	"net/rpc"
  "crypto/sha1"
  "encoding/base64"
  "net/rpc/jsonrpc"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ServerConfig struct{
  Prefix       string
  QoS          byte
  Instance     string
}


type Server struct{
	Mqtt         mqtt.Client
  Config       *ServerConfig
}

func NewServer(client mqtt.Client, config *ServerConfig) *Server{
	return &Server{Mqtt: client, Config: config}
}

func (r *Server) Register(serverMethods interface{}) error{
  rpc.Register(serverMethods)
	return nil
}

func (r *Server) Start(errors chan error){
	conn := strings.Join([]string{r.Config.Prefix, r.Config.Instance, "conn"}, "/")
	connections := make(chan string, 100)

	token := r.Mqtt.Subscribe(conn, r.Config.QoS, func(mqttClient mqtt.Client, msg mqtt.Message) {
		connections <- string(msg.Payload())
	})

	if token.Wait() && token.Error() != nil {
		errors <- token.Error()
	}

	for{
		select {
		case instanceTo := <-connections:
			go r.ServeConnection(instanceTo, errors)
		}
	}
}

func (r *Server) ServeConnection(instanceTo string, errors chan error){
	hasher := sha1.New()
	hasher.Write([]byte(instanceTo))
	sessionId := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	write := strings.Join([]string{r.Config.Prefix, r.Config.Instance, instanceTo, sessionId, "res"}, "/")
	read := strings.Join([]string{r.Config.Prefix,  r.Config.Instance, instanceTo, sessionId, "req"}, "/")

	mqttConn := NewMqttRpcConn(r.Mqtt, read, write, r.Config.QoS)
	err := mqttConn.Open()
	if err != nil{
		errors <-  err
		return
	}

	connAck := strings.Join([]string{r.Config.Prefix, instanceTo, "conn", "ack"}, "/")
	token := r.Mqtt.Publish(connAck, r.Config.QoS, false, []byte(sessionId))
	if token.Wait() && token.Error() != nil {
		// return token.Error()
		mqttConn.Close()
		errors <-  token.Error()

		return
	}

	jsonrpc.ServeConn(mqttConn)
}
