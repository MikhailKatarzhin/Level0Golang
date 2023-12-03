package broker

type Unsubscriber interface {
	Unsubscribe() error
}

type Subscription struct {
	UUID           string
	SubCh          SubCh
	Unsubscription Unsubscriber
}

type SubCh chan Message

type Message struct {
	Subject           string
	ReceiveTimeMicros int64
	Body              []byte
	ReplyCallback     func([]byte) error
	AckCallback       func() error
}

type Subscribe struct {
	Unsubscribe Unsubscriber
	Ch          SubCh
}
