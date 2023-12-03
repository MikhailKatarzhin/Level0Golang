package stan

import (
	"fmt"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
	"go.uber.org/zap"

	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker/nats"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker/submanager"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/logger"
)

type Client struct {
	nats       *nats.Client
	subManager *submanager.Manager
	stanConn   stan.Conn
	log        *zap.Logger
}

func New(config broker.NATSConfig) *Client {
	natsConn := nats.New(config)

	return NewWithNats(natsConn)
}

func NewWithNats(nats *nats.Client) *Client {
	return &Client{
		nats:       nats,
		subManager: submanager.New(),
		log:        logger.MustGetLogger(),
	}
}

func (c *Client) Connect(clusterID string, clientID string) error {
	if !c.nats.IsConnected() {
		if err := c.nats.Connect(); err != nil {
			return broker.ErrCanNotConnect("can not connect to nats", err)
		}
	}

	var err error

	clientID = fmt.Sprint(clientID, "-", gonanoid.Must(5))

	c.stanConn, err = c.nats.StanConn(clusterID, clientID)
	if err != nil {
		return broker.ErrCanNotConnect("can not acquire stan connection", err)
	}

	return nil
}

func (c *Client) Disconnect() error {
	if err := c.nats.Disconnect(); err != nil {
		return err
	}

	if err := c.subManager.Drain(); err != nil {
		return err
	}

	return nil
}

func (c *Client) SetReconnectHandler(handler func()) {
	c.nats.SetReconnectHandler(handler)
}

func (c *Client) Subscribe(subject, consumer string) (broker.Subscribe, error) {
	if !c.nats.IsConnected() {
		return broker.Subscribe{}, broker.ErrBrokerNotConnected
	}

	subCh := make(broker.SubCh, c.nats.Cfg.SubscriptionQueueLen)

	var opts []stan.SubscriptionOption
	if consumer != "" {
		opts = append(opts, stan.DurableName(consumer))
	}

	opts = append(opts, stan.StartAt(pb.StartPosition_NewOnly))

	subscribe, err := c.stanConn.Subscribe(subject, func(msg *stan.Msg) {
		subCh <- broker.Message{
			Subject:           subject,
			ReceiveTimeMicros: time.Now().UnixMicro(),
			Body:              msg.Data,
			AckCallback: func() error {
				return msg.Ack()
			},
		}
	}, opts...)
	if err != nil {
		return broker.Subscribe{}, broker.ErrCanNotSubscribe(subject, err)
	}

	id := gonanoid.Must()

	c.subManager.Add(subject, broker.Subscription{
		UUID:           id,
		SubCh:          subCh,
		Unsubscription: subscribe,
	})

	c.log.Info("new STAN subscription",
		zap.String("subject", subject),
		zap.String("subscriptionID", id),
	)

	return broker.Subscribe{
		Unsubscribe: broker.NewUnsubscriber(c.subManager, subject, id),
		Ch:          subCh,
	}, nil
}

func (c *Client) Publish(subject string, data []byte) error {
	if !c.nats.IsConnected() {
		return broker.ErrBrokerNotConnected
	}

	return c.stanConn.Publish(subject, data)
}

func (c *Client) IsStanExists(name string) error {
	if !c.nats.IsConnected() {
		return broker.ErrBrokerNotConnected
	}

	serverName := c.stanConn.NatsConn().ConnectedServerName()
	if name != serverName {
		return ErrExistsServerName
	}

	return nil
}

func (c *Client) QueueSubscribeWithAck(subject, queue string, ackOpt ...AckOption) (broker.Subscribe, error) {
	if !c.nats.IsConnected() {
		return broker.Subscribe{}, broker.ErrBrokerNotConnected
	}

	cfg := defaultAckConfig()
	for _, option := range ackOpt {
		option(&cfg)
	}

	var (
		subCh   = make(broker.SubCh, c.nats.Cfg.SubscriptionQueueLen)
		options = []stan.SubscriptionOption{
			stan.SetManualAckMode(),
			stan.AckWait(cfg.ackWait),
			stan.MaxInflight(cfg.maxDeliver),
		}
		handler = func(msg *stan.Msg) {
			subCh <- broker.Message{
				Subject:           subject,
				ReceiveTimeMicros: time.Now().UnixMicro(),
				Body:              msg.Data,
				AckCallback: func() error {
					return msg.Ack()
				},
			}
		}
	)

	subscribe, err := c.stanConn.QueueSubscribe(subject, queue, handler, options...)
	if err != nil {
		return broker.Subscribe{}, broker.ErrCanNotSubscribe(subject, err)
	}

	id := gonanoid.Must()

	c.subManager.Add(subject, broker.Subscription{
		UUID:           id,
		SubCh:          subCh,
		Unsubscription: subscribe,
	})

	c.log.Info("new STAN subscription",
		zap.String("subject", subject),
		zap.String("subscriptionID", id),
	)

	return broker.Subscribe{
		Unsubscribe: broker.NewUnsubscriber(c.subManager, subject, id),
		Ch:          subCh,
	}, nil
}
