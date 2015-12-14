package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/runner-mei/wsman"
)

// var (
// 	endpoint = flag.String("host", "127.0.0.1:5985", "")
// 	user     = flag.String("user", "administrator", "")
// 	pass     = flag.String("password", "", "")
// )

func Sendfile(file string) {
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
	if _, e := sendfile(shell, file); nil != e {
		fmt.Println(e)
		Exit(shell, -1)
		return
	}
}

func mkdir(shell *wsman.Shell, dir string) error {
	// fmt.Println("mkdir", dir)
	return execCmd(shell, "cmd /c \"if not exist "+dir+" mkdir "+dir+"\"", os.Stdout, os.Stderr)
}

func mkalldir(shell *wsman.Shell, root, dir string) error {
	ss := strings.Split(filepath.ToSlash(dir), "/")
	for _, s := range ss {
		if "" == s {
			continue
		}

		root = filepath.Join(root, s)
		if e := mkdir(shell, root); nil != e {
			return errors.New("failed to create " + dir + ", " + e.Error())
		}
	}
	return nil
}

func sendfile(shell *wsman.Shell, file string) (string, error) {
	if e := execCmd(shell, "cmd /c \"if not exist tpt_scripts mkdir tpt_scripts\"", os.Stdout, os.Stderr); nil != e {
		return "", errors.New("failed to create tpt_scripts, " + e.Error())
	}

	f, e := os.Open(file)
	if nil != e {
		return "", e
	}

	filename := filepath.Join("tpt_scripts", file)
	if s := filepath.Dir(file); "" != s {
		if e := mkalldir(shell, "tpt_scripts", s); nil != e {
			return "", e
		}
	}
	reader := bufio.NewReader(f)

	fmt.Println("send '" + filename + "' in the process")
	defer fmt.Println()

	if_first := true
	for {
		bs, _, e := reader.ReadLine()
		if nil != e {
			if io.EOF == e {
				break
			}
			return "", e
		}

		var cmd_str string
		if len(bytes.TrimSpace(bs)) == 0 {
			if if_first {
				if_first = false
				cmd_str = "echo= >" + filename
			} else {
				cmd_str = "echo= >>" + filename
			}
		} else {
			if if_first {
				if_first = false
				cmd_str = "echo " + escape(string(bs), false) + " >" + filename
			} else {
				cmd_str = "echo " + escape(string(bs), false) + " >>" + filename
			}
		}

		if e := execCmd(shell, cmd_str, os.Stdout, os.Stderr); nil != e {
			return "", e
		}

		fmt.Print(".")
	}
	return filename, nil
}

func execCmd(shell *wsman.Shell, cmd string, out, err io.Writer) error {
	cmd_id, e := shell.NewCommand(cmd, nil)
	if nil != e {
		return e
	}

	for {
		res, e := shell.Read(cmd_id)
		if e != nil {
			if e := shell.Signal(cmd_id, wsman.SIGNAL_TERMINATE); nil != e {
				fmt.Println("[error]", e)
			}
			return e
		}
		for _, bs := range res.Stderr {
			err.Write(bs)
		}
		for _, bs := range res.Stdout {
			out.Write(bs)
		}
		if res.IsDone() {
			if e := shell.Signal(cmd_id, wsman.SIGNAL_TERMINATE); nil != e {
				fmt.Println("[error]", e)
			}
			if res.ExitCode != "" && res.ExitCode != "0" {
				return errors.New("exit with " + res.ExitCode)
			}
			return nil
		}
	}
}
