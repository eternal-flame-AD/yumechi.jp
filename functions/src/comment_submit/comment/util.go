package comment

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
)

func refStr(str string) *string {
	return &str
}

func SHA1Bytes(p []byte) []byte {
	s := sha1.New()
	s.Write(p)
	return s.Sum(nil)
}

func Base64Bytes(p []byte) string {
	var b bytes.Buffer
	base64.NewEncoder(base64.StdEncoding, &b).Write(p)
	return b.String()
}
