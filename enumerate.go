package wsman

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/runner-mei/wsman/envelope"
)

type Enumerator struct {
	*Endpoint
	Namespace   string
	Name        string
	SelectorSet map[string]string
	OptionSet   map[string]string
	Context     string

	is_debug          bool
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

func Enumerate(ep *Endpoint, ns, name string, selectorSet map[string]string) *Enumerator {
	return &Enumerator{Namespace: ns, Endpoint: ep, Name: name, SelectorSet: selectorSet}
}

func (c *Enumerator) EnableDebug() {
	c.is_debug = true
}

func (c *Enumerator) Close() (err error) {
	if nil != c.reader {
		err = closeReader(c.reader)
		c.reader = nil
	}
	c.decoder = nil
	return
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
			input = &envelope.Enumerate{Namespace: c.Namespace, MessageId: Uuid(),
				Name: c.Name, SelectorSet: c.SelectorSet, OptionSet: c.OptionSet}
			responseName = "EnumerateResponse"
			c.is_pull = true
		} else {
			if "" == c.Context {
				c.err = errors.New("EnumerationContext is empty.")
				return false
			}
			input = &envelope.Pull{Namespace: c.Namespace, MessageId: Uuid(),
				Name: c.Name, SelectorSet: c.SelectorSet, OptionSet: c.OptionSet, Context: c.Context}
			responseName = "PullResponse"
		}

		if c.is_debug {
			fmt.Println(input.Xml())
		}

		reader, err := c.Deliver(bytes.NewBufferString(input.Xml()))
		if nil != err {
			c.err = err
			return false
		}

		c.reader = reader

		if c.is_debug {
			reader = io.TeeReader(reader, os.Stdout)
		}
		c.decoder = xml.NewDecoder(reader)
		if err = ReadEnvelopeResponse(c.decoder, responseName); nil != err {
			c.err = err
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

func (c *Enumerator) Value() (map[string]interface{}, error) {
	//fmt.Println("=================== map")
	if c.is_end || nil != c.err {
		return nil, c.err
	}
	if nil != c.current_value {
		return c.current_value, nil
	}

	var e error
	switch c.current_name.Local {
	case "Event":
		e = errors.New("EVENT IS NOT IMPLEMENTED")
	default:
		c.current_value, e = toMap(c.decoder)
	}
	if nil != e {
		c.err = e
		return nil, e
	}

	return c.current_value, nil
}

func ReadEvent(decoder *xml.Decoder) (map[string]interface{}, error) {
	results := map[string]interface{}{}
	for {
		t, err := decoder.Token()
		if nil != err {
			if io.EOF == err {
				return results, nil
			}
			return nil, err
		}

		switch v := t.(type) {
		case xml.EndElement:
			return results, nil
		case xml.StartElement:
			switch v.Name.Local {
			case "System":
				if err = ReadEventSystemElements(decoder, results); nil != err {
					return nil, err
				}
			case "EventData":
				if err = ReadEventDataElements(decoder, results); nil != err {
					return nil, err
				}
			default:
				if err = skipElement(decoder, 0); nil != err {
					return nil, err
				}
			}
		}
	}
}

func ReadEventSystemElements(decoder *xml.Decoder, results map[string]interface{}) error {
	return errors.New("NOT IMPLEMENTED")
}

func ReadEventDataElements(decoder *xml.Decoder, results map[string]interface{}) error {
	return errors.New("NOT IMPLEMENTED")
}
