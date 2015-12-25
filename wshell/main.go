package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"net/textproto"
	"net/url"
	"os"
	"os/signal"
	"strconv"

	"github.com/runner-mei/wsman"
	"github.com/runner-mei/wsman/rs"
)

var (
	endpoint = flag.String("host", "127.0.0.1:5985", "远程 windows 的地址")
	user     = flag.String("user", "administrator", "远程 windows 的用户名")
	pass     = flag.String("password", "", "远程 windows 的用户密码")
)

func NewShell() (*wsman.Shell, error) {
	url_str := *endpoint
	if u, e := url.Parse(url_str); nil != e || "" == u.Scheme {
		if _, _, e = net.SplitHostPort(url_str); nil == e {
			url_str = "http://" + url_str + "/wsman"
		} else {
			url_str = "http://" + url_str + ":5985/wsman"
		}
	}

	return wsman.NewShell(url_str, *user, *pass, "")
}

func main() {
	flag.StringVar(&rs.WgetFile, "wget", "wget.js", "下载工具 wget.js 的路径")

	flag.Parse()
	if 0 == len(flag.Args()) {
		fmt.Println("command is missing.")
		Exit(nil, -1)
		return
	}

	args := flag.Args()
	if 2 == len(args) {
		if "[sendfile]" == args[0] {
			sendfile(args[1])
			return
		}
	}
	if len(args) >= 2 {
		if "[exec]" == args[0] {
			if "wget.js" == rs.WgetFile {
				for _, nm := range []string{rs.WgetFile, "tools/wget.js", "../tools/wget.js", "../wget.js"} {
					if rs.FileExists(nm) {
						flag.Set("wget", nm)
						break
					}
				}
			}
			remoteExec(args[1:])
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
	defer func() {
		if nil == shell {
			return
		}
		shell.Close()
	}()

	cmd_id, e := shell.NewCommand(args[0], args[1:])
	if nil != e {
		fmt.Println(e)
		Exit(shell, -1)
		return
	}

	is_terminate := false
	var terminate = func() {
		if is_terminate {
			return
		}

		if nil == shell {
			return
		}

		if e := shell.Signal(cmd_id, wsman.SIGNAL_TERMINATE); nil != e {
			fmt.Println("[error]", e)
		}
		shell = nil
		is_terminate = true
	}

	defer terminate()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		<-c
		terminate()
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
			return
		}
		for _, bs := range res.Stderr {
			os.Stderr.Write(bs)
		}
		for _, bs := range res.Stdout {
			os.Stdout.Write(bs)
		}
		if res.IsDone() {
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

func remoteExec(args []string) {
	shell, e := NewShell()
	if nil != e {
		fmt.Println(e)
		Exit(shell, -1)
		return
	}
	defer shell.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		Exit(shell, -1)
	}()

	e = rs.RemoteExec(shell, args[0], args[1:])
	if nil != e {
		fmt.Println(e)
		Exit(shell, -1)
		return
	}
}

func sendfile(file string) {
	// 	flag.Parse()
	// 	if 0 == len(flag.Args()) {
	// 		fmt.Println("file is missing.")
	// 		Exit(shell, -1)	// 		return
	// 	}

	// 	if 1 != len(flag.Args()) {
	// 		fmt.Println("arguments is to much.")
	// 		Exit(shell, -1)	// 		return
	// 	}
	// 	file := flag.Args()[0]

	shell, e := NewShell()
	if nil != e {
		fmt.Println(e)
		Exit(shell, -1)
		return
	}
	defer shell.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		// if nil == shell {
		// 	return
		// }
		// if e := shell.Signal(cmd_id, wsman.SIGNAL_TERMINATE); nil != e {
		// 	fmt.Println(e)
		// 	Exit(shell, -1)		// 	return
		// }

		Exit(shell, -1)
	}()
	if _, e := rs.Sendfile(shell, file); nil != e {
		fmt.Println(e)
		Exit(shell, -1)
		return
	}
}
