package crypto

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/ssh"
)

type ShortKey struct {
	User        string `json:"user"`
	Fingerprint string `json:"fingerprint"`
}

type signingData struct {
	Expires   time.Time `json:"expire"`
	PublicKey string    `json:"key"`
	User      string    `json:"user"`
}

type Key struct {
	Expires   time.Time `json:"expire"`
	PublicKey string    `json:"key"`
	User      string    `json:"user"`
	Signature string    `json:"signature"`
}

func (k *Key) IsValid(keyring openpgp.KeyRing) bool {
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(k.PublicKey))
	if err != nil {
		return false
	}

	data := bytes.NewBuffer([]byte{})
	json.NewEncoder(data).Encode(&struct {
		Expires   time.Time `json:"expire"`
		PublicKey string    `json:"key"`
		User      string    `json:"user"`
	}{
		Expires:   k.Expires,
		PublicKey: k.PublicKey,
		User:      k.User,
	})
	sigReader := strings.NewReader(k.Signature)
	_, err = openpgp.CheckDetachedSignature(keyring, data, sigReader)
	if err != nil {
		return false
	}

	return true
}

func (k *Key) Sign(signer *openpgp.Entity) error {
	keyData, err := k.SigningData()
	if err != nil {
		return err
	}

	out := bytes.NewBuffer([]byte{})
	keyReader := bytes.NewBuffer(keyData)
	err = openpgp.DetachSign(out, signer, keyReader, nil)
	if err != nil {
		return err
	}

	k.Signature = out.String()

	return nil
}

func (k *Key) Fingerprint() string {
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(k.PublicKey))
	if err != nil {
		return ""
	}

	h := md5.New()
	h.Write(key.Marshal())
	return fmt.Sprintf("%0x", h.Sum(nil))
}

func (k *Key) Equals(key *Key) bool {
	return k.User == key.User && k.PublicKey == key.PublicKey
}

func (k *Key) Shorten() *ShortKey {
	return &ShortKey{
		User:        k.User,
		Fingerprint: k.Fingerprint(),
	}
}

func (k *Key) SigningData() ([]byte, error) {
	data := &signingData{
		User:      k.User,
		PublicKey: k.PublicKey,
		Expires:   k.Expires,
	}

	out := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(out).Encode(data); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
