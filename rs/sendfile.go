package rs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/runner-mei/wsman"
)

func mkdir(shell *wsman.Shell, dir string) error {
	// fmt.Println("mkdir", dir)
	return ExecCmd(shell, "cmd /c \"if not exist "+dir+" mkdir "+dir+"\"", os.Stdout, os.Stderr)
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

func Sendfile(shell *wsman.Shell, file string) (string, error) {
	if e := ExecCmd(shell, "cmd /c \"if not exist tpt_scripts mkdir tpt_scripts\"", os.Stdout, os.Stderr); nil != e {
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

	fmt.Println("send batch file in the process")
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
				cmd_str = "echo " + Escape(string(bs), false) + " >" + filename
			} else {
				cmd_str = "echo " + Escape(string(bs), false) + " >>" + filename
			}
		}

		if e := ExecCmd(shell, cmd_str, os.Stdout, os.Stderr); nil != e {
			return "", e
		}

		fmt.Print(".")
	}
	return filename, nil
}

func ExecCmd(shell *wsman.Shell, cmd string, out, err io.Writer) error {
	cmd_id, e := shell.NewCommand(cmd, nil)
	if nil != e {
		return e
	}
	defer func() {
		if e := shell.Signal(cmd_id, wsman.SIGNAL_TERMINATE); nil != e {
			if !strings.Contains(e.Error(), "The parameter is incorrect.") {
				fmt.Println("[error] terminate", e)
			}
		}
	}()

	for {
		res, e := shell.Read(cmd_id)
		if e != nil {
			return e
		}
		if nil != err {
			for _, bs := range res.Stderr {
				err.Write(bs)
			}
		}
		if nil != out {
			for _, bs := range res.Stdout {
				out.Write(bs)
			}
		}
		if res.IsDone() {
			if res.ExitCode != "" && res.ExitCode != "0" {
				return errors.New("exit with " + res.ExitCode)
			}
			return nil
		}
	}
}
