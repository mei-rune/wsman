package winrm

import (
	"bytes"
	"fmt"
	"github.com/runner-mei/wsman"
	"io"
	"io/ioutil"
	"launchpad.net/xmlpath"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type Fixture struct {
	mux      *http.ServeMux
	server   *httptest.Server
	Endpoint string
}

func NewFixture() *Fixture {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	return &Fixture{mux, server, server.URL}
}

func (f *Fixture) HandleFunc(handler func(w http.ResponseWriter, r *Request)) {
	f.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if wsman.CanHandle(r) {
			handler(w, newRequest(r))
		}

		go f.server.Close()
	})
}

type Request struct {
	*http.Request
	reader io.ReadSeeker
}

func newRequest(r *http.Request) *Request {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	return &Request{
		Request: r, reader: bytes.NewReader(body),
	}
}

func (r *Request) XmlString(query string) string {
	xpath, err := xmlpath.Compile(query)

	if err != nil {
		return ""
	}

	buffer, _ := ioutil.ReadAll(r.reader)
	r.reader.Seek(0, 0)

	node, err := xmlpath.Parse(bytes.NewReader(buffer))

	if err != nil {
		return ""
	}

	result, _ := xpath.String(node)
	return strings.Trim(result, " \r\n\t")
}

func Test_creating_a_shell(t *testing.T) {
	fixture := NewFixture()

	fixture.HandleFunc(func(w http.ResponseWriter, r *Request) {
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
	})

	s, err := NewShell(fixture.Endpoint, "vagrant", "vagrant")

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if s.Endpoint != fixture.Endpoint {
		t.Fatal("bad endpoint:", s.Endpoint)
	}

	if s.Id != "ABCXYZ" {
		t.Fatal("bad shell id:", s.Id)
	}

	if s.Owner != "vagrant" {
		t.Fatal("bad owner:", s.Owner)
	}
}

func Test_creating_a_shell_command(t *testing.T) {
	fixture := NewFixture()

	fixture.HandleFunc(func(w http.ResponseWriter, r *Request) {
		if r.XmlString("//Header/SelectorSet[Selector='ABCXYZ']") == "" {
			t.Fatal("bad request: selector")
		}
		if r.XmlString("//Body/CommandLine[Command='foo bar']") == "" {
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
	})

	s := &Shell{
		Id:       "ABCXYZ",
		Endpoint: fixture.Endpoint,
	}

	c, err := s.NewCommand("foo bar")

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if c.Id != "123456" {
		t.Fatal("bad command id:", c.Id)
	}

	if c.CommandText != "foo bar" {
		t.Fatal("bad command text:", c.CommandText)
	}
}

func Test_deleting_a_shell(t *testing.T) {
	fixture := NewFixture()

	fixture.HandleFunc(func(w http.ResponseWriter, r *Request) {
		if r.XmlString("//Header/SelectorSet[Selector='ABCXYZ']") == "" {
			t.Fatal("bad request: selector")
		}
		fmt.Fprintf(w, `
			<Envelope>
				<s:Body></s:Body>
			</Envelope>`)
	})

	s := &Shell{
		Id:       "ABCXYZ",
		Endpoint: fixture.Endpoint,
	}

	err := s.Delete()

	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func Test_authentication_failure(t *testing.T) {
	fixture := NewFixture()

	fixture.HandleFunc(func(w http.ResponseWriter, r *Request) {
		w.WriteHeader(401)
	})

	_, err := NewShell(fixture.Endpoint, "", "")

	if err == nil {
		t.Fatal("bad: no error")
	}

	herr, ok := err.(*wsman.HttpError)
	if !ok {
		t.Fatal("bad: not an http error")
	}

	if herr.StatusCode != 401 {
		t.Fatal("bad: http status code", herr.StatusCode)
	}
}
