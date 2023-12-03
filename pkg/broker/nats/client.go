package nats

import (
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"go.uber.org/zap"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"

	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker/submanager"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/logger"
)

type Client struct {
	Cfg               broker.NATSConfig
	subManager        *submanager.Manager
	conn              *nats.Conn
	reconnectHandlers []func()
	log               *zap.Logger
}

func New(config broker.NATSConfig) *Client {
	if config.MaxReconnects <= 0 {
		config.MaxReconnects = broker.DefaultMaxReconnects
	}

	if config.SubscriptionQueueLen <= 0 {
		config.SubscriptionQueueLen = broker.DefaultQueueLength
	}

	if config.ReconnectsTimeout <= 0 {
		config.ReconnectsTimeout = broker.DefaultReconnectsTimeout
	}

	return &Client{
		Cfg:        config,
		subManager: submanager.New(),
		log:        logger.MustGetLogger(),
	}
}

func (c *Client) Connect() error {
	if c.IsConnected() {
		return nil
	}

	var (
		url = c.Cfg.Addr
		err error
	)

	c.conn, err = nats.Connect(
		url,
		nats.UserInfo(c.Cfg.User, c.Cfg.Password),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(c.Cfg.MaxReconnects),
		nats.ReconnectWait(c.Cfg.ReconnectsTimeout),
		nats.ReconnectHandler(reconnectHandler(c)),
		nats.DisconnectErrHandler(disconnectHandler),
	)
	if err != nil {
		return broker.ErrCanNotConnect("can not connect to nats", err)
	}

	c.log.Debug("client connected", zap.String("url", url))

	return nil
}

func (c *Client) Disconnect() error {
	if !c.IsConnected() {
		return nil
	}

	if err := c.subManager.Drain(); err != nil {
		return err
	}

	c.conn.Close()

	return nil
}

func (c *Client) SetReconnectHandler(handler func()) {
	c.reconnectHandlers = append(c.reconnectHandlers, handler)
}

func (c *Client) Subscribe(subject string) (broker.Subscribe, error) {
	if !c.IsConnected() {
		return broker.Subscribe{}, broker.ErrBrokerNotConnected
	}

	subCh := make(broker.SubCh, c.Cfg.SubscriptionQueueLen)

	subscribe, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		subCh <- broker.Message{
			Subject:           subject,
			ReceiveTimeMicros: time.Now().UnixMicro(),
			Body:              msg.Data,
			ReplyCallback:     msg.Respond,
			AckCallback: func() error {
				return msg.Ack()
			},
		}
	})
	if err != nil {
		return broker.Subscribe{}, broker.ErrCanNotSubscribe(subject, err)
	}

	id := gonanoid.Must()

	c.subManager.Add(subject, broker.Subscription{
		UUID:           id,
		SubCh:          subCh,
		Unsubscription: subscribe,
	})

	c.log.Info(
		"new Client subscription to %s subject, ID of subscriptions %s",
		zap.String("subject", subject),
		zap.String("subscriptionID", id),
	)

	return broker.Subscribe{
		Unsubscribe: broker.NewUnsubscriber(c.subManager, subject, id),
		Ch:          subCh,
	}, nil
}

func (c *Client) Request(subject string, body []byte, timeout time.Duration) (broker.Message, error) {
	if !c.IsConnected() {
		return broker.Message{}, broker.ErrBrokerNotConnected
	}

	request, err := c.conn.Request(subject, body, timeout)
	if err != nil {
		return broker.Message{}, err
	}

	return broker.Message{
		Subject:           subject,
		ReceiveTimeMicros: time.Now().UnixMicro(),
		Body:              request.Data,
		ReplyCallback:     request.Respond,
	}, nil
}

func (c *Client) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

func reconnectHandler(n *Client) nats.ConnHandler {
	return func(conn *nats.Conn) {
		n.log.Warn("client is reconnected")
		n.conn = conn

		for _, handler := range n.reconnectHandlers {
			handler()
		}
	}
}

func disconnectHandler(_ *nats.Conn, err error) {
	logger.MustGetLogger().Error("client disconnected", zap.Error(err))
}

func (c *Client) StanConn(clusterID, clientID string) (stan.Conn, error) {
	if !c.IsConnected() {
		return nil, broker.ErrBrokerNotConnected
	}

	return stan.Connect(clusterID, clientID, stan.NatsConn(c.conn))
}
