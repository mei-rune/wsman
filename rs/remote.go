package rs

import (
	"bytes"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/runner-mei/wsman"
)

var WgetFile string = "wget.js"

func Escape(s string, isqoute bool) string {
	var buf bytes.Buffer
	EscapeTo(s, &buf, isqoute)
	return buf.String()
}

func EscapeTo(s string, buf *bytes.Buffer, isqoute bool) {
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

func Join(args []string) string {
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
				EscapeTo(word, &buf, true)
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

func FileExists(nm string) bool {
	st, e := os.Stat(nm)
	if nil != e {
		return false
	}
	return !st.IsDir()
}

func RemoteExec(shell *wsman.Shell, cmd string, args []string) error {
	rname, e := SendfileIfNeed(shell, cmd)
	if nil != e {
		return e
	}

	if e := ExecCmd(shell, rname+" "+Join(args), os.Stdout, os.Stderr); nil != e {
		return e
	}
	return nil
}

func SendfileIfNeed(shell *wsman.Shell, file string) (string, error) {
	u, e := url.Parse(file)
	if nil != e || "" == u.Scheme {
		return Sendfile(shell, file)
	}

	rname := filepath.Join("tpt_scripts", u.Path)
	if s := filepath.Dir(u.Path); "" != s {
		if e := mkalldir(shell, "tpt_scripts", s); nil != e {
			return "", e
		}
	}

	if e := ExecCmd(shell, "cscript //nologo //e:jscript tpt_scripts\\"+WgetFile+" "+file+" "+rname, nil, nil); nil == e {
		return rname, nil
	}

	if st, e := os.Stat(WgetFile); nil != e {
		return "", e
	} else if st.IsDir() {
		return "", errors.New("'" + WgetFile + "' is directory.")
	}

	rget, e := Sendfile(shell, WgetFile)
	if nil != e {
		return "", errors.New("'" + WgetFile + "' send failed, " + e.Error())
	}

	if e := ExecCmd(shell, "cscript //nologo //e:jscript "+rget+" "+file+" "+rname, nil, nil); nil != e {
		return "", e
	}

	return rname, nil
}
