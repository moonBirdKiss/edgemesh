package conn

const (
	MaxChannel = 4
)

var AvailableChannel chan struct{}

func init() {
	AvailableChannel = make(chan struct{}, MaxChannel)
}
