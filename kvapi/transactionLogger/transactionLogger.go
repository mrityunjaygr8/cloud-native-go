package transactionlogger

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)

	Err() <-chan error
	ReadEvents() (<-chan Event, <-chan error)
	Run()
	Close()
}

type EventType byte

const (
	_                     = iota
	EventDelete EventType = iota
	EventPut
)

type Event struct {
	Sequence  uint64
	EventType EventType
	Key       string
	Value     string
}
