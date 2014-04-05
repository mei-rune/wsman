package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/runner-mei/wsman"
	"net/textproto"
	"os"
)

const default_endpoint = "http://localhost:5985/wsman"

var user = flag.String("user", "meifakun", "user to run as")
var pass = flag.String("pass", "mfk", "user's password")

func main() {
	flag.Parse()
	args := flag.Args()
	var endpoint, cmd string
	if nil == args || 0 == len(args) {
		endpoint = default_endpoint
		cmd = "cmd"
	} else if 1 == len(args) {
		endpoint = args[0]
		cmd = "cmd"
	} else if 2 == len(args) {
		endpoint = args[0]
		cmd = args[1]
	} else {
		fmt.Println("参数太多")
		return
	}

	sh, e := wsman.NewShell(endpoint, *user, *pass)
	if nil != e {
		fmt.Println(e)
		return
	}
	defer sh.Close()

	cmd_id, e := sh.NewCommand(cmd)
	if nil != e {
		fmt.Println(e)
		return
	}

	go func() {
		for {
			res, e := sh.Read(cmd_id)
			if e != nil {
				fmt.Println(e)
				return
			}
			fmt.Print(wsman.ToString(res.Stdout))
			fmt.Print(wsman.ToString(res.Stderr))
			if res.IsDone() {
				break
			}
		}
		sh.Close()
		os.Exit(0)
	}()

	rd := textproto.NewReader(bufio.NewReader(os.Stdin))
	for {
		line, e := rd.ReadLine()
		if nil != e {
			fmt.Println("[local]", e)
			return
		}
		if "exit" == line {
			sh.Signal(cmd_id)
			break
		}

		e = sh.Send(cmd_id, line+"\r\n")
		if nil != e {
			fmt.Println("[local]", e)
			return
		}
	}
}
