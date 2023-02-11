
package main

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
)

var UnfinishedErr = errors.New("Unfinished string")

func scanArgs(data []byte, atEOF bool)(i int, token []byte, err error){
	for ; i < len(data) && data[i] == ' '; i++ {}
	if i == len(data) {
		return
	}
	if data[i] == '"' {
		j := bytes.IndexByte(data[i + 1:], '"')
		if j == -1 {
			if atEOF { err = UnfinishedErr }
			return
		}
		i++
		return i + j, data[i:i + j], nil
	}
	if data[i] == '\'' {
		j := bytes.IndexByte(data[i + 1:], '\'')
		if j == -1 {
			if atEOF { err = UnfinishedErr }
			return
		}
		i++
		return i + j, data[i:i + j], nil
	}
	for j := i + 1; j < len(data); j++ {
		if data[j] == ' ' {
			return j, data[i:j], nil
		}
	}
	if !atEOF {
		return 0, nil, nil
	}
	return len(data), data[i:], nil
}

func ParseCommand(src string)(args []string, err error){
	sc := bufio.NewScanner(strings.NewReader(src))
	sc.Split(scanArgs)
	for sc.Scan() {
		args = append(args, sc.Text())
	}
	err = sc.Err()
	return
}
