package main

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"io/ioutil"

	"github.com/vmihailenco/msgpack"
)

type MessageDecoder struct {
	*rsa.PrivateKey
}

type Message map[string]interface{}

func (d *MessageDecoder) Decode(b []byte) (Message, error) {
	// Messages are encoded as followed:
	//	RSA(zlib(msgpack(message)))
	b, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, d.PrivateKey, b, []byte(""))
	if err != nil {
		return nil, err
	}
	reader, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	b, err = ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var out Message
	if err = msgpack.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}
