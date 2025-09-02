package broker

type Handler interface {
	Handle(message Message) error
}

type Message struct {
	Topic string
	Key   string
	Value []byte
}

type Broker interface {
	Publish(message Message) error
	PublishBatch(messages []Message) error
	Subscribe(topic string, handler Handler) error
}
