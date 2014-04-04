package wsman

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var simple_response = `
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:n="http://schemas.xmlsoap.org/ws/2004/09/enumeration" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xml:lang="zh-CN">
  <s:Header>
    <a:Action>http://schemas.xmlsoap.org/ws/2004/09/enumeration/EnumerateResponse</a:Action>
    <a:MessageID>uuid:3394F5B0-2BAF-48B8-A27B-130554C0AAD3</a:MessageID>
    <a:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:To>
    <a:RelatesTo>uuid:5743765A-973D-4D14-92A7-E7423A4C4455</a:RelatesTo>
  </s:Header>
  <s:Body>
    <n:EnumerateResponse>
      <n:EnumerationContext>uuid:CD02E6BD-C6F3-47F5-9AF5-6DCBE89A448E</n:EnumerationContext>
      <w:Items>
        <p:Win32_Process xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wmi/root/cimv2/Win32_Process" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cim="http://schemas.dmtf.org/wbem/wscim/1/common" xsi:type="p:Win32_Process_Type">
          <p:Caption>System Idle Process</p:Caption>
          <p:CommandLine xsi:nil="true"/>
        </p:Win32_Process>
      </w:Items>
      <n:EndOfSequence/>
    </n:EnumerateResponse>
  </s:Body>
</s:Envelope>`

var syntex_error_response = `
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:n="http://schemas.xmlsoap.org/ws/2004/09/enumeration" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xml:lang="zh-CN">
  <s:Header>
    <a:Action>http://schemas.xmlsoap.org/ws/2004/09/enumeration/EnumerateResponse</a:Action>
    <a:MessageID>uuid:3394F5B0-2BAF-48B8-A27B-130554C0AAD3</a:MessageID>
    <a:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:To>
    <a:RelatesTo>uuid:5743765A-973D-4D14-92A7-E7423A4C4455</a:RelatesTo>
  </s:Header>
  <s:Body>
    <n:EnumerateResponse>
      <n:EnumerationContext>uuid:CD02E6BD-C6F3-47F5-9AF5-6DCBE89A448E</n:EnumerationContext>
      <w:Items>
        <p:Win32_Process xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wmi/root/cimv2/Win32_Process" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cim="http://schemas.dmtf.org/wbem/wscim/1/common" xsi:type="p:Win32_Process_Type">
          <p:Caption>System Idle Process</p:Caption>
          <p:CommandLine xsi:nil="true"/>
        </p:Win32_Process>
      </n:Items>
      <n:EndOfSequence/>
    </n:EnumerateResponse>
  </s:Body>
</s:Envelope>`

var list_first_response = `
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:n="http://schemas.xmlsoap.org/ws/2004/09/enumeration" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xml:lang="zh-CN">
  <s:Header>
    <a:Action>http://schemas.xmlsoap.org/ws/2004/09/enumeration/EnumerateResponse</a:Action>
    <a:MessageID>uuid:3394F5B0-2BAF-48B8-A27B-130554C0AAD3</a:MessageID>
    <a:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:To>
    <a:RelatesTo>uuid:5743765A-973D-4D14-92A7-E7423A4C4455</a:RelatesTo>
  </s:Header>
  <s:Body>
    <n:EnumerateResponse>
      <n:EnumerationContext>uuid:CD02E6BD-C6F3-47F5-9AF5-6DCBE89A448E</n:EnumerationContext>
      <w:Items>
        <p:Win32_Process xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wmi/root/cimv2/Win32_Process" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cim="http://schemas.dmtf.org/wbem/wscim/1/common" xsi:type="p:Win32_Process_Type">
          <p:Caption>System Idle Process</p:Caption>
          <p:CommandLine xsi:nil="true"/>
        </p:Win32_Process>
        <p:Win32_Process xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wmi/root/cimv2/Win32_Process" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cim="http://schemas.dmtf.org/wbem/wscim/1/common" xsi:type="p:Win32_Process_Type">
          <p:Caption>System Idle Process</p:Caption>
          <p:CommandLine xsi:nil="true"/>
        </p:Win32_Process>
      </w:Items>
    </n:EnumerateResponse>
  </s:Body>
</s:Envelope>`

