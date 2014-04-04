package wsman

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/runner-mei/wsman/envelope"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var ElementEndError = errors.New("current element is end.")
var EMPTY_NAME = xml.Name{}

func ElementNotExists(nm string) error {
	return errors.New("'" + nm + "' is not exists.")
}

func readXmlText(decoder *xml.Decoder) (string, error) {
	var context string
	for {
		token, err := decoder.Token()
		if nil != err {
			return context, err
		}
		switch v := token.(type) {
		case xml.EndElement:
			return context, nil
		case xml.CharData:
			context = string(v)
		default:
			return context, fmt.Errorf("token '%T' is not excepted, excepted is CharData", v)
		}
	}
}

func locateElement(decoder *xml.Decoder, nm string) (bool, error) {
	depth := 0
	for {
		t, err := decoder.Token()
		if nil != err {
			return false, err
		}
		switch t := t.(type) {
		case xml.EndElement:
			depth--
			if depth < 0 {
				return false, nil
			}
		case xml.StartElement:
			if 0 == depth && nm == t.Name.Local {
				return true, nil
			}
			depth++
		}
	}
}

func nextElement(decoder *xml.Decoder) (xml.Name, []xml.Attr, error) {
	for {
		t, err := decoder.Token()
		if nil != err {
			return EMPTY_NAME, nil, err
		}
		switch el := t.(type) {
		case xml.EndElement:
			return el.Name, nil, ElementEndError
		case xml.StartElement:
			return el.Name, el.Attr, nil
		}
	}
}

func skipElement(decoder *xml.Decoder, depth int) error {
	for {
		t, err := decoder.Token()
		if nil != err {
			return err
		}
		switch t.(type) {
		case xml.EndElement:
			depth--
			if depth < 0 {
				return nil
			}
		case xml.StartElement:
			depth++
		}
	}
}

func locateElements(decoder *xml.Decoder, names []string) (bool, error) {
	for _, nm := range names {
		ok, e := locateElement(decoder, nm)
		if nil != e {
			return false, e
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func is_nil(attrs []xml.Attr) bool {
	for _, attr := range attrs {
		if "nil" == attr.Name.Local && "true" == attr.Value {
			return true
		}
	}
	return false
}

func toMap(decoder *xml.Decoder) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	for {
		t, err := decoder.Token()
		if nil != err {
			return nil, err
		}
		switch v := t.(type) {
		case xml.EndElement:
			return m, nil
		case xml.StartElement:
			if is_nil(v.Attr) {
				m[v.Name.Local] = nil
				if e := skipElement(decoder, 0); nil != e {
					return nil, e
				}
				break
			}

			txt, e := readXmlText(decoder)
			if nil != e {
				if ElementEndError != e {
					return nil, e
				} else {
					m[v.Name.Local] = nil
				}
			} else {
				m[v.Name.Local] = txt
			}
		}
	}
}

type Endpoint struct {
	Url      string
	User     string
	Password string
}

func (ep *Endpoint) Deliver(reader io.Reader) (io.Reader, error) {
	if WSMAN_DEBUG {
		var buffer *bytes.Buffer
		if buf, ok := reader.(*bytes.Buffer); ok {
			buffer = buf
		} else if rd, ok := reader.(*bytes.Reader); ok {
			buffer := bytes.NewBuffer(make([]byte, 0, rd.Len()))
			io.Copy(buffer, rd)
			rd.Seek(0, 0)
		}
		log.Println("wsman: sending", buffer.String())
	}

	request, _ := http.NewRequest("POST", ep.Url, reader)
	request.SetBasicAuth(ep.User, ep.Password)
	request.Header.Add("Content-Type", "application/soap+xml;charset=UTF-8")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		defer response.Body.Close()
		return nil, handleError(response)
	}

	if WSMAN_DEBUG {
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		log.Println("wsman: receiving", string(body))
		return bytes.NewReader(body), nil
	}

	return response.Body, nil
}

type Enumerator struct {
	*Endpoint
	Name    string
	Context string

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

func Enumerate(ep *Endpoint, name string) *Enumerator {
	return &Enumerator{Endpoint: ep, Name: name}
}

func (c *Enumerator) Close() {
	if nil != c.reader {
		io.Copy(ioutil.Discard, c.reader)
		if closer, ok := c.reader.(io.Closer); ok {
			closer.Close()
		}
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
			io.Copy(ioutil.Discard, c.reader)
			if closer, ok := c.reader.(io.Closer); ok {
				closer.Close()
			}
			c.reader = nil
		}

		var input Deliverable
		var responseName string
		if !c.is_pull {
			input = &envelope.Enumerate{MessageId: uuid.TimeOrderedUUID(),
				Name: c.Name}
			responseName = "EnumerateResponse"
			c.is_pull = true
		} else {
			if "" == c.Context {
				c.err = errors.New("EnumerationContext is empty.")
				return false
			}
			input = &envelope.Pull{MessageId: uuid.TimeOrderedUUID(),
				Name: c.Name, Context: c.Context}
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
		c.err = errors.New("'" + nm.Local + "' is unknown element in the .")
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
