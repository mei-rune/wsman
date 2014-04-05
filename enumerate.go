package wsman

import (
	"bytes"
	"encoding/xml"
	"errors"
	"github.com/runner-mei/wsman/envelope"
	"io"
)

type Enumerator struct {
	*Endpoint
	Name        string
	SelectorSet map[string]string
	Context     string

	is_end            bool
	is_pull           bool
	items_enumerating bool
	decoder           *xml.Decoder
	reader            io.Reader
	err               error

	current_name  xml.Name
	current_attrs []xml.Attr
	current_value map[string]interface{}
}

func Enumerate(ep *Endpoint, name string, selectorSet map[string]string) *Enumerator {
	return &Enumerator{Endpoint: ep, Name: name, SelectorSet: selectorSet}
}

func (c *Enumerator) Close() {
	if nil != c.reader {
		closeReader(c.reader)
		c.reader = nil
	}
	c.decoder = nil
}

func (c *Enumerator) Next() bool {
	c.current_value = nil
next_with_context:
	if c.is_end || nil != c.err {
		return false
	}

	if nil == c.decoder {
		c.items_enumerating = false

		if nil != c.reader {
			closeReader(c.reader)
			c.reader = nil
		}

		var input Deliverable
		var responseName string
		if !c.is_pull {
			input = &envelope.Enumerate{MessageId: Uuid(),
				Name: c.Name, SelectorSet: c.SelectorSet}
			responseName = "EnumerateResponse"
			c.is_pull = true
		} else {
			if "" == c.Context {
				c.err = errors.New("EnumerationContext is empty.")
				return false
			}
			input = &envelope.Pull{MessageId: Uuid(),
				Name: c.Name, SelectorSet: c.SelectorSet, Context: c.Context}
			responseName = "PullResponse"
		}

		reader, err := c.Deliver(bytes.NewBufferString(input.Xml()))
		if nil != err {
			c.err = err
			return false
		}

		c.reader = reader
		c.decoder = xml.NewDecoder(reader)
		ok, err := locateElements(c.decoder, []string{"Envelope", "Body", responseName})
		if nil != err {
			c.err = err
			return false
		}

		if !ok {
			c.err = ElementNotExists("Envelope/Body/" + responseName)
			return false
		}
	}

next_element:
	nm, attrs, err := nextElement(c.decoder)
	//fmt.Println("1===================", nm.Local, "items_enumerating =", c.items_enumerating, "err =", err)

	if c.items_enumerating {
		if nil == err {
			c.current_name = nm
			c.current_attrs = attrs
			return true
		}

		if ElementEndError != err {
			c.err = err
			return false
		}

		// current is EnumerateResponse or PullResponse if ElementEndError == err
		nm, attrs, err = nextElement(c.decoder)
		c.items_enumerating = false
		//fmt.Println("2===================", nm.Local, "items_enumerating =", c.items_enumerating, "err =", err)
	}

	if nil != err {
		if ElementEndError == err {
			c.decoder = nil
			goto next_with_context
		}

		c.err = err
		return false
	}

	switch nm.Local {
	case "EnumerationContext":
		c.Context, c.err = readXmlText(c.decoder)
		if nil != c.err {
			return false
		}
		goto next_element
	case "Items":
		c.items_enumerating = true
		goto next_element
	case "EndOfSequence":
		c.is_end = true
		return false
	default:
		c.err = errors.New("'" + nm.Local + "' is unknown element in the EnumerateResponse.")
		return false
	}
}

func (c *Enumerator) Err() error {
	return c.err
}

func (c *Enumerator) Map() (map[string]interface{}, error) {
	//fmt.Println("=================== map")
	if c.is_end || nil != c.err {
		return nil, c.err
	}
	if nil != c.current_value {
		return c.current_value, nil
	}

	m, e := toMap(c.decoder)
	if nil != e {
		c.err = e
		return nil, e
	}
	c.current_value = m
	return m, nil
}
