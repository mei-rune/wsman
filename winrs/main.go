package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/runner-mei/wsman/winrm"
	"net/textproto"
	"os"
)

const default_endpoint = "http://localhost:5985/wsman"

var user = flag.String("user", "meifakun", "user to run as")
var pass = flag.String("pass", "mfk", "user's password")

func main() {
	flag.Parse()
	args := flag.Args()
	var endpoint string
	if nil == args || 0 == len(args) {
		endpoint = default_endpoint
	} else if 1 == len(args) {
		endpoint = args[0]
	} else {
		fmt.Println("参数太多")
		return
	}

	sh, e := winrm.NewShell(endpoint, *user, *pass)
	if nil != e {
		fmt.Println(e)
		return
	}

	sh.Stdout = os.Stdout
	sh.Stderr = os.Stderr
	fmt.Println(">")
	rd := textproto.NewReader(bufio.NewReader(os.Stdin))
	for {
		line, e := rd.ReadLine()
		if nil != e {
			fmt.Println("[local]", e)
			return
		}

		cmd, e := sh.NewCommand(line)
		if nil != e {
			fmt.Println(e)
			return
		}
		if _, e = cmd.Receive(); e != nil {
			fmt.Println(e)
			return
		}
		fmt.Println(">")
	}
}
