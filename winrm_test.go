package wsman

import (
	"flag"
	"runtime"
	"testing"
)

var (
	win_url      = flag.String("win_url", "http://127.0.0.1:5985/wsman", "")
	win_user     = flag.String("win_user", "meifakun", "")
	win_password = flag.String("win_password", "mfk", "")
)

func TestSimpleEnumerateOS(t *testing.T) {
	if "windows" != runtime.GOOS {
		t.Skip("linux is not supported.")
	}
	it := Enumerate(&Endpoint{Url: *win_url, User: *win_user, Password: *win_password},
		"Win32_OperatingSystem", nil)
	defer it.Close()
	count := 0
	for it.Next() {
		count++
		m, e := it.Map()
		if nil != e {
			t.Error(e)
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
}

func TestSimpleEnumerateProcess(t *testing.T) {
	if "windows" != runtime.GOOS {
		t.Skip("linux is not supported.")
	}
	it := Enumerate(&Endpoint{Url: *win_url, User: *win_user, Password: *win_password},
		"Win32_Process", nil)
	defer it.Close()
	count := 0
	for it.Next() {
		count++
		m, e := it.Map()
		if nil != e {
			t.Error(e)
			break
		}

		t.Log(m)
	}

	if 20 > count {
		t.Error("excepted count is 1, actual is ", count)
	}
	if nil != it.Err() {
		t.Error(it.Err())
	}
}

// func TestSimpleEnumerateService(t *testing.T) {
// 	WSMAN_DEBUG = true
// 	it := Enumerate(&Endpoint{Url: *win_url, User: *win_user, Password: *win_password},
// 		"Win32_Service", map[string]string{"Name": "spooler"})
// 	defer it.Close()
// 	count := 0
// 	for it.Next() {
// 		count++
// 		m, e := it.Map()
// 		if nil != e {
// 			t.Error(e)
// 			break
// 		}

// 		t.Log(m)
// 	}

// 	if 20 > count {
// 		t.Error("excepted count is 1, actual is ", count)
// 	}
// 	if nil != it.Err() {
// 		t.Error(it.Err())
// 	}
// }

func TestSimpleEnumerateWin32_NetworkAdapterConfiguration(t *testing.T) {
	WSMAN_DEBUG = true
	it := Enumerate(&Endpoint{Url: *win_url, User: *win_user, Password: *win_password},
		"Win32_NetworkAdapterConfiguration", nil)
	defer it.Close()
	count := 0
	for it.Next() {
		count++
		m, e := it.Map()
		if nil != e {
			t.Error(e)
			break
		}

		t.Log(m)
	}

	if 20 > count {
		t.Error("excepted count is 1, actual is ", count)
	}
	if nil != it.Err() {
		t.Error(it.Err())
	}
}

func TestSimpleGetOS(t *testing.T) {
	if "windows" != runtime.GOOS {
		t.Skip("linux is not supported.")
	}
	m, e := Get(&Endpoint{Url: *win_url, User: *win_user, Password: *win_password},
		"Win32_OperatingSystem", nil)
	if nil != e {
		t.Error(e)
		return
	}

	t.Log(m)
}

func TestSimpleGetServiceWithName(t *testing.T) {
	if "windows" != runtime.GOOS {
		t.Skip("linux is not supported.")
	}
	m, e := Get(&Endpoint{Url: *win_url, User: *win_user, Password: *win_password},
		"Win32_Service", map[string]string{"Name": "spooler"})
	if nil != e {
		t.Error(e)
		return
	}

	t.Log(m)
}
