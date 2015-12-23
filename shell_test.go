package wsman

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_creating_a_shell(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Basic dmFncmFudDp2YWdyYW50" {
			t.Fatal("bad authorization")
		}
		fmt.Fprintf(w, `
			<Envelope>
				<s:Body>
			    <x:ResourceCreated>
			      <a:Address>http://127.0.0.1:7090/wsman</a:Address>
			      <a:ReferenceParameters>
			        <w:ResourceURI>http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd</w:ResourceURI>
			        <w:SelectorSet>
			          <w:Selector Name="ShellId">ABCXYZ</w:Selector>
			        </w:SelectorSet>
			      </a:ReferenceParameters>
			    </x:ResourceCreated>
				</s:Body>
			</Envelope>`)
	}))
	defer hsrv.Close()

	s, err := NewShell(hsrv.URL, "vagrant", "vagrant", "")

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// if s.Endpoint != fixture.Endpoint {
	// 	t.Fatal("bad endpoint:", s.Endpoint)
	// }

	if s.Id != "ABCXYZ" {
		t.Fatal("bad shell id:", s.Id)
	}

	if s.User != "vagrant" {
		t.Fatal("bad owner:", s.User)
	}
}
func Test_creating_a_shell2(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Basic dmFncmFudDp2YWdyYW50" {
			t.Fatal("bad authorization")
		}
		fmt.Fprintf(w, `
			<Envelope>
				<s:Body>
					<rsp:Shell>
						<rsp:ShellId>ABCXYZ</rsp:ShellId>
					</rsp:Shell>
				</s:Body>
			</Envelope>`)
	}))
	defer hsrv.Close()

	s, err := NewShell(hsrv.URL, "vagrant", "vagrant", "")

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// if s.Endpoint != fixture.Endpoint {
	// 	t.Fatal("bad endpoint:", s.Endpoint)
	// }

	if s.Id != "ABCXYZ" {
		t.Fatal("bad shell id:", s.Id)
	}

	if s.User != "vagrant" {
		t.Fatal("bad owner:", s.User)
	}
}

func Test_creating_a_shell_command(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		buffer, _ := ioutil.ReadAll(r.Body)
		body := string(buffer)

		if !strings.Contains(body, "<w:Selector Name=\"ShellId\">ABCXYZ</w:Selector>") {
			t.Log(body)
			t.Fatal("bad request: selector")
		}

		if !strings.Contains(body, "<rsp:Command>foo bar</rsp:Command>") {
			t.Log(body)
			t.Fatal("bad request: command")
		}

		fmt.Fprintf(w, `
			<Envelope>
				<s:Body>
					<rsp:CommandResponse>
						<rsp:CommandId>123456</rsp:CommandId>
					</rsp:CommandResponse>
				</s:Body>
			</Envelope>`)
	}))
	defer hsrv.Close()

	s := &Shell{
		Id:       "ABCXYZ",
		Endpoint: &Endpoint{Url: hsrv.URL},
	}

	cmd_id, err := s.NewCommand("foo bar", nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if cmd_id != "123456" {
		t.Fatal("bad command id:", cmd_id)
	}

	// if c.CommandText != "foo bar" {
	// 	t.Fatal("bad command text:", c.CommandText)
	// }
}

func Test_deleting_a_shell(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		buffer, _ := ioutil.ReadAll(r.Body)
		body := string(buffer)

		if !strings.Contains(body, "<w:Selector Name=\"ShellId\">ABCXYZ</w:Selector>") {
			t.Log(body)
			t.Fatal("bad request: selector")
		}
		// if r.XmlString("//Header/SelectorSet[Selector='ABCXYZ']") == "" {
		// 	t.Fatal("bad request: selector")
		// }
		fmt.Fprintf(w, `
			<Envelope>
				<s:Body></s:Body>
			</Envelope>`)
	}))
	defer hsrv.Close()

	s := &Shell{
		Id:       "ABCXYZ",
		Endpoint: &Endpoint{Url: hsrv.URL},
	}

	err := s.Close()

	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func Test_authentication_failure(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer hsrv.Close()

	_, err := NewShell(hsrv.URL, "", "", "")

	if err == nil {
		t.Fatal("bad: no error")
	}

	herr, ok := err.(*HttpError)
	if !ok {
		t.Fatal("bad: not an http error")
	}

	if herr.StatusCode != 401 {
		t.Fatal("bad: http status code", herr.StatusCode)
	}
}

