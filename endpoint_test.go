package wsman

import (
	"encoding/xml"
	"strings"

	"testing"
)

func TestReadEnvelopeFaultWithS12(t *testing.T) {
	txt := `<s:Envelope xmlns:s="http://www.w3.org/2004/08/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:n="http://schemas.xmlsoap.org/ws/2004/09/enumeration" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xml:lang="zh-CN">
<s:Envelope> 
 <s:Header>
   <wsa:Action>
     http://schemas.xmlsoap.org/ws/2004/08/addressing/fault
   </wsa:Action>
   <!-- Headers elided for clarity.  -->
 </s:Header>
 <s:Body>
  <s:Fault>
   <s:Code>
     <s:Value>[Code]</s:Value>
     <s:Subcode>
      <s:Value>[Subcode]</s:Value>
     </s:Subcode>
   </s:Code>
   <s:Reason>
     <s:Text xml:lang="en">[Reason]</s:Text>
   </s:Reason>
   <s:Detail>
     [Detail]
   </s:Detail>    
  </s:Fault>
 </s:Body>
</s:Envelope>`

	decoder := xml.NewDecoder(strings.NewReader(txt))
	e := ReadEnvelopeBody(decoder)
	if nil == e {
		t.Error("error is nil")
		return
	}

	fault, ok := e.(*ErrSoapFault)
	if !ok {
		t.Error(e)
		return
	}

	if "[Code]" != fault.Code {
		t.Error("code is error -", fault.Code)
	}
	if "[Subcode]" != fault.Subcode {
		t.Error("Subcode is error -", fault.Subcode)
	}
	if "[Detail]" != fault.Detail {
		t.Error("Detail is error -", fault.Detail)
	}
	if "[Reason]" != fault.Reason {
		t.Error("Reason is error -", fault.Reason)
	}
}

func TestReadEnvelopeFaultWithS11(t *testing.T) {
	txt := `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:n="http://schemas.xmlsoap.org/ws/2004/09/enumeration" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xml:lang="zh-CN">
<s:Envelope> 
 <s:Header>
   <wsa:Action>
     http://schemas.xmlsoap.org/ws/2004/08/addressing/fault
   </wsa:Action>
   <!-- Headers elided for clarity.  -->
 </s:Header>
 <s:Body>
  <s:Fault>
   <faultcode>[Subcode]</faultcode>
   <faultstring xml:lang="en">[Reason]</faultstring>   
  </s:Fault>
 </s:Body>
</s:Envelope>`

	decoder := xml.NewDecoder(strings.NewReader(txt))
	e := ReadEnvelopeBody(decoder)
	if nil == e {
		t.Error("error is nil")
		return
	}

	fault, ok := e.(*ErrSoapFault)
	if !ok {
		t.Error(e)
		return
	}

	if "[Subcode]" != fault.Code {
		t.Error("code is error -", fault.Code)
	}
	if "[Subcode]" != fault.Subcode {
		t.Error("Subcode is error -", fault.Subcode)
	}
	if "[Reason]" != fault.Detail {
		t.Error("Detail is error -", fault.Detail)
	}
	if "[Reason]" != fault.Reason {
		t.Error("Reason is error -", fault.Reason)
	}
}
