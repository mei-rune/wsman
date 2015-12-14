package envelope

import (
	"bytes"
	"text/template"
)

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

var create_shell_template = template.Must(template.New("CreateShell").Parse(CreateShellTemplate))

type CreateShell struct {
	MessageId string
}

func (m *CreateShell) Xml() string {
	return applyTemplate(create_shell_template, m)
}

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

var delete_shell_template = template.Must(template.New("DeleteShell").Parse(DeleteShellTemplate))

type DeleteShell struct {
	MessageId string
	ShellId   string
}

func (m *DeleteShell) Xml() string {
	return applyTemplate(delete_shell_template, m)
}

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
      <rsp:Command>{{.CommandText}}</rsp:Command>{{range $value := .Arguments}}
      <rsp:Arguments>{{$value}}</rsp:Arguments>{{end}}
    </rsp:CommandLine>
  </s:Body>
</s:Envelope>`

var create_template = template.Must(template.New("CreateCommand").Parse(CreateCommandTemplate))

type CreateCommand struct {
	MessageId   string
	ShellId     string
	CommandText string
	Arguments   []string
}

func (m *CreateCommand) Xml() string {
	return applyTemplate(create_template, m)
}

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

var send_template = template.Must(template.New("Send").Parse(SendTemplate))

type Send struct {
	MessageId string
	ShellId   string
	CommandId string
	Content   string
}

func (m *Send) Xml() string {
	return applyTemplate(send_template, m)
}

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

var receive_template = template.Must(template.New("Receive").Parse(ReceiveTemplate))

type Receive struct {
	MessageId string
	ShellId   string
	CommandId string
}

func (m *Receive) Xml() string {
	return applyTemplate(receive_template, m)
}

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
      <rsp:Code>{{.Value}}</rsp:Code>
    </rsp:Signal>
  </s:Body>
</s:Envelope>`

var signal_template = template.Must(template.New("Signal").Parse(SignalTemplate))

type Signal struct {
	MessageId string
	ShellId   string
	CommandId string
	Value     string
}

func (m *Signal) Xml() string {
	return applyTemplate(signal_template, m)
}

const EnumerateTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:p="` + NS_WSMAN_MSFT + `" xmlns:w="` + NS_WSMAN_DMTF + `" xmlns:n="` + NS_ENUM + `">
  <s:Header>
    <a:Action s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/enumeration/Enumerate</a:Action>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <a:To>http://localhost:5985/wsman</a:To>
    <w:ResourceURI s:mustUnderstand="true">{{.Namespace}}/{{.Name}}</w:ResourceURI>
    {{if .SelectorSet}}<w:SelectorSet>
      {{range $key, $value := .SelectorSet}}<w:Selector Name="{{$key}}">{{$value}}</w:Selector>{{end}}
    </w:SelectorSet>{{end}}
    <a:ReplyTo>
      <a:Address s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize s:mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
    {{if .OptionSet}}<w:OptionSet>
      {{range $key, $value := .OptionSet}}<w:Option Name="{{$key}}">{{$value}}</w:Option>{{end}}
    </w:OptionSet>
  {{end}}</s:Header>
  <s:Body>
    <n:Enumerate>
      <w:OptimizeEnumeration/>
      <w:MaxElements>200</w:MaxElements>
    </n:Enumerate>
  </s:Body>
