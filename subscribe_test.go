package wsman

import (
	"runtime"
	"testing"

	"github.com/runner-mei/wsman/envelope"
)

func TestSimpleSubscribe(t *testing.T) {
	if "windows" != runtime.GOOS {
		t.Skip("linux is not supported.")
	}

	WSMAN_DEBUG = true
	//http: //schemas.microsoft.com/wbem/wsman/1/windows/EventLog

	// srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	// 	fmt.Println("=============================")
	// 	fmt.Println(req.Method, req.URL.String())
	// 	for k, a := range req.Header {
	// 		fmt.Println(k, a)
	// 	}
	// 	//req.
	// 	//fmt.Println(req.)

	// 	fmt.Println()
	// 	if nil != req.Body {
	// 		if n, e := io.Copy(os.Stderr, req.Body); nil != e {
	// 			t.Error(e)
	// 			return
	// 		} else {
	// 			fmt.Println("body is empty - ", n)
	// 		}

	// 		req.Body.Close()
	// 	}

	// 	w.WriteHeader(http.StatusAccepted)
	// 	io.WriteString(w, "OK")
	// }))
	// defer srv.Close()

	m, e := SubscribeByPull(&Endpoint{Url: *win_url, User: *win_user, Password: *win_password},
		envelope.NS_WINDOWS, "EventLog", nil, map[string][]envelope.QueryFilter{"0": {{Path: "Application", Value: "*"}}},
		true)
	if nil != e {
		t.Error(e)
		return
	}

	defer m.Close()

	for m.Next() {
		v, e := m.Value()
		if nil != e {
			t.Error(e)
			break
		}
		t.Log(v)
	}
	if nil != m.Err() {
		t.Error(m.Err())
	}
}
