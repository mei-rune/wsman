package wsman

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/runner-mei/wsman/envelope"
)

var simple_enumerate_response = `
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
          <p:CreationDate>
            <cim:Datetime>2014-04-03T14:45:50.46875+08:00</cim:Datetime>
          </p:CreationDate>
          <p:AutomaticStartupActionDelay>
              <cim:Interval>P0DT0H1M40S</cim:Interval>
          </p:AutomaticStartupActionDelay>
          <p:IPAddress>192.168.1.103</p:IPAddress>
          <p:IPAddress>fe80::6962:c157:2318:423d</p:IPAddress>
          <p:IPAddress>192.168.1.102</p:IPAddress>
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
      <n:EnumerationContext>uuid:CD02E6BE-C6F3-47FD-9AF5-6DCBE89A448E</n:EnumerationContext>
      <w:Items>
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
      <n:EnumerationContext>uuid:CD02E6BF-C6F3-47F5-9AF5-6DCBE89A448E</n:EnumerationContext>
      <w:Items>
        <p:Win32_Process xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wmi/root/cimv2/Win32_Process" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cim="http://schemas.dmtf.org/wbem/wscim/1/common" xsi:type="p:Win32_Process_Type">
          <p:Caption>SystemTest</p:Caption>
          <p:CommandLine xsi:nil="true"/>
        </p:Win32_Process>
      </w:Items>
      <n:EndOfSequence/>
    </n:PullResponse>
  </s:Body>
</s:Envelope>`

var simple_get_response = `
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:x="http://schemas.xmlsoap.org/ws/2004/09/transfer" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd" xml:lang="zh-CN">
  <s:Header>
    <a:Action>http://schemas.xmlsoap.org/ws/2004/09/transfer/GetResponse</a:Action>
    <a:MessageID>uuid:84066B11-A269-4F9B-BFFD-966CD033F69A</a:MessageID>
    <a:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:To>
    <a:RelatesTo>uuid:33CB26B4-2992-4DF9-BE76-13A06F311DA1</a:RelatesTo>
  </s:Header>
  <s:Body>
    <p:Win32_OperatingSystem xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wmi/root/cimv2/Win32_OperatingSystem" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cim="http://schemas.dmtf.org/wbem/wscim/1/common" xsi:type="p:Win32_OperatingSystem_Type">
      <p:BootDevice>\Device\HarddiskVolume2</p:BootDevice>
      <p:BuildNumber>7601</p:BuildNumber>
      <p:BuildType>Multiprocessor Free</p:BuildType>
    </p:Win32_OperatingSystem>
  </s:Body>
</s:Envelope>`

const enumerateResponseWithAddressing = `<s:Envelope xml:lang="zh-CN" xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:n="http://schemas.xmlsoap.org/ws/2004/09/enumeration" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd">
    <s:Header>
        <a:Action>http://schemas.xmlsoap.org/ws/2004/09/enumeration/EnumerateResponse</a:Action>
        <a:MessageID>uuid:B82A9456-26E4-487B-BED6-A95736911AB0</a:MessageID>
        <a:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:To>
        <a:RelatesTo>uuid:9776229c-9e00-4f22-c7a9-86664c90959a</a:RelatesTo>
    </s:Header>
    <s:Body>
        <n:EnumerateResponse>
            <n:EnumerationContext>uuid:40DF132C-91C8-4A4C-B99B-E8012432DA5D</n:EnumerationContext>
            <w:Items>
                <p:Msvm_SettingsDefineCapabilities xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wmi/root/virtualization/Msvm_SettingsDefineCapabilities" xmlns:cim="http://schemas.dmtf.org/wbem/wscim/1/common" xsi:type="p:Msvm_SettingsDefineCapabilities_Type">
                    <p:GroupComponent xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd">
                        <a:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
                        <a:ReferenceParameters>
                            <w:ResourceURI>http://schemas.microsoft.com/wbem/wsman/1/wmi/root/virtualization/Msvm_AllocationCapabilities</w:ResourceURI>
                            <w:SelectorSet>
                                <w:Selector Name="InstanceID">Microsoft:06FF76FA-2D58-4BAF-9F8D-455773824F37</w:Selector>
                            </w:SelectorSet>
                        </a:ReferenceParameters>
                    </p:GroupComponent>
                    <p:PartComponent xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd">
                        <a:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
                        <a:ReferenceParameters>
                            <w:ResourceURI>http://schemas.microsoft.com/wbem/wsman/1/wmi/root/virtualization/Msvm_Synthetic3DDisplayControllerSettingData</w:ResourceURI>
                            <w:SelectorSet>
                                <w:Selector Name="InstanceID">Microsoft:Definition\06FF76FA-2D58-4BAF-9F8D-455773824F37\Default</w:Selector>
                            </w:SelectorSet>
                        </a:ReferenceParameters>
                    </p:PartComponent>
                    <p:PropertyPolicy>0</p:PropertyPolicy>
                    <p:ValueRange>0</p:ValueRange>
                    <p:ValueRole>0</p:ValueRole>
                </p:Msvm_SettingsDefineCapabilities>
            </w:Items>
            <n:EndOfSequence/>
        </n:EnumerateResponse>
    </s:Body>
</s:Envelope>`

