package services

import (
	"automation-hub-idp/internal/app/services/iservice"
	"fmt"
	"github.com/IBM/sarama"
	"log"
)

type KafkaLogger struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaLogger(brokers []string, topic string) (iservice.Logger, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &KafkaLogger{
		producer: producer,
		topic:    topic,
	}, nil
}

func (k *KafkaLogger) sendMessage(level, message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf("[%s] %s", level, fmt.Sprintf(message, args...))
	msg := &sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.StringEncoder(formattedMessage),
	}

	_, _, err := k.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to send %s message to Kafka: %v", level, err)
	}
}

func (k *KafkaLogger) Info(message string, args ...interface{}) {
	k.sendMessage("INFO", message, args...)
}

func (k *KafkaLogger) Error(message string, args ...interface{}) {
	k.sendMessage("ERROR", message, args...)
}

func (k *KafkaLogger) Warn(message string, args ...interface{}) {
	k.sendMessage("WARN", message, args...)
}

func (k *KafkaLogger) Debug(message string, args ...interface{}) {
	k.sendMessage("DEBUG", message, args...)
}
