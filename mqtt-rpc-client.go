package mqttRpc

import(
	"time"
	"errors"
	"strings"
	"net/rpc"
	"net/rpc/jsonrpc"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ClientConfig struct{
	Instance 	string
	Prefix 	  string
	QoS  		  byte
}

type Client struct{
	Mqtt      mqtt.Client
	Client 		*rpc.Client
	Config 		*ClientConfig
}

func NewClient(client mqtt.Client, config *ClientConfig) *Client{
  return &Client{Mqtt: client, Config: config}
}

func (s *Client) Connect(instanceTo string, timeout int) error{
	ack := make(chan string, 1)
	connAck := strings.Join([]string{s.Config.Prefix, s.Config.Instance, "conn", "ack"}, "/")
	token := s.Mqtt.Subscribe(connAck, s.Config.QoS, func(mqttClient mqtt.Client, msg mqtt.Message) {
		ack <- string(msg.Payload())
	})
	if token.Wait() && token.Error() != nil {
  	return token.Error()
  }
	conn := strings.Join([]string{s.Config.Prefix, instanceTo, "conn"}, "/")
	token = s.Mqtt.Publish(conn, s.Config.QoS, false, []byte(s.Config.Instance))
	if token.Wait() && token.Error() != nil {
    return token.Error()
  }

	select {
		case sessionId := <-ack:
				token := s.Mqtt.Unsubscribe(connAck)
				if token.Wait() && token.Error() != nil {
			  	return token.Error()
			  }
				write := strings.Join([]string{s.Config.Prefix, instanceTo, s.Config.Instance, sessionId, "req"}, "/")
				read := strings.Join([]string{s.Config.Prefix, instanceTo, s.Config.Instance, sessionId, "res"}, "/")

				mqttConn := NewMqttRpcConn(s.Mqtt, read, write, s.Config.QoS)
				err := mqttConn.Open()
				if err != nil {
					return err
				}

				s.Client = jsonrpc.NewClient(mqttConn)

		    return nil
		case <-time.After(time.Millisecond * time.Duration(timeout)):
		    return errors.New("RPC client timeout")
	}
}

func (s *Client) Call(method string, params interface{}, reply interface{}) error{
	return s.Client.Call(method, params, reply)
}

func (s *Client) CallTimeout(method string, params interface{}, reply interface{}, timeout int) error{
	c := make(chan error, 1)
	go func() {
    c <- s.Client.Call(method, params, reply)
  }()
	select {
	  case err := <-c:
	    // use err and result
	    return err
	  case <-time.After(time.Millisecond * time.Duration(timeout)):
	    // call timed out
	    return errors.New("RPC timeout")
	}
}

func (s *Client) Close(){
	s.Client.Close()
}
