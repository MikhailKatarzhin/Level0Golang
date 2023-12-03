package broker

type Remover interface {
	// Remove subscription by provided ID from the provided subject.
	Remove(subject, id string)
}

type DefaultUnsubscriber struct {
	Subject string
	ID      string
	remover Remover
}

func NewUnsubscriber(remover Remover, subject, id string) *DefaultUnsubscriber {
	return &DefaultUnsubscriber{
		Subject: subject,
		ID:      id,
		remover: remover,
	}
}

func (uns *DefaultUnsubscriber) Unsubscribe() error {
	uns.remover.Remove(uns.Subject, uns.ID)
	return nil
}
