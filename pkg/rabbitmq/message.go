package rabbitmq

type Message struct {
	Id        string `json:"id"`
	Body      any    `json:"body"`
	Timestamp int64  `json:"timestamp"`
}
