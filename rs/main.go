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
	"strings"
	"unicode"

	"github.com/runner-mei/wsman"
)

var (
	endpoint = flag.String("host", "127.0.0.1:5985", "")
	user     = flag.String("user", "administrator", "")
	pass     = flag.String("password", "", "")
)

func escapeTo(s string, buf *bytes.Buffer) {
	for _, c := range s {
		if '"' == c {
			buf.WriteString("\\\"")
		} else if '"' == c {
			buf.WriteString("\\\\")
		} else {
			buf.WriteRune(c)
		}
	}
}

func main() {
	flag.Parse()
	if 0 == len(flag.Args()) {
		fmt.Println("command is missing.")
		os.Exit(-1)
		return
	}

	url_str := *endpoint
	if u, e := url.Parse(url_str); nil != e || "" == u.Scheme {
		if _, _, e = net.SplitHostPort(url_str); nil == e {
			url_str = "http://" + url_str + "/wsman"
		} else {
			url_str = "http://" + url_str + ":5985/wsman"
		}
	}

	shell, e := wsman.NewShell(url_str, *user, *pass)
	if nil != e {
		fmt.Println(e)
		os.Exit(-1)
		return
	}
	defer shell.Close()

	var cmd_str string

	if 1 == len(flag.Args()) {
		cmd_str = flag.Args()[0]
	} else {
		var buf bytes.Buffer
		for idx, word := range flag.Args() {
			if 0 != idx {
				buf.WriteString(" ")
			}

			if strings.Contains(word, "\"") {
				buf.WriteString("\"")
				escapeTo(word, &buf)
				buf.WriteString("\"")
			} else if p := strings.IndexFunc(word, unicode.IsSpace); p >= 0 {
				buf.WriteString("\"")
				buf.WriteString(word)
				buf.WriteString("\"")
			} else {
				buf.WriteString(word)
			}
		}

		cmd_str = buf.String()
	}

	cmd_id, e := shell.NewCommand(cmd_str)
	if nil != e {
		fmt.Println(e)
		os.Exit(-1)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		<-c
		if nil == shell {
			return
		}
		if e := shell.Signal(cmd_id, wsman.SIGNAL_TERMINATE); nil != e {
			fmt.Println(e)
			os.Exit(-1)
			return
		}
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
			os.Exit(-1)

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
			break
		}
	}
}
