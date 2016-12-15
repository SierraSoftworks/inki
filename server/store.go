package server

import (
	"sync"
	"time"

	"github.com/SierraSoftworks/inki/crypto"
)

var keyStore []crypto.Key
var keyStoreLock sync.Mutex

type KeyPredicate func(k *crypto.Key) bool

func (p KeyPredicate) And(pred KeyPredicate) KeyPredicate {
	return func(k *crypto.Key) bool {
		return p(k) && pred(k)
	}
}

func (p KeyPredicate) Or(pred KeyPredicate) KeyPredicate {
	return func(k *crypto.Key) bool {
		return p(k) || pred(k)
	}
}

func init() {
	keyStore = []crypto.Key{}
}

func HasKey(key *crypto.Key) bool {
	keyStoreLock.Lock()
	defer keyStoreLock.Unlock()

	for _, k := range keyStore {
		if k.Equals(key) {
			return true
		}
	}

	return false
}

func AddKey(key *crypto.Key) {
	keyStoreLock.Lock()
	defer keyStoreLock.Unlock()

	for i, k := range keyStore {
		if k.Equals(key) {
			// Update the expiry time
			k.Expires = key.Expires
			keyStore[i] = k
			return
		}
	}

	keyStore = append(keyStore, *key)
}

func GetAllKeys() []crypto.Key {
	return keyStore
}

func GetKeyBy(pred KeyPredicate) *crypto.Key {
	keyStoreLock.Lock()
	defer keyStoreLock.Unlock()

	for _, k := range keyStore {
		if pred(&k) {
			return &k
		}
	}

	return nil
}

func GetKeysBy(pred KeyPredicate) []crypto.Key {
	keyStoreLock.Lock()
	defer keyStoreLock.Unlock()

	results := []crypto.Key{}
	for _, k := range keyStore {
		if pred(&k) {
			results = append(results, k)
		}
	}

	return results
}

func RemoveKey(key *crypto.Key) {
	RemoveKeyBy(func(k *crypto.Key) bool {
		return k.Equals(key)
	})
}

func RemoveKeyBy(pred KeyPredicate) {
	keyStoreLock.Lock()
	defer keyStoreLock.Unlock()

	for i, k := range keyStore {
		if pred(&k) {
			keyStore = append(keyStore[:i-1], keyStore[i+1:]...)
		}
	}
}

func KeyEquals(key *crypto.Key) KeyPredicate {
	return func(k *crypto.Key) bool {
		return k.Equals(key)
	}
}

func KeyValid() KeyPredicate {
	return func(k *crypto.Key) bool {
		return k.Expires.After(time.Now())
	}
}

func KeyExpired() KeyPredicate {
	return func(k *crypto.Key) bool {
		return k.Expires.Before(time.Now())
	}
}

func UserEquals(user string) KeyPredicate {
	return func(k *crypto.Key) bool {
		return k.User == user
	}
}

func FingerprintEquals(fingerprint string) KeyPredicate {
	return func(k *crypto.Key) bool {
		return k.Fingerprint() == fingerprint
	}
}
