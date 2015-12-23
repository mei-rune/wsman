package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"net/textproto"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"unicode"

	"github.com/runner-mei/wsman"
)

var (
	endpoint = flag.String("host", "127.0.0.1:5985", "远程 windows 的地址")
	user     = flag.String("user", "administrator", "远程 windows 的用户名")
	pass     = flag.String("password", "", "远程 windows 的用户密码")
	wget     = flag.String("wget", "wget.js", "下载工具 wget.js 的路径")
)

func escape(s string, isqoute bool) string {
	var buf bytes.Buffer
	escapeTo(s, &buf, isqoute)
	return buf.String()
}

func escapeTo(s string, buf *bytes.Buffer, isqoute bool) {
	for _, c := range s {
		switch c {
		case '"':
			if isqoute {
				buf.WriteString("\\\"")
			} else {
				buf.WriteString("\"")
			}
		case '\\':
			if isqoute {
				buf.WriteString("\\\\")
			} else {
				buf.WriteString("\\")
			}
		case '%':
			if isqoute {
				buf.WriteString("%%")
			} else {
				buf.WriteString("%")
			}
		case '^':
			if isqoute {
				buf.WriteString("^")
			} else {
				buf.WriteString("^^")
			}
		case '>':
			if isqoute {
				buf.WriteString(">")
			} else {
				buf.WriteString("^>")
			}
		case '<':
			if isqoute {
				buf.WriteString("<")
			} else {
				buf.WriteString("^<")
			}
		case '|':
			if isqoute {
				buf.WriteString("|")
			} else {
				buf.WriteString("^|")
			}
		case '&':
			if isqoute {
				buf.WriteString("&")
			} else {
				buf.WriteString("^&")
			}
		// case '\'':
		// 	buf.WriteString("^'")
		// case '`':
		// 	buf.WriteString("^`")
		// case ';':
		// 	buf.WriteString("^;")
		// case '=':
		// 	buf.WriteString("^=")
		// case '(':
		// 	buf.WriteString("^(")
		// case ')':
		// 	buf.WriteString("^)")
		// case '!':
		// 	buf.WriteString("^^!")
		// case ',':
		// 	buf.WriteString("^,")
		// case '[':
		// 	buf.WriteString("^[")
		// case ']':
		// 	buf.WriteString("^]")
		default:
			buf.WriteRune(c)
		}
	}
}

func NewShell() (*wsman.Shell, error) {
	url_str := *endpoint
	if u, e := url.Parse(url_str); nil != e || "" == u.Scheme {
		if _, _, e = net.SplitHostPort(url_str); nil == e {
			url_str = "http://" + url_str + "/wsman"
		} else {
			url_str = "http://" + url_str + ":5985/wsman"
		}
	}

	return wsman.NewShell(url_str, *user, *pass)
}

func join(args []string) string {
	if 1 == len(args) {
		return args[0]
	} else {
		var buf bytes.Buffer
		for idx, word := range args {
			if 0 != idx {
				buf.WriteString(" ")
			}

			if strings.Contains(word, "\"") {
				buf.WriteString("\"")
				escapeTo(word, &buf, true)
				buf.WriteString("\"")
			} else if p := strings.IndexFunc(word, unicode.IsSpace); p >= 0 {
				buf.WriteString("\"")
				buf.WriteString(word)
				buf.WriteString("\"")
			} else {
				buf.WriteString(word)
			}
		}

		return buf.String()
	}
}

func fileExists(nm string) bool {
	st, e := os.Stat(nm)
	if nil != e {
		return false
	}
	return !st.IsDir()
}

func main() {
	flag.Parse()
	if 0 == len(flag.Args()) {
		fmt.Println("command is missing.")
		Exit(nil, -1)
		return
	}

	args := flag.Args()
	if 2 == len(args) {
		if "[sendfile]" == args[0] {
			Sendfile(args[1])
			return
		}
	}
	if len(args) >= 2 {
		if "[exec]" == args[0] {
			if "wget.js" == *wget {
				for _, nm := range []string{*wget, "tools/wget.js", "../tools/wget.js", "../wget.js"} {
					if fileExists(nm) {
						flag.Set("wget", nm)
						break
					}
				}
			}
			RemoteExec(args[1:])
			return
		}
	}

	if 1 == len(args) && "[enum]" == args[0] {
		url_str := *endpoint
		if u, e := url.Parse(url_str); nil != e || "" == u.Scheme {
			if _, _, e = net.SplitHostPort(url_str); nil == e {
				url_str = "http://" + url_str + "/wsman"
			} else {
				url_str = "http://" + url_str + ":5985/wsman"
			}
		}

		enum := wsman.Enumerate(&wsman.Endpoint{url_str, *user, *pass},
			"http://schemas.microsoft.com/wbem/wsman/1/windows/shell", "cmd", nil)
		for enum.Next() {
			fmt.Println(enum.Value)
		}
		if nil != enum.Err() {
			fmt.Println(enum.Err())
		}
		return
	}

	shell, e := NewShell()
	if nil != e {
		fmt.Println(e)
		Exit(shell, -1)
		return
	}
	defer shell.Close()

	cmd_id, e := shell.NewCommand(args[0], args[1:])
	if nil != e {
		fmt.Println(e)
		Exit(shell, -1)
		return
	}

	is_terminate := false
	defer func() {
		if is_terminate {
			return
		}
		if e := shell.Signal(cmd_id, wsman.SIGNAL_TERMINATE); nil != e {
			fmt.Println("[error]", e)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		<-c
		if nil == shell {
			return
		}
		if e := shell.Signal(cmd_id, wsman.SIGNAL_TERMINATE); nil != e {
			fmt.Println(e)
			Exit(shell, -1)
			return
		}
		is_terminate = true
	}()

	go func() {
		rd := textproto.NewReader(bufio.NewReader(os.Stdin))
		for {
			line, e := rd.ReadLine()
			if nil != e {
				fmt.Println("[local]", e)
				return
			}
			e = shell.Send(cmd_id, line+"\r\n")
			if nil != e {
				fmt.Println("[local]", e)
				return
			}
		}
	}()

	for {
		res, e := shell.Read(cmd_id)
		if e != nil {
			fmt.Println(e)
			Exit(shell, -1)
			shell = nil
			c <- os.Kill
			return
		}
		for _, bs := range res.Stderr {
			os.Stderr.Write(bs)
		}
		for _, bs := range res.Stdout {
			os.Stdout.Write(bs)
		}
		if res.IsDone() {
			shell = nil
			c <- os.Kill

			if res.ExitCode != "" && res.ExitCode != "0" {
				fmt.Println("exit code is", res.ExitCode)
				code, e := strconv.ParseInt(res.ExitCode, 10, 0)
				if nil != e {
					code = -1
				}
				Exit(shell, int(code))
				return
			}

			break
		}
	}
}

func Exit(shell *wsman.Shell, exitCode int) {
	if nil != shell {
		shell.Close()
	}
	os.Exit(exitCode)
}
