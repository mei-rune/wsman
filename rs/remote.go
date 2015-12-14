package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/runner-mei/wsman"
)

// var (
// 	endpoint = flag.String("host", "127.0.0.1:5985", "")
// 	user     = flag.String("user", "administrator", "")
// 	pass     = flag.String("password", "", "")
// )

func RemoteExec(args []string) {
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

	rname, e := sendfileIfNeed(shell, args[0])
	if nil != e {
		fmt.Println(e)
		Exit(shell, -1)
		return
	}

	if e := execCmd(shell, rname+" "+join(args[1:]), os.Stdout, os.Stderr); nil != e {
		fmt.Println(e)
		Exit(shell, -1)
		return
	}
}

func sendfileIfNeed(shell *wsman.Shell, file string) (string, error) {
	u, e := url.Parse(file)
	if nil != e || "" == u.Scheme {
		return sendfile(shell, file)
	}

	if st, e := os.Stat(*wget); nil != e {
		return "", e
	} else if st.IsDir() {
		return "", errors.New("'" + *wget + "' is directory.")
	}

	rget, e := sendfile(shell, *wget)
	if nil != e {
		return "", errors.New("'" + *wget + "' send failed, " + e.Error())
	}

	rname := filepath.Join("tpt_scripts", u.Path)
	if s := filepath.Dir(u.Path); "" != s {
		if e := mkalldir(shell, "tpt_scripts", s); nil != e {
			return "", e
		}
	}

	if e := execCmd(shell, "cscript //nologo //e:jscript "+rget+" "+file+" "+rname, nil, nil); nil != e {
		return "", e
	}

	return rname, nil
}
