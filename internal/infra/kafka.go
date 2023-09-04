package infra

import (
	"fmt"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"idp-automations-hub/internal/app/config"
)

func NewKafkaProducer(brokers []string, client string) (*kafka.Producer, error) {
	producerConfig := &kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"client.id":         client,
		"acks":              "all",
	}

	producer, err := kafka.NewProducer(producerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	return producer, nil
}

func GetDefaultKafkaProducer() (*kafka.Producer, error) {
	return NewKafkaProducer(config.KafkaConfig.BrokersAddr, config.KafkaConfig.ClientID)
}