func TestEnumerateSimple(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CanHandle(r) {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, simple_enumerate_response)
		}
	}))
	defer hsrv.Close()

	check := func() {
		it := Enumerate(&Endpoint{Url: hsrv.URL, User: "apd", Password: "123"}, envelope.NS_WMI_CIMV2, "Win32_Process", nil)
		count := 0
		for it.Next() {
			count++
			m, e := it.Value()
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

			if "2014-04-03T14:45:50.46875+08:00" != m["CreationDate"] {
				t.Error("value of 'CreationDate' is not excepted, actual is", m["CreationDate"])
			}

			v, ok = m["AutomaticStartupActionDelay"]
			if !ok || nil == v {
				t.Error("'AutomaticStartupActionDelay' is not exists or nil.")
				return
			}
			if duration, ok := v.(time.Duration); !ok {
				t.Error("'AutomaticStartupActionDelay' is not a duration.")
				return
			} else if duration != 1*time.Minute+40*time.Second {
				t.Error("'AutomaticStartupActionDelay' except 0 got", duration)
				return
			}

			v, ok = m["IPAddress"]
			if !ok || nil == v {
				t.Error("'IPAddress' is not exists or nil.")
				return
			}
			ss, ok := v.([]interface{})
			if !ok {
				t.Errorf("'IPAddress' is not a []interface{}, actual is [%T] %v", v, v)
				return
			}
			if 3 != len(ss) {
				t.Error("count of 'IPAddress' is not 3, actual is ", ss)
			} else {
				if !reflect.DeepEqual("192.168.1.103", ss[0]) {
					t.Error("IPAddress[0] is not equals '192.168.1.103'", ss[0])
				}
				if !reflect.DeepEqual("fe80::6962:c157:2318:423d", ss[1]) {
					t.Error("IPAddress[1] is not equals 'fe80::6962:c157:2318:423d'", ss[0])
				}
				if !reflect.DeepEqual("192.168.1.102", ss[2]) {
					t.Error("IPAddress[2] is not equals '192.168.1.103'", ss[2])
				}
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

func TestEnumerateWithAddressing(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CanHandle(r) {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, enumerateResponseWithAddressing)
		}
	}))
	defer hsrv.Close()

	check := func() {
		it := Enumerate(&Endpoint{Url: hsrv.URL, User: "apd", Password: "123"}, envelope.NS_WMI+"/root/virtualization", "Msvm_SettingsDefineCapabilities", nil)
		count := 0
		for it.Next() {
			count++
			m, e := it.Value()
			if nil != e {
				t.Error(e)
				break
			}
			o := m["GroupComponent"]
			groupComponent, _ := o.(*Reference)
			if groupComponent == nil {
				t.Error("excepted GroupComponent got nil")
				break
			}
			if groupComponent.Address != "http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous" {
				t.Error("GroupComponent.Address excepted \"http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous\" got", groupComponent.Address)
				break
			}
			if groupComponent.ResourceURI != "http://schemas.microsoft.com/wbem/wsman/1/wmi/root/virtualization/Msvm_AllocationCapabilities" {
				t.Error("GroupComponent.ResourceURI excepted \"http://schemas.microsoft.com/wbem/wsman/1/wmi/root/virtualization/Msvm_AllocationCapabilities\" got", groupComponent.Address)
				break
			}

			if len(groupComponent.SelectorSet) != 1 {
				t.Error("len(GroupComponent.SelectorSet) except 1 got", len(groupComponent.SelectorSet))
				t.Log(groupComponent.SelectorSet)
				break
			}
			if groupComponent.SelectorSet["InstanceID"] != "Microsoft:06FF76FA-2D58-4BAF-9F8D-455773824F37" {
				t.Error("GroupComponent.SelectorSet[\"InstanceID\"] except \"Microsoft:06FF76FA-2D58-4BAF-9F8D-455773824F37\" got", groupComponent.SelectorSet["InstanceID"])
				t.Log(groupComponent.SelectorSet)
				break
			}
			o = m["PartComponent"]
			partComponent, _ := o.(*Reference)
			if partComponent == nil {
				t.Error("excepted PartComponent got nil")
				break
			}
			if partComponent.Address != "http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous" {
				t.Error("GroupComponent.Address excepted \"http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous\" got", partComponent.Address)
				break
			}
			if partComponent.ResourceURI != "http://schemas.microsoft.com/wbem/wsman/1/wmi/root/virtualization/Msvm_Synthetic3DDisplayControllerSettingData" {
				t.Error("GroupComponent.ResourceURI excepted \"http://schemas.microsoft.com/wbem/wsman/1/wmi/root/virtualization/Msvm_Synthetic3DDisplayControllerSettingData\" got", partComponent.Address)
				break
			}

			if len(partComponent.SelectorSet) != 1 {
				t.Error("len(GroupComponent.SelectorSet) except 1 got", len(partComponent.SelectorSet))
				t.Log(partComponent.SelectorSet)
				break
			}
			if partComponent.SelectorSet["InstanceID"] != "Microsoft:Definition\\06FF76FA-2D58-4BAF-9F8D-455773824F37\\Default" {
				t.Error("GroupComponent.SelectorSet[\"InstanceID\"] except \"Microsoft:Definition\\06FF76FA-2D58-4BAF-9F8D-455773824F37\\Default\" got", partComponent.SelectorSet["InstanceID"])
				t.Log(partComponent.SelectorSet)
				break
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
	// WSMAN_DEBUG = true
	// check()
}

func TestGetSimple(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CanHandle(r) {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, simple_get_response)
		}
	}))
	defer hsrv.Close()

	check := func() {
		m, e := Get(&Endpoint{Url: hsrv.URL, User: "apd", Password: "123"}, envelope.NS_WMI_CIMV2, "Win32_OperatingSystem", nil)

		if nil != e {
			t.Error(e)
			return
		}
		//     <p:BootDevice>\Device\HarddiskVolume2</p:BootDevice>
		// <p:BuildNumber>7601</p:BuildNumber>
		// <p:BuildType>Multiprocessor Free</p:BuildType>

		if "\\Device\\HarddiskVolume2" != m["BootDevice"] {
			t.Error("value of 'BootDevice' is not excepted, actual is", m["BootDevice"])
		}
		if "7601" != m["BuildNumber"] {
			t.Error("value of 'BuildNumber' is not excepted, actual is", m["BuildNumber"])
		}
		if "Multiprocessor Free" != m["BuildType"] {
			t.Error("value of 'BuildType' is not excepted, actual is", m["BuildType"])
		}

		t.Log(m)
	}

	WSMAN_DEBUG = false
	check()
	WSMAN_DEBUG = true
	check()
}

