package stan

import (
	"context"
	"errors"
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/onsi/ginkgo/v2"
	"go.uber.org/zap"

	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/logger"
)

type WorkerPool struct {
	client *Client
	// map of workers by subject and consumer name.
	workers map[[2]string]workers
}

type workers []*broker.Subscribe

func (w workers) messageChannels() []chan broker.Message {
	messageChannels := make([]chan broker.Message, 0, len(w))

	for i := range w {
		messageChannels = append(messageChannels, w[i].Ch)
	}

	return messageChannels
}

func (s *WorkerPool) RunWorkers(
	workersNum int, subject string, consumer string,
	clusterID string, clientID string,
	ackOpts ...AckOption,
) ([]chan broker.Message, error) {
	clientID = fmt.Sprint(clientID, "-", gonanoid.Must(5))

	if err := s.client.Connect(clusterID, clientID); err != nil {
		return nil, errors.Join(ErrCanNotToConnect, err)
	}

	key := s.key(subject, consumer)

	for i := 0; i < workersNum; i++ {
		subscription, err := s.client.QueueSubscribeWithAck(
			subject, consumer, ackOpts...,
		)
		if err != nil {
			return nil, errors.Join(ErrCanNotSubscribe, err)
		}

		s.workers[key] = append(s.workers[key], &subscription)
	}

	return s.workers[key].messageChannels(), nil
}

func (s *WorkerPool) RunWorkersWithFunc(
	ctx context.Context, workersNum int, subject string, consumer string,
	clusterID string, clientID string,
	handler func(msg []byte) error, ackOpts ...AckOption,
) error {
	channels, err := s.RunWorkers(
		workersNum, subject, consumer, clusterID, clientID, ackOpts...,
	)

	if err != nil {
		return err
	}

	for i := range channels {
		go func(ch chan broker.Message) {
			defer ginkgo.GinkgoRecover()

			for {
				select {
				case msg := <-ch:
					if err := handler(msg.Body); err != nil {
						logger.L().Error(
							"can not to handle message in the stan worker",
							zap.Error(err),
							zap.String("subject", subject),
							zap.String("consumer", consumer),
							zap.String("clusterID", clusterID),
							zap.String("clientID", clientID),
							zap.String("message", string(msg.Body)),
						)

						continue
					}

					if err := msg.AckCallback(); err != nil {
						logger.L().Error(
							"can not to ack message in the stan worker",
							zap.Error(err),
							zap.String("subject", subject),
							zap.String("consumer", consumer),
							zap.String("clusterID", clusterID),
							zap.String("clientID", clientID),
							zap.String("message", string(msg.Body)),
						)
					}
				case <-ctx.Done():
					return
				}
			}
		}(channels[i])
	}

	return nil
}

func (s *WorkerPool) key(subject, consumer string) [2]string {
	return [2]string{subject, consumer}
}

func (s *WorkerPool) Close() error {
	err := s.client.Disconnect()

	for _, workers := range s.workers {
		for i := range workers {
			if err := workers[i].Unsubscribe.Unsubscribe(); err != nil {
				logger.L().Error("can't unsubscribe from stan", zap.Error(err))
			}

			close(workers[i].Ch)
		}
	}

	if err != nil {
		return errors.Join(ErrCanNotToDisconnect, err)
	}

	return nil
}
