package wsman

import (
	"encoding/xml"
	"log"
	"strings"

	"github.com/runner-mei/wsman/envelope"
)

func Invoke(ep *Endpoint, ns, name string, selectorSet map[string]string,
	method string, inParameters map[string]interface{}) (map[string]interface{}, error) {
	input := &envelope.InvokeMethod{
		Namespace:    ns,
		MessageId:    Uuid(),
		Name:         name,
		SelectorSet:  selectorSet,
		Method:       method,
		InParameters: inParameters,
	}

	reader, err := ep.Deliver(strings.NewReader(input.Xml()))
	if nil != err {
		return nil, err
	}
	defer closeReader(reader)

	decoder := xml.NewDecoder(reader)
	if err = ReadEnvelopeBody(decoder); nil != err {
		return nil, err
	}
	tagName, _, err := nextElement(decoder)
	if nil != err {
		return nil, err
	}
	if tagName.Local != method+"_OUTPUT" {
		log.Println(tagName)
	}

	return toMap(decoder)
}
