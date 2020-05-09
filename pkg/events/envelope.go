package events

type Header struct {
	Headers map[string]string
}

type Envelope struct {
	Header  Header
	Message Message
}

func (e *Envelope) AddHeader(key, value string) {
	e.Header.Headers[key] = value
}
