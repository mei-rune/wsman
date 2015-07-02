package wsman

import (
	"bytes"
	"encoding/xml"

	"github.com/runner-mei/wsman/envelope"
)

func Get(ep *Endpoint, ns, name string, selectorSet map[string]string) (map[string]interface{}, error) {
	input := &envelope.Get{
		Namespace:   ns,
		MessageId:   Uuid(),
		Name:        name,
		SelectorSet: selectorSet}
	reader, err := ep.Deliver(bytes.NewBufferString(input.Xml()))
	if nil != err {
		return nil, err
	}
	defer closeReader(reader)

	decoder := xml.NewDecoder(reader)
	ok, err := locateElements(decoder, []string{"Envelope", "Body"})
	if nil != err {
		return nil, err
	}
	if !ok {
		return nil, ElementNotExists("Envelope/Body")
	}

	_, _, err = nextElement(decoder)
	if nil != err {
		return nil, err
	}
	return toMap(decoder)
}