</s:Envelope>`

var enumerate_template = template.Must(template.New("Enumerate").Parse(EnumerateTemplate))

type Enumerate struct {
	Namespace   string
	MessageId   string
	Name        string
	SelectorSet map[string]string
	OptionSet   map[string]string
}

func (m *Enumerate) Xml() string {
	return applyTemplate(enumerate_template, m)
}

const PullTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:p="` + NS_WSMAN_MSFT + `" xmlns:w="` + NS_WSMAN_DMTF + `" xmlns:n="` + NS_ENUM + `">
  <s:Header>
    <a:Action s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/enumeration/Pull</a:Action>
    <a:To>http://localhost:5985/wsman</a:To>
    <w:ResourceURI s:mustUnderstand="true">{{.Namespace}}/{{.Name}}</w:ResourceURI>
    {{if .SelectorSet}}<w:SelectorSet>
      {{range $key, $value := .SelectorSet}}<w:Selector Name="{{$key}}">{{$value}}</w:Selector>{{end}}
    </w:SelectorSet>{{end}}
    <a:ReplyTo>
      <a:Address s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize s:mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <w:OperationTimeout>{{if eq .Timeout 0}}PT60S{{else}}PT{{.Timeout}}S{{end}}</w:OperationTimeout>
    {{if .OptionSet}}<w:OptionSet>
      {{range $key, $value := .OptionSet}}<w:Option Name="{{$key}}">{{$value}}</w:Option>{{end}}
    </w:OptionSet>
  {{end}}</s:Header>
  <s:Body>
    <n:Pull>
      <n:EnumerationContext>{{.Context}}</n:EnumerationContext>
      <n:MaxElements>200</n:MaxElements>
    </n:Pull>
  </s:Body>
</s:Envelope>`

var pull_template = template.Must(template.New("Pull").Parse(PullTemplate))

type Pull struct {
	Namespace   string
	MessageId   string
	Name        string
	SelectorSet map[string]string
	Context     string
	Timeout     uint
	OptionSet   map[string]string
}

func (m *Pull) Xml() string {
	return applyTemplate(pull_template, m)
}

const GetTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:p="` + NS_WSMAN_MSFT + `" xmlns:w="` + NS_WSMAN_DMTF + `" xmlns:n="` + NS_ENUM + `">
  <s:Header>
    <a:Action s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/transfer/Get</a:Action>
    <a:To>http://localhost:5985/wsman</a:To>
    <w:ResourceURI s:mustUnderstand="true">{{.Namespace}}/{{.Name}}</w:ResourceURI>
    {{if .SelectorSet}}<w:SelectorSet>
      {{range $key, $value := .SelectorSet}}<w:Selector Name="{{$key}}">{{$value}}</w:Selector>{{end}}
      </w:SelectorSet>{{end}}
    <a:ReplyTo>
      <a:Address s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize s:mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
    {{if .OptionSet}}<w:OptionSet xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
      {{range $key, $value := .OptionSet}}<w:Option Name="{{$key}}">{{$value}}</w:Option>{{end}}
    </w:OptionSet>
  {{end}}</s:Header>
  <s:Body/>
</s:Envelope>
`

var get_template = template.Must(template.New("Get").Parse(GetTemplate))

type Get struct {
	Namespace   string
	MessageId   string
	Name        string
	SelectorSet map[string]string
	OptionSet   map[string]string
}

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

const SubscribeTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:w="` + NS_WSMAN_DMTF + `" xmlns:wse="` + NS_EVENTING + `">
  <s:Header>
    <a:To>http://localhost:5985/wsman</a:To>
    <w:ResourceURI s:mustUnderstand="true">{{.Namespace}}/{{.Name}}</w:ResourceURI>
    {{if .SelectorSet}}<w:SelectorSet>
      {{range $key, $value := .SelectorSet}}<w:Selector Name="{{$key}}">{{$value}}</w:Selector>{{end}}
    </w:SelectorSet>
    {{end}}<a:ReplyTo>
      <a:Address s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <a:Action s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/eventing/Subscribe</a:Action>
    <w:MaxEnvelopeSize s:mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
    <w:OptionSet xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
      <w:Option Name="SubscriptionName">hw_subscription</w:Option>
      <w:Option Name="ContentFormat">RenderedText</w:Option>
      <w:Option Name="IgnoreChannelError" xsi:nil="true"/>
    </w:OptionSet> 
  </s:Header>
  <s:Body>
   <wse:Subscribe>{{if ne  .DeliveryMode "http://schemas.dmtf.org/wbem/wsman/1/wsman/Pull"}}<wse:EndTo>
      <a:Address>{{.EndToAddress}}</a:Address>
      <a:ReferenceProperties>
        <wse:Identifier>{{.EndToIdentifier}}</wse:Identifier>
      </a:ReferenceProperties>
    </wse:EndTo>{{end}}{{if eq  .DeliveryMode "http://schemas.dmtf.org/wbem/wsman/1/wsman/Pull"}}
    <wse:Delivery Mode="http://schemas.dmtf.org/wbem/wsman/1/wsman/Pull"/>
    {{else}}<wse:Delivery Mode="{{.DeliveryMode}}">
      <w:Heartbeats>PT300S</w:Heartbeats>
      <wse:NotifyTo>
        <a:Address>{{.RecvAddress}}</a:Address>
        <a:ReferenceProperties>
          <wse:Identifier>{{.RecvIdentifier}}</wse:Identifier>
        </a:ReferenceProperties>
      </wse:NotifyTo>
      {{if eq  .DeliveryMode "http://schemas.dmtf.org/wbem/wsman/1/wsman/Events"}}
      <w:MaxElements>200</w:MaxElements>
      <w:MaxTime>PT30S</w:MaxTime>
      <w:MaxEnvelopeSize Policy="Notify">1536000</w:MaxEnvelopeSize>
      <w:ContentEncoding>UTF-8</w:ContentEncoding>
      <w:ConnectionRetry Total="3">PT180S</w:ConnectionRetry>{{end}}
    </wse:Delivery>{{end}}
    <wse:Expires>PT300S</wse:Expires>
    {{if .QueryList}}<w:Filter Dialect="http://schemas.microsoft.com/win/2004/08/events/eventquery">
      <QueryList>
        {{range $key, $value := .QueryList}}<Query Id="{{$key}}">
        {{range $pvalue := $value}}<Select Path="{{$pvalue.Path}}">{{$pvalue.Value}}</Select>{{end}}
        </Query>{{end}}
      </QueryList>
    </w:Filter>{{end}}
    {{if .SendBookmarks}}<w:SendBookmarks/>{{end}}
    </wse:Subscribe>
  </s:Body>
  </s:Envelope>`

