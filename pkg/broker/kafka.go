package broker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"shop/pkg/inbox"
	"sync"

	"github.com/IBM/sarama"
)

type KafkaBroker struct {
	producer sarama.SyncProducer
	consumer sarama.ConsumerGroup
	inbox    inbox.Inbox
	logger   *log.Logger
	subs     map[string]Handler
	mu       sync.Mutex
}

func NewKafkaBroker(producer sarama.SyncProducer, consumer sarama.ConsumerGroup, logger *log.Logger) *KafkaBroker {
	return &KafkaBroker{producer: producer, consumer: consumer, logger: logger, subs: make(map[string]Handler)}
}

func (b *KafkaBroker) Publish(message Message) error {
	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}

	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(message.Key),
		Topic: message.Topic,
		Value: sarama.StringEncoder(msg),
	}

	partition, offset, err := b.producer.SendMessage(m)
	if err != nil {
		b.logger.Fatalf("Failed to send message: %v", err)
	}

	b.logger.Printf("Message sent to topic %s, partition %d, offset %d\n", message.Topic, partition, offset)

	return nil
}

func (b *KafkaBroker) PublishBatch(messages []Message) error {
	err := b.producer.BeginTxn()
	if err != nil {
		b.logger.Fatalf("Failed to begin transaction: %v", err)
		return err
	}

	var saramaMessages []*sarama.ProducerMessage
	for _, msg := range messages {
		saramaMessages = append(saramaMessages, &sarama.ProducerMessage{
			Topic: msg.Topic,
			Key:   sarama.StringEncoder(msg.Key),
			Value: sarama.StringEncoder(msg.Value),
		})
	}

	err = b.producer.SendMessages(saramaMessages)
	if err != nil {
		b.logger.Printf("Failed to send message: %v", err)
		err = b.producer.AbortTxn()
		if err != nil {
			b.logger.Printf("Failed to abort transaction: %v", err)
			return err
		}
		return err
	}

	err = b.producer.CommitTxn()
	if err != nil {
		b.logger.Printf("Failed to commit transaction: %v", err)
		return err
	}

	b.logger.Printf("Messages sent")

	return nil
}

func (b *KafkaBroker) Subscribe(topic string, handler Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subs[topic]; ok {
		return errors.New("already subscribed")
	}

	b.subs[topic] = handler

	return nil
}

func (b *KafkaBroker) StartConsume(topics []string) {
	b.logger.Println("Start consume kafka")

	for {
		err := b.consumer.Consume(
			context.Background(),
			topics,
			&consumerGroupHandler{broker: b},
		)
		if err != nil {
			b.logger.Printf("Error while consuming: %v", err)
			b.consumer.Close()
			break
		}
	}
}

type consumerGroupHandler struct {
	broker *KafkaBroker
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.broker.logger.Println("Setup kafka consumer")
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.broker.logger.Println("Cleanup kafka consumer")
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.broker.logger.Printf("Message claimed: value = %s, topic = %s, partition = %d, offset = %d", message.Value, message.Topic, message.Partition, message.Offset)
		msg := Message{
			Topic: message.Topic,
			Key:   string(message.Key),
			Value: message.Value,
		}

		handler, ok := h.broker.subs[message.Topic]
		if !ok {
			h.broker.logger.Printf("Cannot find handler for topic %s\n", message.Topic)
			return errors.New("not subscribed")
		}

		err := handler.Handle(msg)
		if err != nil {
			h.broker.logger.Printf("Error handling message: %v", err)
			return err
		}

		h.broker.logger.Printf("Marking message as processed for topic %s\n", message.Topic)
		session.MarkMessage(message, "")
		h.broker.logger.Printf("Message marked as processed for topic %s\n", message.Topic)

		// manual commit (slow) (very slow?)
		//session.Commit()
	}
	return nil
}
