package destinations

type Destination interface {
	Send(data []byte) error
	Validate() error
	GetURL() string 
}