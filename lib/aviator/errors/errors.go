package errors

type Message struct {
	FR string
	EN string
}

type AviatorError struct {
	Id       string
	Message  Message
	ApiError int
}

func (p AviatorError) Error() string {
	return p.Message.EN
}
