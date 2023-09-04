package config

import (
	"errors"
	"strings"
)

const (
	loggerTopic string = "LOGGER_TOPIC"
	mailTopic   string = "MAIL_TOPIC"
	clientID    string = "KAFKA_CLIENT_ID"
	brokersAddr string = "BROKERS_ADDR"
)

type kafkaConfig struct {
	LoggerTopic string
	MailTopic   string
	ClientID    string
	BrokersAddr []string
}

func newKafkaConfig() (*kafkaConfig, error) {
	logTopic := getEnvString(loggerTopic, "NULL")
	if logTopic == "NULL" {
		return nil, errors.New("error: LoggerTopic is not set, please check the environment variable: " + loggerTopic)
	}
	emailTopic := getEnvString(mailTopic, "NULL")
	if emailTopic == "NULL" {
		return nil, errors.New("error: MailTopic is not set, please check the environment variable: " + mailTopic)
	}
	brokers := getEnvString(brokersAddr, "NULL")
	if brokers == "NULL" {
		return nil, errors.New("error: Kafka Brokers Logger are not set, please check the environment variable: " + brokersAddr)
	}

	brokersList := strings.Split(brokers, ",")

	return &kafkaConfig{
		LoggerTopic: logTopic,
		MailTopic:   emailTopic,
		ClientID:    getEnvString(clientID, "IDP-AUTOMATIONS-HUB"),
		BrokersAddr: brokersList,
	}, nil
}
