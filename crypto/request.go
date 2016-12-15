package crypto

import (
	"fmt"

	"bytes"

	"encoding/json"

	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/clearsign"
	"golang.org/x/crypto/openpgp/packet"
)

type Request struct {
	Payload   []byte
	Signature *armor.Block
}

func (r *Request) EncodeJSON(from interface{}) error {
	b := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(b).Encode(from)
	if err != nil {
		return err
	}

	r.Payload = b.Bytes()
	return nil
}

func (r *Request) DecodeJSON(into interface{}) error {
	return json.NewDecoder(bytes.NewBuffer(r.Payload)).Decode(into)
}

func ReadRequests(data []byte) ([]Request, error) {
	reqs := []Request{}
	for d := data; len(d) > 0; {
		b, r := clearsign.Decode(d)
		if b == nil {
			break
		}
		d = r

		reqs = append(reqs, Request{
			Payload:   b.Bytes,
			Signature: b.ArmoredSignature,
		})
	}

	if len(reqs) == 0 && len(data) > 0 {
		return nil, fmt.Errorf("couldn't find a valid requests in the data")
	}

	return reqs, nil
}

func WriteRequests(reqs []Request, key *packet.PrivateKey) ([]byte, error) {
	out := bytes.NewBuffer([]byte{})

	for _, r := range reqs {
		w, err := clearsign.Encode(out, key, nil)
		if err != nil {
			return nil, err
		}

		w.Write(r.Payload)
		w.Close()
	}

	return out.Bytes(), nil
}
