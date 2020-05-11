package events

import "encoding/json"

// Header is the data structure for headers used when building Envelopes.
// It is a flexible way of storing information to be used in any part of the publishing process
type Header struct {
	Headers map[string]string `json:"headers"`
}

// Envelope is the data structure that holds the information to be published by the Broker
type Envelope struct {
	Header  *Header     `json:"header"`
	Message interface{} `json:"message"`
}

// AddHeader adds a new header for a particular Envelope
func (e *Envelope) AddHeader(key, value string) {
	if e.Header == nil {
		e.Header = &Header{
			Headers: map[string]string{},
		}
	}
	e.Header.Headers[key] = value
}

// Encode returns the encoded version of the Envelope.
// This is useful when sending non structure data on the wire
func (e *Envelope) Encode() ([]byte, error) {
	return json.Marshal(e)
}
