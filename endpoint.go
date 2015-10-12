package wsman

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

var (
	WSMAN_DEBUG = "" != os.Getenv("WSMAN_DEBUG")
)

var ElementEndError = errors.New("current element is end.")
var ErrHttpAuthenticate = &HttpError{401, "failed to authenticate"}
var ErrHttpNotFound = &HttpError{404, "nothing listening on the endpoint"}
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
		case xml.StartElement:
			switch v.Name.Local {
			case "Datetime":
				txt, e := readXmlText(decoder)
				if nil != e {
					return "", e
				}
				if e = exitElement(decoder, 0); nil != e {
					return txt, e
				}
				return txt, nil
			default:
				return context, errors.New("element '" + v.Name.Local + "' is not excepted, excepted is CharData")
			}
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

func exitElement(decoder *xml.Decoder, depth int) error {
	for {
		t, err := decoder.Token()
		if nil != err {
			return err
		}
		switch v := t.(type) {
		case xml.EndElement:
			depth--
			if depth < 0 {
				return nil
			}
		case xml.StartElement:
			return errors.New("StartElement '" + v.Name.Local + "' is not excepted, excepted is EndElement")
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
			var value interface{}
			if is_nil(v.Attr) {
				if e := skipElement(decoder, 0); nil != e {
					return nil, e
				}
				value = nil
			} else {
				txt, e := readXmlText(decoder)
				if nil != e {
					if ElementEndError != e {
						return nil, e
					} else {
						value = nil
					}
				} else {
					if "" == txt && len(v.Attr) > 0 {
						for _, attr := range v.Attr {
							if "SystemTime" == attr.Name {
								txt = attr.Value
							}
						}
					}
					value = txt
				}
			}

			old, ok := m[v.Name.Local]
			if !ok {
				m[v.Name.Local] = value
			} else if aa, ok := old.([]interface{}); ok {
				m[v.Name.Local] = append(aa, value)
			} else {
				m[v.Name.Local] = []interface{}{old, value}
			}
		}
	}
}