func Test_receive_a_shell(t *testing.T) {
	hsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		buffer, _ := ioutil.ReadAll(r.Body)
		body := string(buffer)

		if !strings.Contains(body, "<w:Selector Name=\"ShellId\">ABCXYZ</w:Selector>") {
			t.Fatal("bad request: selector")
		}

		fmt.Fprintf(w, `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:rsp="http://schemas.microsoft.com/wbem/wsman/1/windows/shell" xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd" xml:lang="zh-CN">
  <s:Header>
    <a:Action>http://schemas.microsoft.com/wbem/wsman/1/windows/shell/ReceiveResponse</a:Action>
    <a:MessageID>uuid:60559B42-2690-437F-87E7-F3ABD5349DB0</a:MessageID>
    <a:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:To>
    <a:RelatesTo>uuid:630F9545-5D3A-4265-84BB-64149CBC7CF2</a:RelatesTo>
  </s:Header>
  <s:Body>
    <rsp:ReceiveResponse>
      <rsp:Stream Name="stdout" CommandId="8A9349A1-B973-4CCD-8E3E-E9562896FDCB">IMf9tq/G9yBDINbQtcS+7crHIE9TDQo=</rsp:Stream>
      <rsp:Stream Name="stdout" CommandId="8A9349A1-B973-4CCD-8E3E-E9562896FDCB">IL7ttcTQ8sHQusXKxyAzNjUxLTE2MUYNCg0KIEM6XFVzZXJzXG1laWZha3VuILXExL/CvA0KDQo=</rsp:Stream>
      <rsp:Stream Name="stdout" CommandId="8A9349A1-B973-4CCD-8E3E-E9562896FDCB">MjAxNC8wNC8wMyAgMTY6NTQgICAgPERJUj4gICAgICAgICAgLg0KMjAxNC8wNC8wMyAgMTY6NTQgICAgPERJUj4gICAgICAgICAgLi4NCjIwMTMvMDgvMjMgIDEzOjM0ICAgIDxESVI+ICAgICAgICAgIC5hbmRyb2lkDQoyMDEzLzAzLzI1ICAyMDoyOCAgICAgICAgICAgICAyLDcxNiAuYmFzaF9oaXN0b3J5DQoyMDEzLzAxLzIyICAyMDoxMSAgICA8RElSPiAgICAgICAgICAuZ2VtDQoyMDEzLzA0LzI2ICAyMDo1MyAgICAgICAgICAgICAgIDE3MCAuZ2VtcmMNCjIwMTQvMDMvMDggIDExOjQwICAgICAgICAgICAgICAgMTA3IC5naXRjb25maWcNCjIwMTMvMDgvMDIgIDEwOjE0ICAgICAgICAgICAgIDIsMTgyIC5rZGlmZjNyYw0KMjAxMy8wMi8xNCAgMTc6MjkgICAgICAgICAgICAgICAgIDAgLm1vbmdvcmMuanMNCjIwMTMvMDcvMzEgIDEyOjIzICAgICAgICAgICAgICAgIDIwIC5uZ3Jvaw0KMjAxMy8xMi8xOSAgMTA6MjEgICAgPERJUj4gICAgICAgICAgLm5vZGUtZ3lwDQoyMDEzLzA5LzE4ICAxNTowMiAgICA8RElSPiAgICAgICAgICAucXJzYm94DQoyMDEzLzA5LzE3ICAxNjo1OSAgICAgICAgICAgICAxLDgyMyAucmVkaXNjbGlfaGlzdG9yeQ0KMjAxMy8xMi8yNCAgMTA6NDggICAgICAgICAgICAgMSwwMjQgLnJuZA0KMjAxMy8wOC8wMyAgMTE6MjIgICAgPERJUj4gICAgICAgICAgLnNzaA0KMjAxNC8wMy8zMCAgMTU6MjcgICAgPERJUj4gICAgICAgICAgLlZpcnR1YWxCb3gNCjIwMTMvMTIvMTkgIDE3OjA2ICAgICAgICAgICAgICAgICAwIGRhZW1vbnByb2Nlc3MudHh0DQoyMDEzLzEyLzE5ICAxNjowOSAgICA8RElSPiAgICAgICAgICBEZXNrdG9wDQoyMDEzLzAxLzA5ICAyMzowMCAgICA8RElSPiAgICAgICAgICBEb3dubG9hZHMNCjIwMTMvMDUvMzEgIDIwOjQ4ICAgICAgICAgNyw2MTIsODMyIGVkYl9ucGdzcWwuZXhlDQoyMDEzLzA1LzMxICAyMTowMSAgICAgICAgMTIsOTMxLDIxNiBlZGJfcGdhZ2VudC5leGUNCjIwMTMvMDUvMzEgIDIxOjIzICAgICAgICAxNCw1NDUsNjgwIGVkYl9wZ2JvdW5jZXIuZXhlDQoyMDEzLzA1LzMxICAyMToyNCAgICAgICAgIDcsMTgyLDc1MiBlZGJfcGdqZGJjLmV4ZQ0KMjAxMy8wMS8wOSAgMjI6NTIgICAgPERJUj4gICAgICAgICAgRmF2b3JpdGVzDQoyMDEzLzEyLzE5ICAxNjozNyAgICA8RElSPiAgICAgICAgICBHTlMzDQoyMDEyLzAxLzE0ICAwMjo1NyAgICAgICAgICAgNzA3LDM5MiBpd2x3aWZpLTIwMzAtNi51Y29kZQ0KMjAxMi8wMS8xNCAgMDI6NTcgICAgICAgICAgICAgMiwwNDYgTElDRU5TRS5pd2x3aWZpLTIwMzAtdWNvZGUNCjIwMTMvMDcvMzEgIDEwOjU3ICAgICAgICAgICAgICAgIDQzIG1lcmN1cmlhbC5pbmkNCjIwMTMvMDkvMDUgIDExOjExICAgIDxESVI+ICAgICAgICAgIE9yYWNsZQ0KMjAxMi8wMS8xNCAgMDI6NTcgICAgICAgICAgICAgNCw3NzggUkVBRE1FLml3bHdpZmktMjAzMC11Y29kZQ0KMjAxNC8wMi8yMiAgMTE6MzUgICAgPERJUj4gICAgICAgICAgVmlydHVhbEJveCBWTXMNCjIwMTQvMDQvMDMgIDE2OjU0ICAgICAgICAgICAgICAgICAwIHdpbnJtDQoyMDEzLzEyLzIxICAxNDowMyAgICAgICAgICAgICA0LDAzMSBfdmltaW5mbw0KMjAxMy8xMS8yNiAgMTQ6NTAgICAgPERJUj4gICAgICAgICAgztK1xLG4t93OxLz+DQogICAgICAgICAgICAgIDE5ILj2zsS8/iAgICAgNDIsOTk4LDgxMiDX1r3aDQogICAgICAgICAgICAgIDE1ILj2xL/CvCA5MSw3NjgsMzc3LDM0NCC/ydPD19a92g0K</rsp:Stream>
      <rsp:Stream Name="stdout" CommandId="8A9349A1-B973-4CCD-8E3E-E9562896FDCB" End="true"/>
      <rsp:Stream Name="stderr" CommandId="8A9349A1-B973-4CCD-8E3E-E9562896FDCB" End="true"/>
      <rsp:CommandState CommandId="8A9349A1-B973-4CCD-8E3E-E9562896FDCB" State="http://schemas.microsoft.com/wbem/wsman/1/windows/shell/CommandState/Done">
        <rsp:ExitCode>0</rsp:ExitCode>
      </rsp:CommandState>
    </rsp:ReceiveResponse>
  </s:Body>
</s:Envelope>`)
	}))
	defer hsrv.Close()

	s := &Shell{
		Id:       "ABCXYZ",
		Endpoint: &Endpoint{Url: hsrv.URL},
	}

	res, err := s.Read("8A9349A1-B973-4CCD-8E3E-E9562896FDCB")
	if nil != err {
		t.Error("bad: no error")
		return
	}

	if "0" != res.ExitCode {
		t.Error("res.ExitCode is not excepted - ", res.ExitCode)
	}

	if !res.IsDone() {
		t.Error("res.State is not excepted - ", res.State)
	}

	stdout := ToString(res.Stdout)
	stderr := ToString(res.Stderr)

	if !strings.Contains(stdout, "2012/01/14  02:57           707,392 iwlwifi-2030-6.ucode") {
		t.Error("stdout is not exepted, ", stdout)
	}
	if "" != stderr {
		t.Error("stdout is not exepted, ", stdout)
	}
}
