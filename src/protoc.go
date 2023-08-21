package clipboardshare

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
)

type protoc struct {
	Prefix string
}

// @jmrxlen(data)_xxxxxx@jmrxlen(data)_xxxxxx
func (p *protoc) read(buffer io.Reader) ([]byte, error) {
	buf := bufio.NewReader(buffer)
	for {
		b0, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		if b0 != byte('@') {
			continue
		}
		b1 := make([]byte, len([]byte(p.Prefix))-1)
		n, err := buf.Read(b1)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal([]byte(p.Prefix), append([]byte{byte('@')}, b1[:n]...)) {
			continue
		}
		b2, err := buf.ReadBytes(byte('_'))
		if err != nil {
			return nil, err
		}
		lenData := bytes.TrimSuffix(b2, []byte{byte('_')})
		dataLen, err := strconv.Atoi(string(lenData))
		if err != nil {
			return nil, err
		}
		body := make([]byte, dataLen)
		n, err = buf.Read(body)
		if err != nil {
			return nil, err
		}
		if n != dataLen {
			return nil, errors.New("协议错误")
		}
		return body, nil
	}
}

func (p *protoc) pkg(data []byte) []byte {
	buf := append([]byte{}, []byte(p.Prefix)...)
	buf = append(buf, []byte(strconv.Itoa(len(data)))...)
	buf = append(buf, byte('_'))
	buf = append(buf, data...)
	return buf
}
