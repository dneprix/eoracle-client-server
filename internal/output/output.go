package output

type Output interface {
	Write(data string)
	Close() error
}
