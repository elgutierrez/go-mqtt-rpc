package mqttRpc

import (
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttRpcConn struct{
	Mqtt 						mqtt.Client
	Inbound 			  chan []byte
	ReadTopic 			string
	WriteTopic			string
	QoS 						byte
	Instance 				string
}

func NewMqttRpcConn(mqtt mqtt.Client, read string, write string, qos byte ) *MqttRpcConn{
	return &MqttRpcConn{Mqtt: mqtt, Inbound: make(chan []byte, 100), ReadTopic: read, WriteTopic: write, QoS: qos}
}

func (m *MqttRpcConn) Open() error{
	  token := m.Mqtt.Subscribe(m.ReadTopic, m.QoS, func(client mqtt.Client, msg mqtt.Message) {
      m.Inbound <- msg.Payload()
    })
    if token.Wait() && token.Error() != nil {
      return token.Error()
    }

    return nil
}

func (m *MqttRpcConn) Read(p []byte) (int, error) {
	b, err := <-m.Inbound
	if !err {
		return 0, errors.New("Error reading from channel")
	}
	copy(p, b)
	return len(p), nil
}

func (m *MqttRpcConn) Write(p []byte) (int, error) {
	token := m.Mqtt.Publish(m.WriteTopic, m.QoS, false,  p)
	if token.Wait() && token.Error() != nil {
    return 0, token.Error()
  }

	return len(p), nil
}

func (m *MqttRpcConn) Close() error {
	token := m.Mqtt.Unsubscribe(m.ReadTopic)
	if token.Wait() && token.Error() != nil {
  	return token.Error()
  }

	return nil
}
