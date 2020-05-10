package events

type Header struct {
	Headers map[string]string `json:"headers"`
}

type Envelope struct {
	Header  *Header     `json:"header"`
	Message interface{} `json:"message"`
}

func (e *Envelope) AddHeader(key, value string) {
	if e.Header == nil {
		e.Header = &Header{
			Headers: map[string]string{},
		}
	}
	e.Header.Headers[key] = value
}
