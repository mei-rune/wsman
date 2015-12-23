package wsman

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"strings"

	"github.com/runner-mei/wsman/envelope"
)

type Shell struct {
	*Endpoint

	Id string
}

func NewShell(endpoint, user, pass, code_page string) (*Shell, error) {
	if "" == code_page {
		code_page = "936"
	}
	env := &envelope.CreateShell{code_page, Uuid()}
	ep := &Endpoint{Url: endpoint, User: user, Password: pass}

	reader, err := ep.Deliver(bytes.NewBufferString(env.Xml()))
	if err != nil {
		return nil, err
	}
	defer closeReader(reader)

	decoder := xml.NewDecoder(reader)
	if err := ReadEnvelopeBody(decoder); nil != err {
		return nil, err
	}
	for {
		nm, _, err := nextElement(decoder)
		switch nm.Local {
		case "ResourceCreated":

			ok, err := locateElements(decoder, []string{"ReferenceParameters", "SelectorSet"})
			if nil != err {
				return nil, errors.New("locate 'Envelope/Body/ResourceCreated/ReferenceParameters/SelectorSet' failed, " + err.Error())
			}
			if !ok {
				return nil, ElementNotExists("Envelope/Body/ResourceCreated/ReferenceParameters/SelectorSet")
			}

			var id string
			for {
				nm, attrs, err := nextElement(decoder)
				if nil != err {
					return nil, errors.New("enumerate 'SelectorSet' failed, " + err.Error())
				}
				if "Selector" != nm.Local {
					return nil, errors.New("enumerate 'SelectorSet' failed, '" + nm.Local + "' is unknown.")
				}
				for _, attr := range attrs {
					if "ShellId" == attr.Value {
						if id, err = readXmlText(decoder); nil != err {
							return nil, errors.New("read 'ShellId' from the response failed, " + err.Error())
						}
						break
					}
				}
				if "" != id {
					break
				}
				exitElement(decoder, 0)
			}

			if "" == id {
				return nil, errors.New("ShellId is not found in the response.")
			}

			return &Shell{ep, id}, nil
		case "Shell":
			ok, err := locateElement(decoder, "ShellId")
			if nil != err {
				return nil, errors.New("locate 'Envelope/Body/Shell/ShellId' failed, " + err.Error())
			}
			if !ok {
				return nil, ElementNotExists("Envelope/Body/Shell/ShellId")
			}
			id, err := readXmlText(decoder)
			if nil != err {
				return nil, errors.New("read 'ShellId' from the response failed, " + err.Error())
			}
			return &Shell{ep, id}, nil
		default:
			if err = skipElement(decoder, 0); nil != err {
				return nil, err
			}
		}
	}
	return nil, errors.New("Envelope/Body/ResourceCreated or Envelope/Body/Shell isn't exists.")
}

func (s *Shell) NewCommand(cmd string, arguments []string) (string, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, len(cmd)))
	if e := xml.EscapeText(buf, []byte(cmd)); nil != e {
		return "", e
	}

	env := &envelope.CreateCommand{Uuid(), s.Id, strings.Replace(buf.String(), "&#34;", "&quot;", -1), arguments}
	reader, err := s.Deliver(bytes.NewBufferString(env.Xml()))
	if err != nil {
		return "", err
	}
	defer closeReader(reader)

	decoder := xml.NewDecoder(reader)

	if err := ReadEnvelopeBody(decoder); nil != err {
		return "", err
	}

	ok, err := locateElements(decoder, []string{"CommandResponse", "CommandId"})
	if nil != err {
		return "", errors.New("locate 'Envelope/Body/CommandResponse/CommandId' failed, " + err.Error())
	}
	if !ok {
		return "", ElementNotExists("Envelope/Body/CommandResponse/CommandId")
	}

	id, e := readXmlText(decoder)
	if nil != e {
		return "", errors.New("read CommandId from the response failed, " + e.Error())
	}

	return id, nil
}

type CommandResult struct {
	State    string
	ExitCode string
	Stdout   [][]byte
	Stderr   [][]byte
}

func (c *CommandResult) IsDone() bool {
	// https://msdn.microsoft.com/en-us/library/cc761137.aspx
	// http://schemas.microsoft.com/wbem/wsman/1/windows/shell/CommandState/Running
	return "http://schemas.microsoft.com/wbem/wsman/1/windows/shell/CommandState/Done" == c.State
}

