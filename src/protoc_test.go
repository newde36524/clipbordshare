package clipboardshare

import (
	"bytes"
	"fmt"
	"testing"
)

func TestXxx(t *testing.T) {
	prot := protoc{
		Prefix:   "@jmrx#@!%",
		pageSize: 1,
	}
	content := "zsk"
	f := protocFrame{
		Type:     dataRaw,
		DataType: txt,
		Data:     []byte(content),
	}
	data := f.Marshal("00000000000")
	buf := bytes.NewBuffer(nil)
	err := prot.w(data, buf)
	if err != nil {
		panic(err)
	}

	f2, err := prot.r(buf)
	if err != nil {
		panic(err)
	}
	if content != string(f2.Data) {
		t.Error()
	}
	fmt.Println(string(f2.Data))
}