type Deliverable interface {
	Xml() string
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
		defer func() {
			io.Copy(ioutil.Discard, response.Body)
			response.Body.Close()
		}()

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

func closeReader(reader io.Reader) error {
	io.Copy(ioutil.Discard, reader)
	if closer, ok := reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

type HttpError struct {
	StatusCode int
	Status     string
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Status)
}

func handleError(r *http.Response) error {
	if r.StatusCode == 404 {
		return ErrHttpNotFound
	}
	if r.StatusCode == 401 {
		return ErrHttpAuthenticate
	}

	if h := r.Header.Get("Content-Type"); strings.HasPrefix(h, "application/soap+xml") {
		return handleFault(r)
	}

	return &HttpError{r.StatusCode, r.Status}
}

func handleFault(r *http.Response) error {
	var decoder *xml.Decoder

	if os.Getenv("WINRM_DEBUG") != "" {
		body, _ := ioutil.ReadAll(r.Body)
		log.Println("winrm: fault", string(body))
		decoder = xml.NewDecoder(bytes.NewBuffer(body))
	} else {
		decoder = xml.NewDecoder(r.Body)
	}

	f := &HttpError{500, "Unparsable SOAP error"}

	locateElements(decoder, []string{"Envelope", "Body", "Fault", "Reason", "Text"})

	if reason, _ := readXmlText(decoder); "" != reason {
		f.Status = "FAULT: " + reason
	}

	return f
}

func doFault(decoder *xml.Decoder) (bool, error) {
	return false, nil
}

// 	ok, err := locateElements(decoder, []string{"Envelope", "Body", "Fault"})
// 	if !ok {
// 		return false, err
// 	}

// 	if reason, _ := readXmlText(decoder); "" != reason {
// 		f.Status = "FAULT: " + reason
// 	}
// }

type ErrSoapFault struct {
	Code    string
	Subcode string
	Reason  string
	Detail  string
}

func (fault *ErrSoapFault) Error() string {
	return fault.Detail
}

func ToBytes(bs [][]byte) []byte {
	var buf bytes.Buffer
	for _, b := range bs {
		if nil != b {
			buf.Write(b)
		}
	}
	return buf.Bytes()
}

func ToString(bs [][]byte) string {
	var buf bytes.Buffer
	for _, b := range bs {
		if nil != b {
			buf.Write(b)
		}
	}
	return buf.String()
}

func ToUtf16String(bs [][]byte, o binary.ByteOrder) string {
	return UTF16BytesToString(ToBytes(bs), o)
}

func ToWinString(bs [][]byte) string {
	return ToUtf16String(bs, binary.LittleEndian)
}

func UTF16BytesToString(b []byte, o binary.ByteOrder) string {
	utf := make([]uint16, (len(b)+(2-1))/2)
	for i := 0; i+(2-1) < len(b); i += 2 {
		utf[i/2] = o.Uint16(b[i:])
	}
	if len(b)/2 < len(utf) {
		utf[len(utf)-1] = utf8.RuneError
	}
	return string(utf16.Decode(utf))
}

func Uuid() string {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic("create uuid failed, " + err.Error())
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

func ReadEnvelopeBody(decoder *xml.Decoder) error {
	ok, err := locateElement(decoder, "Envelope")
	if nil != err {
		return err
	}
	if !ok {
		return ElementNotExists("Envelope")
	}

	var action string
	for {
		nm, _, err := nextElement(decoder)
		switch nm.Local {
		case "Header":
			action, err = ReadEnvelopeHeader(decoder)
			if nil != err {
				return err
			}
		case "Body":
			action = strings.ToLower(strings.TrimSpace(action))
			if strings.HasSuffix(action, "/fault") {
				return ReadEnvelopeFault(decoder)
			}
			return nil
		}
	}
}

func ReadEnvelopeFault(decoder *xml.Decoder) error {
	ok, err := locateElement(decoder, "Fault")
	if nil != err {
		return err
	}
	if !ok {
		return errors.New("fault isn't exists in the FaultMessage.")
	}

	e := &ErrSoapFault{}
	var state = 0
	for {
		t, err := decoder.Token()
		if nil != err {
			if io.EOF == err {
				return e
			}
			return err
		}

		switch v := t.(type) {
		case xml.EndElement:
			switch state {
			case 2: // SubCode
				state = 1 // Code
			default:
				state = 0
			}
		case xml.StartElement:
			switch v.Name.Local {
			case "Code":
				state = 1
			case "Subcode":
				state = 2
			case "Value", "faultcode":
				if code, err := readXmlText(decoder); nil != err {
					return err
				} else {
					switch state {
					case 2:
						e.Subcode = code
					default:
						e.Code = code
						if "" == e.Subcode {
							e.Subcode = code
						}
					}
				}
			case "Text":
				if txt, err := readXmlText(decoder); nil != err {
					return err
				} else {
					e.Reason = txt
				}
			case "Reason":
				state = 3
			case "Detail":
				if detail, err := readXmlText(decoder); nil != err {
					return err
				} else {
					e.Detail = strings.TrimSpace(detail)
				}
			case "faultstring":
				if detail, err := readXmlText(decoder); nil != err {
					return err
				} else {
					e.Reason = detail
					e.Detail = detail
				}
			default:
				if e := skipElement(decoder, 0); nil != e {
					return e
				}
			}
		}
	}
}

func ReadEnvelopeHeader(decoder *xml.Decoder) (string, error) {
	var action string
	for {
		t, err := decoder.Token()
		if nil != err {
			return "", err
		}

		switch v := t.(type) {
		case xml.EndElement:
			return action, nil
		case xml.StartElement:
			switch v.Name.Local {
			case "Action":
				action, err = readXmlText(decoder)
				if nil != err {
					return "", err
				}
			default:
				if e := skipElement(decoder, 0); nil != e {
					return "", e
				}
			}
		}
	}
}

func ReadEnvelopeResponse(decoder *xml.Decoder, responseName string) error {
	if err := ReadEnvelopeBody(decoder); nil != err {
		return err
	}

	if ok, err := locateElement(decoder, responseName); nil != err {
		return err
	} else if !ok {
		return ElementNotExists("Envelope/Body/" + responseName)
	}
	return nil
}