func (s *Shell) Read(cmd_id string) (*CommandResult, error) {
	env := &envelope.Receive{Uuid(), s.Id, cmd_id}
	reader, err := s.Deliver(bytes.NewBufferString(env.Xml()))
	if err != nil {
		return nil, err
	}
	defer closeReader(reader)

	decoder := xml.NewDecoder(reader)

	if err := ReadEnvelopeResponse(decoder, "ReceiveResponse"); nil != err {
		return nil, err
	}

	var state, exitCode string
	var stdout, stderr [][]byte
	for {
		nm, attrs, e := nextElement(decoder)
		if nil != e {
			if ElementEndError == e {
				return &CommandResult{State: state, ExitCode: exitCode,
					Stdout: stdout, Stderr: stderr}, nil
			}
			return nil, errors.New("enumerate 'SelectorSet' failed, " + e.Error())
		}

		switch nm.Local {
		case "Stream":
			t, e := findBy(attrs, "Name")
			if nil != e {
				return nil, errors.New("name of Stream is not found, " + e.Error())
			}

			commandId, e := findBy(attrs, "CommandId")
			if nil != e {
				return nil, errors.New("name of CommandId is not found, " + e.Error())
			}
			if cmd_id != commandId {
				panic("muti command is not supported.")
			}

			txt, e := collectStream(decoder)
			if nil != e {
				return nil, errors.New("read Stream failed, " + e.Error())
			}
			if nil != txt {
				if "stdout" == t {
					stdout = append(stdout, txt)
				} else {
					stderr = append(stderr, txt)
				}
			}

		case "CommandState":
			commandId, e := findBy(attrs, "CommandId")
			if nil != e {
				return nil, errors.New("CommandId is not found in the CommandState, " + e.Error())
			}
			if cmd_id != commandId {
				panic("muti command is not supported.")
			}
			state, e = findBy(attrs, "State")
			if nil != e {
				return nil, errors.New("State is not found in the CommandState, " + e.Error())
			}

			exitCode, err = readCommandState(decoder)
			if nil != err {
				return nil, err
			}
		}
	}
}

func findBy(attrs []xml.Attr, nm string) (string, error) {
	for _, attr := range attrs {
		if nm == attr.Name.Local {
			return attr.Value, nil
		}
	}
	return "", errors.New("'" + nm + "' is not found.")
}

func collectStream(decoder *xml.Decoder) ([]byte, error) {
	txt, e := readXmlText(decoder)
	if nil != e {
		return nil, e
	}
	if len(txt) <= 0 {
		return nil, nil
	}

	return base64.StdEncoding.DecodeString(strings.TrimSpace(txt))
}

func readCommandState(decoder *xml.Decoder) (code string, err error) {
	for {
		nm, _, e := nextElement(decoder)
		if nil != e {
			if ElementEndError == e {
				return
			}

			err = errors.New("enumerate 'CommandState' failed, " + e.Error())
			return
		}

		switch nm.Local {
		case "ExitCode":
			code, err = readXmlText(decoder)
			if nil != err {
				err = errors.New("enumerate 'CommandState' failed, " + err.Error())
				return
			}
		default:
			err = errors.New("'" + nm.Local + "' is unknown element in the EnumerateResponse.")
			return
		}
	}
}

func (s *Shell) Send(cmd_id, txt string) error {
	txt = base64.StdEncoding.EncodeToString([]byte(txt))
	env := &envelope.Send{Uuid(), s.Id, cmd_id, txt}
	reader, err := s.Deliver(bytes.NewBufferString(env.Xml()))
	if err != nil {
		return err
	}
	closeReader(reader)
	// decoder := xml.NewDecoder(reader)
	// ok, err := locateElements(decoder, []string{"Envelope", "Body", "ResourceCreated",
	// 	"ReferenceParameters", "SelectorSet"})
	// if nil != err {
	// 	return nil, err
	// }
	// if !ok {
	// 	return nil, ElementNotExists("Envelope/Body/ResourceCreated/ReferenceParameters/SelectorSet")
	// }
	return nil
}

// https://msdn.microsoft.com/en-us/library/cc761132.aspx
const (
	SIGNAL_TERMINATE  = "http://schemas.microsoft.com/wbem/wsman/1/windows/shell/signal/terminate"
	SIGNAL_CTRL_C     = "http://schemas.microsoft.com/wbem/wsman/1/windows/shell/signal/ctrl_c"
	SIGNAL_CTRL_BREAK = "http://schemas.microsoft.com/wbem/wsman/1/windows/shell/signal/ctrl_break"
)

func (s *Shell) Signal(cmd_id, signal string) error {
	env := &envelope.Signal{Uuid(), s.Id, cmd_id, signal}
	reader, err := s.Deliver(bytes.NewBufferString(env.Xml()))
	if err != nil {
		return err
	}
	closeReader(reader)
	// decoder := xml.NewDecoder(reader)
	// ok, err := locateElements(decoder, []string{"Envelope", "Body", "ResourceCreated",
	// 	"ReferenceParameters", "SelectorSet"})
	// if nil != err {
	// 	return nil, err
	// }
	// if !ok {
	// 	return nil, ElementNotExists("Envelope/Body/ResourceCreated/ReferenceParameters/SelectorSet")
	// }
	return nil
}

func (s *Shell) Close() error {
	env := &envelope.DeleteShell{Uuid(), s.Id}
	reader, err := s.Deliver(bytes.NewBufferString(env.Xml()))
	if err != nil {
		return err
	}
	defer closeReader(reader)

	// decoder := xml.NewDecoder(reader)
	// ok, err := locateElements(decoder, []string{"Envelope", "Body", "ResourceCreated",
	// 	"ReferenceParameters", "SelectorSet"})
	// if nil != err {
	// 	return nil, err
	// }
	// if !ok {
	// 	return nil, ElementNotExists("Envelope/Body/ResourceCreated/ReferenceParameters/SelectorSet")
	// }

	return nil
}
