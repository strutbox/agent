package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"time"

	"github.com/fernet/fernet-go"
	"github.com/vmihailenco/msgpack"
)

var encoding = base64.URLEncoding

type MessageDecoder struct {
	*rsa.PrivateKey
}

type Message map[string]interface{}

func (d *MessageDecoder) Decode(b []byte) (Message, error) {
	// Messages are encoded as followed:
	//	RSA(fernetkey) || Fernet(msgpack(data))

	// RSA encrypted key is in the first 256 bytes
	encKey := b[:256]
	decKey, err := rsa.DecryptOAEP(
		sha1.New(),
		rand.Reader,
		d.PrivateKey,
		encKey,
		[]byte(""),
	)
	if err != nil {
		return nil, err
	}

	// Fernet keys must be 32 bytes long
	if len(decKey) != 32 {
		return nil, errors.New("Fernet key not the right length")
	}

	// Convert the decrypted bytes into a fernet.Key object
	key := new(fernet.Key)
	copy(key[:], decKey)

	// With our Key, we can now attempt to decode the message, which
	// is the rest of the bytes. Unfortuantely, the fernet method requires the
	// message to be b64 encoded, so we need to do that first.
	encMessage := b[256:]
	b64Message := make([]byte, encoding.EncodedLen(len(encMessage)))
	encoding.Encode(b64Message, encMessage)
	msg := fernet.VerifyAndDecrypt(b64Message, time.Minute, []*fernet.Key{key})
	if msg == nil {
		return nil, errors.New("Could not decrypt message")
	}

	// Now we have our raw msgpack payload that we can Unmarshall
	var out Message
	if err = msgpack.Unmarshal(msg, &out); err != nil {
		return nil, err
	}
	return out, nil
}
