package submanager

import (
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/logger"
)

// Manager manages provided broker.Subscription's. This struct is thread-safe.
type Manager struct {
	mutex sync.Mutex
	subs  map[string][]broker.Subscription
}

func New() *Manager {
	return &Manager{
		subs: make(map[string][]broker.Subscription),
	}
}

func (sbm *Manager) Add(subject string, sub broker.Subscription) {
	sbm.mutex.Lock()

	sbm.subs[subject] = append(sbm.subs[subject], sub)

	sbm.mutex.Unlock()
}

func (sbm *Manager) Remove(subject string, id string) {
	sbm.mutex.Lock()

	if subs, ok := sbm.subs[subject]; ok {
		for i, sub := range subs {
			if sub.UUID == id {
				if err := sub.Unsubscription.Unsubscribe(); err != nil {
					logger.MustGetLogger().Warn(
						"unsubscribe error",
						zap.String("clientID", sub.UUID),
						zap.Error(err),
					)
				}

				sbm.subs[subject] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}

	sbm.mutex.Unlock()
}

// Drain removes all broker.Subscription's from *Manager. This function
// close broker.SubCh, also, unsubscribe from incoming updates from broker.
func (sbm *Manager) Drain() error {
	sbm.mutex.Lock()
	defer sbm.mutex.Unlock()

	for subject, subscriptions := range sbm.subs {
		for _, sub := range subscriptions {
			if err := sub.Unsubscription.Unsubscribe(); err != nil {
				return fmt.Errorf("err while unsubscrubing from subject: %s, err:  %w", subject, err)
			}
			close(sub.SubCh)
		}
		delete(sbm.subs, subject)
	}

	return nil
}
