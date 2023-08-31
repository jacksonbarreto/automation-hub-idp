package iservice

type MessageSender interface {
	Send(topic string, message interface{}) error
}