var list_end_response = `
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:n="http://schemas.xmlsoap.org/ws/2004/09/enumeration" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xml:lang="zh-CN">
  <s:Header>
    <a:Action>http://schemas.xmlsoap.org/ws/2004/09/enumeration/EnumerateResponse</a:Action>
    <a:MessageID>uuid:3394F5B0-2BAF-48B8-A27B-130554C0AAD3</a:MessageID>
    <a:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:To>
    <a:RelatesTo>uuid:5743765A-973D-4D14-92A7-E7423A4C4455</a:RelatesTo>
  </s:Header>
  <s:Body>
    <n:PullResponse>
      <n:EnumerationContext>uuid:CD02E6BD-C6F3-47F5-9AF5-6DCBE89A448E</n:EnumerationContext>
      <w:Items>
        <p:Win32_Process xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wmi/root/cimv2/Win32_Process" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cim="http://schemas.dmtf.org/wbem/wscim/1/common" xsi:type="p:Win32_Process_Type">
          <p:Caption>System Idle Process</p:Caption>
          <p:CommandLine xsi:nil="true"/>
        </p:Win32_Process>
        <p:Win32_Process xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wmi/root/cimv2/Win32_Process" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cim="http://schemas.dmtf.org/wbem/wscim/1/common" xsi:type="p:Win32_Process_Type">
          <p:Caption>System Idle Process</p:Caption>
          <p:CommandLine xsi:nil="true"/>
        </p:Win32_Process>
      </n:Items>
      <n:EndOfSequence/>
    </n:PullResponse>
  </s:Body>
</s:Envelope>`

func TestSimple(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CanHandle(r) {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, simple_response)
		}
	}))
	defer hsrv.Close()

	check := func() {
		it := Enumerate(&Endpoint{Url: hsrv.URL, User: "apd", Password: "123"}, "Win32_Process")
		count := 0
		for it.Next() {
			count++
			m, e := it.Map()
			if nil != e {
				t.Error(e)
				break
			}
			if "System Idle Process" != m["Caption"] {
				t.Error("value of 'Caption' is not excepted, actual is", m["Caption"])
			}
			v, ok := m["CommandLine"]
			if !ok {
				t.Error("'CommandLine' is not exists.")
			}
			if nil != v {
				t.Error("'CommandLine' is not equals nil -", v)
			}

			t.Log(m)
		}

		if 1 != count {
			t.Error("excepted count is 1, actual is ", count)
		}
		if nil != it.Err() {
			t.Error(it.Err())
		}
		it.Close()
	}

	WSMAN_DEBUG = false
	check()
	WSMAN_DEBUG = true
	check()
}

func TestErrorXml(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CanHandle(r) {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, syntex_error_response)
		}
	}))
	defer hsrv.Close()

	it := Enumerate(&Endpoint{Url: hsrv.URL, User: "apd", Password: "123"}, "Win32_Process")
	count := 0
	for it.Next() {
		count++
		m, e := it.Map()
		if nil != e {
			t.Error(e)
			break
		}
		if "System Idle Process" != m["Caption"] {
			t.Error("value of 'Caption' is not excepted, actual is", m["Caption"])
		}
		v, ok := m["CommandLine"]
		if !ok {
			t.Error("'CommandLine' is not exists.")
		}
		if nil != v {
			t.Error("'CommandLine' is not equals nil -", v)
		}

		t.Log(m)
	}
	if 1 != count {
		t.Error("excepted count is 1, actual is ", count)
	}

	if nil == it.Err() {
		t.Error("excepted is failed, actual is ok")
	} else if !strings.Contains(it.Err().Error(), "XML syntax error") {
		t.Error("excepted contains 'XML syntax error', actual is", it.Err())
	}
	it.Close()
}
