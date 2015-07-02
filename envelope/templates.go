package envelope

import (
	"bytes"
	"text/template"
)

type CreateShell struct {
	MessageId string
}

var create_shell_template = template.Must(template.New("CreateShell").Parse(CreateShellTemplate))

func (m *CreateShell) Xml() string {
	return applyTemplate(create_shell_template, m)
}

type DeleteShell struct {
	MessageId string
	ShellId   string
}

var delete_shell_template = template.Must(template.New("DeleteShell").Parse(DeleteShellTemplate))

func (m *DeleteShell) Xml() string {
	return applyTemplate(delete_shell_template, m)
}

type CreateCommand struct {
	MessageId   string
	ShellId     string
	CommandText string
}

var create_template = template.Must(template.New("CreateCommand").Parse(CreateCommandTemplate))

func (m *CreateCommand) Xml() string {
	return applyTemplate(create_template, m)
}

type Send struct {
	MessageId string
	ShellId   string
	CommandId string
	Content   string
}

var send_template = template.Must(template.New("Send").Parse(SendTemplate))

func (m *Send) Xml() string {
	return applyTemplate(send_template, m)
}

type Receive struct {
	MessageId string
	ShellId   string
	CommandId string
}

var receive_template = template.Must(template.New("Receive").Parse(ReceiveTemplate))

func (m *Receive) Xml() string {
	return applyTemplate(receive_template, m)
}

type Signal struct {
	MessageId string
	ShellId   string
	CommandId string
}

var signal_template = template.Must(template.New("Signal").Parse(SignalTemplate))

func (m *Signal) Xml() string {
	return applyTemplate(signal_template, m)
}

type Enumerate struct {
	Namespace   string
	MessageId   string
	Name        string
	SelectorSet map[string]string
}

var enumerate_template = template.Must(template.New("Enumerate").Parse(EnumerateTemplate))

func (m *Enumerate) Xml() string {
	return applyTemplate(enumerate_template, m)
}

type Pull struct {
	Namespace   string
	MessageId   string
	Name        string
	SelectorSet map[string]string
	Context     string
}

var pull_template = template.Must(template.New("Pull").Parse(PullTemplate))

func (m *Pull) Xml() string {
	return applyTemplate(pull_template, m)
}

type Get struct {
	Namespace   string
	MessageId   string
	Name        string
	SelectorSet map[string]string
}

var get_template = template.Must(template.New("Get").Parse(GetTemplate))

func (m *Get) Xml() string {
	return applyTemplate(get_template, m)
}

func applyTemplate(t *template.Template, data interface{}) string {
	var b bytes.Buffer
	err := t.Execute(&b, data)
	if err != nil {
		panic(err)
	}
	return b.String()
}

const CreateShellTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:rsp="` + NS_WIN_SHELL + `" xmlns:w="` + NS_WSMAN_DMTF + `">
  <s:Header>
    <a:To>http://localhost:5985/wsman</a:To>
    <a:ReplyTo>
      <a:Address mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
    <w:ResourceURI mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd</w:ResourceURI>
    <a:Action mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/transfer/Create</a:Action>
    <w:OptionSet>
      <w:Option Name="WINRS_NOPROFILE">FALSE</w:Option>
      <w:Option Name="WINRS_CODEPAGE">437</w:Option>
    </w:OptionSet>
  </s:Header>
  <s:Body>
    <rsp:Shell>
      <rsp:InputStreams>stdin</rsp:InputStreams>
      <rsp:OutputStreams>stdout stderr</rsp:OutputStreams>
    </rsp:Shell>
  </s:Body>
</s:Envelope>`

const DeleteShellTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:w="` + NS_WSMAN_DMTF + `">
  <s:Header>    
    <a:Action mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/transfer/Delete</a:Action>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <w:ResourceURI mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd</w:ResourceURI>
    <w:SelectorSet>
      <w:Selector Name="ShellId">{{.ShellId}}</w:Selector>
    </w:SelectorSet>
    <a:To>http://localhost:5985/wsman</a:To>
    <a:ReplyTo>
      <a:Address mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <w:OperationTimeout>PT60S</w:OperationTimeout>

  </s:Header>
  <s:Body/>
</s:Envelope>`

const CreateCommandTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:rsp="` + NS_WIN_SHELL + `" xmlns:w="` + NS_WSMAN_DMTF + `">
  <s:Header>
    <a:Action mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Command</a:Action>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <w:ResourceURI mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd</w:ResourceURI>
    <w:SelectorSet>
      <w:Selector Name="ShellId">{{.ShellId}}</w:Selector>
    </w:SelectorSet>
    <a:To>http://localhost:5985/wsman</a:To>
    <a:ReplyTo>
      <a:Address mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
    <w:OptionSet>
      <w:Option Name="WINRS_CONSOLEMODE_STDIN">TRUE</w:Option>
      <w:Option Name="WINRS_SKIP_CMD_SHELL">FALSE</w:Option>
    </w:OptionSet>
  </s:Header>
  <s:Body>
    <rsp:CommandLine>
      <rsp:Command>{{.CommandText}}</rsp:Command>
    </rsp:CommandLine>
  </s:Body>
