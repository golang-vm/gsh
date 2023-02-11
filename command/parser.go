
package sh_command

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
)

var UnfinishedErr = errors.New("Unfinished string")

func ParseCommand(src string)(args []string, err error){
	sc := bufio.NewScanner(strings.NewReader(src))
	sc.Split(scanArgs)
	for sc.Scan() {
		args = append(args, sc.Text())
	}
	err = sc.Err()
	return
}

func scanArgs(data []byte, atEOF bool)(i int, token []byte, err error){
	for ; i < len(data) && data[i] == ' '; i++ {}
	if i == len(data) {
		return
	}
	if data[i] == '"' {
		j := bytes.IndexByte(data[i + 1:], '"')
		if j == -1 {
			if atEOF { err = UnfinishedErr }
			return i, nil, nil
		}
		i++
		return i + j, unescape(data[i:i + j]), nil
	}
	if data[i] == '\'' {
		j := bytes.IndexByte(data[i + 1:], '\'')
		if j == -1 {
			if atEOF { err = UnfinishedErr }
			return i, nil, nil
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

func unescape(src []byte)(res []byte){
	return src // TODO
	// if bytes.IndexByte(src, '\\') == -1 {
	// 	return src
	// }
	// res = make([]byte, 0, len(src))
	// for {
	// 	i := bytes.IndexByte(src, '\\')
	// 	if i == -1 {
	// 		return
	// 	}
	// 	res = append(res, src[:i]...)
	// 	i++
	// 	c := src[i]
	// 	src = src[i + 1:]
	// 	switch {
	// 	// case c == 'x':
	// 	// 	res = append(res, parseHex(src[0:2]))
	// 	case c == '\\' || c == '\'' || c == '"':
	// 		res = append(res, c)
	// 	case 'a' <= c && c <= 'z':
	// 		c -= 'a' + 'A'
	// 		fallthrough
	// 	case 'A' <= c && c <= '[':
	// 		c -= 'A' - 1
	// 	default:
	// 		res = append(res, '\\', c)
	// 	}
	// }
}
