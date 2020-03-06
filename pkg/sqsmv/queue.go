package sqsmv

type Queue struct {
	Source      string
	Destination string
}

type Config struct {
	Queues []Queue
}
