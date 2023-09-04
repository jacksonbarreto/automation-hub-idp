package config

import (
	"errors"
	"strings"
)

const (
	loggerTopic string = "LOGGER_TOPIC"
	brokersAddr string = "BROKERS_ADDR"
)

type kafkaLoggerConfig struct {
	Topic       string
	BrokersAddr []string
}

func newKafkaLoggerConfig() (*kafkaLoggerConfig, error) {
	topic := getEnvString(loggerTopic, "NULL")
	if topic == "NULL" {
		return nil, errors.New("error: Topic is not set, please check the environment variable: " + loggerTopic)
	}
	brokers := getEnvString(brokersAddr, "NULL")
	if brokers == "NULL" {
		return nil, errors.New("error: Kafka Brokers Logger are not set, please check the environment variable: " + brokersAddr)
	}

	brokersList := strings.Split(brokers, ",")

	return &kafkaLoggerConfig{
		Topic:       topic,
		BrokersAddr: brokersList,
	}, nil
}