func TestEnumerateErrorXml(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CanHandle(r) {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, syntex_error_response)
		}
	}))
	defer hsrv.Close()

	it := Enumerate(&Endpoint{Url: hsrv.URL, User: "apd", Password: "123"}, envelope.NS_WMI_CIMV2, "Win32_Process", nil)
	count := 0
	for it.Next() {
		count++
		m, e := it.Value()
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

func TestEnumerateWithPull(t *testing.T) {
	request_count := 0
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CanHandle(r) {
			if 0 == request_count {
				w.WriteHeader(http.StatusOK)
				io.WriteString(w, list_first_response)
			} else if 1 == request_count {
				bs, _ := ioutil.ReadAll(r.Body)
				if !strings.Contains(string(bs), "<n:EnumerationContext>uuid:CD02E6BE-C6F3-47FD-9AF5-6DCBE89A448E</n:EnumerationContext>") {
					w.WriteHeader(http.StatusBadRequest)
					io.WriteString(w, "EnumerationContext is not excepted.")
					t.Log(string(bs))
				} else {
					w.WriteHeader(http.StatusOK)
					io.WriteString(w, list_end_response)
				}
			} else {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, "request is too much.")
			}
			request_count++
		}
	}))
	defer hsrv.Close()

	it := Enumerate(&Endpoint{Url: hsrv.URL, User: "apd", Password: "123"}, envelope.NS_WMI_CIMV2, "Win32_Process", nil)
	defer it.Close()

	count := 0
	for it.Next() {
		count++
		m, e := it.Value()
		if nil != e {
			t.Error(e)
			break
		}

		var excepted_caption string
		if 1 == count {
			excepted_caption = "System Idle Process"
		} else {
			excepted_caption = "SystemTest"
		}
		if excepted_caption != m["Caption"] {
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

	if 2 != count {
		t.Error("excepted count is 2, actual is ", count)
	}
	if nil != it.Err() {
		t.Error(it.Err())
	}
}