</s:Envelope>`

const SendTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:rsp="` + NS_WIN_SHELL + `" xmlns:w="` + NS_WSMAN_DMTF + `">
    <s:Header>
      <a:Action s:mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Send</a:Action>
      <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
      <w:ResourceURI>http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd</w:ResourceURI>
      <w:SelectorSet xmlns="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd">
        <w:Selector Name="ShellId">{{.ShellId}}</w:Selector>
      </w:SelectorSet>
      <a:To>http://localhost:5985/wsman</a:To>
      <a:ReplyTo>
        <a:Address s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
      </a:ReplyTo>
      <w:MaxEnvelopeSize s:mustUnderstand="true">153600</w:MaxEnvelopeSize>
      <w:OperationTimeout>PT60S</w:OperationTimeout>
    </s:Header>
    <s:Body>
      <rsp:Send xmlns:rsp="http://schemas.microsoft.com/wbem/wsman/1/windows/shell">
        <rsp:Stream Name="stdin" CommandId="{{.CommandId}}">{{.Content}}</rsp:Stream>
      </rsp:Send>
    </s:Body>
  </s:Envelope>`

const ReceiveTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:rsp="` + NS_WIN_SHELL + `" xmlns:w="` + NS_WSMAN_DMTF + `">
  <s:Header>
    <a:Action mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Receive</a:Action>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <w:ResourceURI mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd</w:ResourceURI>
    <w:SelectorSet>
      <w:Selector Name="ShellId">{{.ShellId}}</w:Selector>
    </w:SelectorSet>
    <a:To>http://localhost:5985/wsman</a:To>
    <a:ReplyTo>
      <a:Address mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
  </s:Header>
  <s:Body>
    <rsp:Receive>
      <rsp:DesiredStream CommandId="{{.CommandId}}">stdout stderr</rsp:DesiredStream>
    </rsp:Receive>
  </s:Body>
</s:Envelope>`

const SignalTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:rsp="` + NS_WIN_SHELL + `" xmlns:w="` + NS_WSMAN_DMTF + `">
  <s:Header>
    <a:Action s:mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Signal</a:Action>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <w:ResourceURI>http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd</w:ResourceURI>
    <w:SelectorSet xmlns="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd">
      <w:Selector Name="ShellId">{{.ShellId}}</w:Selector>
    </w:SelectorSet>
    <a:To>http://localhost:5985/wsman</a:To>
    <a:ReplyTo>
      <a:Address s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize s:mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
  </s:Header>
  <s:Body>
    <rsp:Signal xmlns:rsp="http://schemas.microsoft.com/wbem/wsman/1/windows/shell" CommandId="{{.CommandId}}">
      <rsp:Code>http://schemas.microsoft.com/wbem/wsman/1/windows/shell/signal/terminate</rsp:Code>
    </rsp:Signal>
  </s:Body>
</s:Envelope>`

const EnumerateTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:p="` + NS_WSMAN_MSFT + `" xmlns:w="` + NS_WSMAN_DMTF + `" xmlns:n="` + NS_ENUM + `">
  <s:Header>
    <a:Action s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/enumeration/Enumerate</a:Action>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <a:To>http://localhost:5985/wsman</a:To>
    <w:ResourceURI s:mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/wmi/{{.Namespace}}/{{.Name}}</w:ResourceURI>
    {{if .SelectorSet}}<w:SelectorSet>
      {{range $key, $value := .SelectorSet}}<w:Selector Name="{{$key}}">{{$value}}</w:Selector>{{end}}
    </w:SelectorSet>{{end}}
    <a:ReplyTo>
      <a:Address s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize s:mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
  </s:Header>
  <s:Body>
    <n:Enumerate>
      <w:OptimizeEnumeration/>
      <w:MaxElements>32000</w:MaxElements>
    </n:Enumerate>
  </s:Body>
</s:Envelope>`

const PullTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:p="` + NS_WSMAN_MSFT + `" xmlns:w="` + NS_WSMAN_DMTF + `" xmlns:n="` + NS_ENUM + `">
  <s:Header>
    <a:Action s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/enumeration/Pull</a:Action>
    <a:To>http://localhost:5985/wsman</a:To>
    <w:ResourceURI s:mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/wmi/{{.Namespace}}/{{.Name}}</w:ResourceURI>
    {{if .SelectorSet}}<w:SelectorSet>
      {{range $key, $value := .SelectorSet}}<w:Selector Name="{{$key}}">{{$value}}</w:Selector>{{end}}
    </w:SelectorSet>{{end}}
    <a:ReplyTo>
      <a:Address s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize s:mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
  </s:Header>
  <s:Body>
    <n:Pull>
      <n:EnumerationContext>{{.Context}}</n:EnumerationContext>
      <n:MaxElements>32000</n:MaxElements>
    </n:Pull>
  </s:Body>
</s:Envelope>`

const GetTemplate = `
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd">
  <s:Header>
    <a:Action s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/transfer/Get</a:Action>
    <a:To>http://localhost:5985/wsman</a:To>
    <w:ResourceURI s:mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/wmi/{{.Namespace}}/{{.Name}}</w:ResourceURI>
    {{if .SelectorSet}}<w:SelectorSet>
      {{range $key, $value := .SelectorSet}}<w:Selector Name="{{$key}}">{{$value}}</w:Selector>{{end}}
      </w:SelectorSet>{{end}}
    <a:ReplyTo>
      <a:Address s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize s:mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
  </s:Header>
  <s:Body/>
</s:Envelope>
`