var subscribe_template = template.Must(template.New("Subscribe").Parse(SubscribeTemplate))

const (
	DELIVERYMODE_XMLSOAP_PUSH        = `http://schemas.xmlsoap.org/ws/2004/08/eventing/DeliveryModes/Push`
	DELIVERYMODE_WSMAN_PUSH_WITH_ACK = `http://schemas.dmtf.org/wbem/wsman/1/wsman/PushWithAck`
	DELIVERYMODE_WSMAN_EVENTS        = `http://schemas.dmtf.org/wbem/wsman/1/wsman/Events`
	DELIVERYMODE_WSMAN_PULL          = `http://schemas.dmtf.org/wbem/wsman/1/wsman/Pull`
)

type Subscribe struct {
	Namespace   string
	MessageId   string
	Name        string
	SelectorSet map[string]string

	DeliveryMode    string
	EndToAddress    string
	EndToIdentifier string
	RecvAddress     string
	RecvIdentifier  string
	QueryList       map[string][]QueryFilter
	SendBookmarks   bool
}

type QueryFilter struct {
	Path  string
	Value string
}

func (m *Subscribe) Xml() string {
	return applyTemplate(subscribe_template, m)
}

const UnsubscribeTemplate = `<s:Envelope xmlns:s="` + NS_SOAP_ENV + `" xmlns:a="` + NS_ADDRESSING + `" xmlns:w="` + NS_WSMAN_DMTF + `" xmlns:wse="` + NS_EVENTING + `">
  <s:Header>
    <a:To>http://localhost:5985/wsman</a:To>
    <w:ResourceURI s:mustUnderstand="true">{{.Namespace}}/{{.Name}}</w:ResourceURI>
    {{if .SelectorSet}}<w:SelectorSet>
    {{range $key, $value := .SelectorSet}}<w:Selector Name="{{$key}}">{{$value}}</w:Selector>{{end}}
    </w:SelectorSet>{{end}}
    <a:ReplyTo>
      <a:Address s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <a:Action s:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/eventing/Unsubscribe</a:Action>
    <w:MaxEnvelopeSize s:mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <a:MessageID>uuid:{{.MessageId}}</a:MessageID>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
  </s:Header>
  <s:Body>
    <wse:Unsubscribe>
    </wse:Unsubscribe>
  </s:Body>
 </s:Envelope>`
