package clipboardshare

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strconv"

	"github.com/google/uuid"
)

type protoc struct {
	Prefix   string
	pageSize int //分包大小,发送数据长度大于当前值时自动分包发送,单位字节
}

type protocType byte

const (
	heart   protocType = 0
	dataRaw protocType = 1
)

type dataType byte

const (
	txt  dataType = 0
	file dataType = 1
)

type protocFrame struct {
	Type       protocType
	Flag       string
	DataType   dataType
	HasNextPag bool
	PagIndex   int
	Data       []byte
}

func (p *protocFrame) Marshal(uid string) []byte {
	p.Flag = uid
	data := make([]byte, 0)
	data = append(data, byte(p.Type))
	data = append(data, byte(len(uid)))
	data = append(data, []byte(uid)...)
	data = append(data, byte(p.DataType))
	if p.HasNextPag {
		data = append(data, 1)
	} else {
		data = append(data, 0)
	}
	pagIndex := make([]byte, 4)
	binary.BigEndian.PutUint32(pagIndex, uint32(p.PagIndex))
	data = append(data, pagIndex...)
	data = append(data, []byte(p.Data)...)
	return data
}

func (p *protocFrame) UnMarshal(data []byte) error {
	buf := bytes.NewBuffer(data)
	b1, err := buf.ReadByte()
	if err != nil {
		return err
	}
	p.Type = protocType(b1)

	b2, err := buf.ReadByte()
	if err != nil {
		return err
	}
	b3 := make([]byte, b2)
	n, err := buf.Read(b3)
	if err != nil {
		return err
	}
	if b2 != byte(n) {
		return errors.New("标识读取失败")
	}
	p.Flag = string(b3)
	b4, err := buf.ReadByte()
	if err != nil {
		return err
	}
	p.DataType = dataType(b4)

	b5, err := buf.ReadByte()
	if err != nil {
		return err
	}
	if b5 == 1 {
		p.HasNextPag = true
	}

	b6 := make([]byte, 4)
	n2, err := buf.Read(b6)
	if err != nil {
		return err
	}
	if n2 != 4 {
		return errors.New("包序号读取失败")
	}
	p.PagIndex = int(binary.BigEndian.Uint32(b6))
	p.Data = buf.Bytes()
	return nil
}

// data 0 or 1 |标识长度[1字节]|数据包唯一标识(分包使用同一个)|数据类型(文本、文件[带文件名和后缀])|是否有下一分包[1字节]|分包序号[4字节]|分包数据
// @jmrxlen(data)_xxxxxx@jmrxlen(data)_xxxxxx
// @jmrx7_data16xxxxxxxxtxt10001dddddddddddd@jmrxlen(data)_xxxxxx
func (p *protoc) read(buffer io.ReadWriter, writer io.Writer) error {
	for {
		data, err := p.r(buffer)
		if err != nil {
			return err
		}
		frame := new(protocFrame)
		if err := frame.UnMarshal(data); err != nil {
			return err
		}
		for {
			n, err := writer.Write([]byte(frame.Data))
			if err != nil {
				return err
			}
			if n != 0 {
				frame.Data = frame.Data[n:]
			} else {
				break
			}
		}
		if !frame.HasNextPag {
			break
		} else {
			frame.Data = []byte("ok")
			err := p.w(frame.Marshal(frame.Flag), buffer)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *protoc) r(buffer io.Reader) ([]byte, error) {
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

func (p *protoc) write(frame protocFrame, rw io.ReadWriter) error {
	uid := uuid.New().String()
	data := frame.Marshal(uid)
	if len(frame.Data) > p.pageSize {
		data := bytes.NewBuffer([]byte(frame.Data))
		for {
			temp := make([]byte, p.pageSize)
			n, err := data.Read(temp)
			if err != nil {
				return err
			}
			frame.HasNextPag = n == p.pageSize
			frame.PagIndex++
			frame.Data = temp[:n]
			err = p.w(frame.Marshal(uid), rw)
			if err != nil {
				return err
			}
			if !frame.HasNextPag {
				break
			}
			data, err := p.r(rw)
			if err != nil {
				return err
			}
			rframe := new(protocFrame)
			if err := rframe.UnMarshal(data); err != nil {
				return err
			}
			if string(rframe.Data) == "ok" && frame.Flag == rframe.Flag {
				continue
			} else {
				return errors.New("包序错乱,传输失败")
			}
		}
	} else {
		err := p.w(data, rw)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *protoc) w(data []byte, writer io.Writer) error {
	pkg := p.pkg(data)
	count := 0
	for {
		n, err := writer.Write(pkg)
		if err != nil {
			return err
		}
		count += n
		if len(pkg) == count {
			break
		}
		pkg = pkg[n:]
		if n == 0 {
			break
		}
	}
	return nil
}

func (p *protoc) pkg(data []byte) []byte {
	buf := append([]byte{}, []byte(p.Prefix)...)
	buf = append(buf, []byte(strconv.Itoa(len(data)))...)
	buf = append(buf, byte('_'))
	buf = append(buf, data...)
	return buf
}
