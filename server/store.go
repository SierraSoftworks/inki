package server

import (
	"sync"

	"github.com/SierraSoftworks/inki/crypto"
)

var keyStore []crypto.Key
var keyStoreLock sync.Mutex

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

	if !HasKey(key) {
		keyStore = append(keyStore, *key)
	}
}

func GetAllKeys() []crypto.Key {
	return keyStore
}

func GetKeysBy(pred func(k *crypto.Key) bool) []crypto.Key {
	keyStoreLock.Lock()
	defer keyStoreLock.Unlock()

	results := []Key{}
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

func RemoveKeyBy(pred func(k *crypto.Key) bool) {
	keyStoreLock.Lock()
	defer keyStoreLock.Unlock()

	for i, k := range keyStore {
		if pred(&k) {
			keyStore = append(keyStore[:i-1], keyStore[i+1:]...)
		}
	}
}

func KeyEquals(key *crypto.Key) func(k *crypto.Key) bool {
	return func(k *Key) bool {
		return k.Equals(key)
	}
}

func UserEquals(user string) func(k *crypto.Key) bool {
	return func(k *Key) bool {
		return k.User == user
	}
}

func FingerprintEquals(fingerprint string) func(k *crypto.Key) bool {
	return func(k *crypto.Key) bool {
		return k.Fingerprint() == fingerprint
	}
}
