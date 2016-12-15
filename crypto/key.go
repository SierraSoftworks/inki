package crypto

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

type ShortKey struct {
	User        string `json:"user"`
	Fingerprint string `json:"fingerprint"`
}

type Key struct {
	Expires   time.Time `json:"expire"`
	PublicKey string    `json:"key"`
	User      string    `json:"user"`
}

func (k *Key) Validate() error {
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(k.PublicKey))
	if err != nil {
		return err
	}

	if time.Now().After(k.Expires) {
		return fmt.Errorf("key has expired")
	}

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
	out := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(out).Encode(k); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
