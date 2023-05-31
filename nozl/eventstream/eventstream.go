package eventstream

import (
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/nats-io/nats.go"

	"github.com/nats-io/nats-server/v2/nozl/shared"
)

type (
	eventstream struct {
		Conn   *nats.Conn
		EnCon  *nats.EncodedConn
		Stream nats.JetStreamContext
		// ServerURL string `env:"NATS_SERVER_URL" env-default:"nats://localhost:4222"`
		ServerURL string
	}
)

var (
	Eventstream        = &eventstream{}
	MessageFilterAllow chan *MessageFilterStatus
	ServiceResponse    chan []byte
)

// func (n *eventstream) ReadEnv() {
// 	url, ok := os.LookupEnv("NATS_SERVER_URL")

// 	if ok {
// 		n.ServerURL = url
// 	} else {
// 		shared.Logger.Warn("Environment variable `NATS_SERVERS_URL` not found. Using default value")
// 	}

// }

func (n *eventstream) SetupUrl(port int) {
	n.ServerURL = fmt.Sprintf("nats://localhost:%d", port)
}

func (n *eventstream) InitializeNats() {
	retryNats := func() error {
		nc, err := nats.Connect(n.ServerURL, nats.MaxReconnects(4), nats.ReconnectWait(4*time.Second))

		if err != nil {
			shared.Logger.Error("Unable to connect with Nats server due to " + err.Error())
		}

		ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)

		if err != nil {
			shared.Logger.Error("Unable to create encoded connection: " + err.Error())
			return err
		}

		js, err := nc.JetStream()

		if err != nil {
			shared.Logger.Error("Unable to create JetStream connection: " + err.Error())
			return err
		}

		n.Conn = nc
		n.EnCon = ec
		n.Stream = js
		MessageFilterAllow = make(chan *MessageFilterStatus)
		ServiceResponse = make(chan []byte)

		return nil
	}

	if err := retry.Do(retryNats, retry.Attempts(5), retry.Delay(1*time.Second)); err != nil {
		// shared.Logger.Error("Unable to establish connection with nats server")
		panic("Unable to establish connection with nats server")
	}
}

func (n *eventstream) PublishEncodedMessage(subject string, msg *Message) {
	if err := n.EnCon.Publish(subject, msg); err != nil {
		log := fmt.Sprintf("Unable publish message on subject `%s` due to %s exception", subject, err.Error())
		shared.Logger.Error(log)
	}

	shared.Logger.Info("Message published successfully")
}

func (n *eventstream) PublishMessage(subject string, msg string) {
	if err := n.Conn.Publish(subject, []byte(msg)); err != nil {
		log := fmt.Sprintf("Unable publish message '%s' on subject `%s` due to %s exception", subject, msg, err.Error())
		shared.Logger.Error(log)
	}

	shared.Logger.Info("Message '" + msg + "' published successfully")
}

func (n *eventstream) CreateKeyValStore(bucket string, description string, replicationFactor int) (nats.KeyValue, error) {
	keyval, err := n.Stream.CreateKeyValue(&nats.KeyValueConfig{
		Bucket:      bucket,
		Description: description,
		Replicas:    replicationFactor,
	})
	if err != nil {
		shared.Logger.Error("Unable to create Key Value Store " + err.Error())
	}

	return keyval, err
}

func (n *eventstream) RetreiveKeyValStore(bucket string) (nats.KeyValue, error) {
	keyval, err := n.Stream.KeyValue(bucket)
	if err != nil {
		shared.Logger.Error("Unable to retreive Key Value Store " + err.Error())
	}

	return keyval, err
}

func (n *eventstream) CheckStreamBucketExists(bucket string) (bool, error) {
	stream, err := n.Stream.StreamInfo(bucket)
	if stream != nil {
		return true, err
	}

	return false, err
}

func (n *eventstream) TestNatsClient() {
	defer n.Conn.Close()

	c := make(chan string)

	_, err := n.Conn.Subscribe("hello", func(msg *nats.Msg) {
		shared.Logger.Info("Messaged received from subject `hello`: " + string(msg.Data))
		c <- string(msg.Data)
	})

	if err != nil {
		shared.Logger.Error(err.Error())
	}

	err = n.Conn.Publish("hello", []byte("Hello, NATS!"))

	if err != nil {
		shared.Logger.Error(err.Error())
	}

	res := <-c

	shared.Logger.Info("Message received from channel: " + res)

	close(c)
}
