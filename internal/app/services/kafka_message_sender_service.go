package services

import (
	"automation-hub-idp/internal/infra"
	"encoding/json"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type KafkaMessageSender struct {
	Producer *kafka.Producer
}

func NewKafkaMessageSender() (*KafkaMessageSender, error) {
	prod, err := infra.GetDefaultKafkaProducer()
	if err != nil {
		return nil, err
	}
	return &KafkaMessageSender{
		Producer: prod,
	}, nil
}

func (k *KafkaMessageSender) Send(topic string, message interface{}) error {

	// Marshal the message into JSON
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Create a Kafka message
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          msgBytes,
	}

	// Produce the message to the Kafka topic
	deliveryChan := make(chan kafka.Event)
	err = k.Producer.Produce(msg, deliveryChan)
	if err != nil {
		return err
	}

	// Wait for the message to be delivered or an error to occur
	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}

	return nil
}
