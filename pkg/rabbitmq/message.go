package rabbitmq

type Message struct {
	ID        string `json:"id"`
	Body      any    `json:"body"`
	Timestamp int64  `json:"timestamp"`
}
