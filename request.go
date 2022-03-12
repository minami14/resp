package resp

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
)

type Request struct {
	Command   string
	Arguments []interface{}
}

func ParseRequest(r *bufio.Reader) (*Request, error) {
	line, _, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	switch line[0] {
	case '+':
		return &Request{
			Command: string(line[1:]),
		}, nil
	case '$':
		l, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, err
		}
		bulkStr, err := readBulkString(r, l)
		if err != nil {
			return nil, err
		}
		return &Request{
			Command: bulkStr,
		}, nil
	case '*':
		l, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, err
		}
		arr, err := readArray(r, l)
		if err != nil {
			return nil, err
		}
		cmd, ok := arr[0].(string)
		if !ok {
			return nil, fmt.Errorf("command type is not string: %v", arr[0])
		}
		return &Request{
			Command:   cmd,
			Arguments: arr[1:],
		}, nil
	default:
		split := bytes.Split(line, []byte{' '})
		cmd := string(split[0])
		var args []interface{}
		for _, arg := range split[1:] {
			args = append(args, string(arg))
		}
		return &Request{
			Command:   cmd,
			Arguments: args,
		}, nil
	}
}

func readArray(r *bufio.Reader, length int) ([]interface{}, error) {
	var res []interface{}
	for i := 0; i < length; i++ {
		line, _, err := r.ReadLine()
		if err != nil {
			return nil, err
		}
		switch line[0] {
		case '*':
			l, err := strconv.Atoi(string(line[1:]))
			if err != nil {
				return nil, err
			}
			arr, err := readArray(r, l)
			if err != nil {
				return nil, err
			}
			res = append(res, arr...)
		case '$':
			l, err := strconv.Atoi(string(line[1:]))
			if err != nil {
				return nil, err
			}
			bulkStr, err := readBulkString(r, l)
			if err != nil {
				return nil, err
			}
			res = append(res, bulkStr)
		default:
			return nil, fmt.Errorf("invalid first byte %v", line)
		}
	}
	return res, nil
}

func readBulkString(r *bufio.Reader, length int) (string, error) {
	buf := make([]byte, length+2)
	_, err := r.Read(buf)
	if err != nil {
		return "", err
	}
	if !bytes.Equal(buf[length:], []byte{'\r', '\n'}) {
		return "", fmt.Errorf("invalid bulk string %v", buf)
	}
	return string(buf[:length]), nil
}
