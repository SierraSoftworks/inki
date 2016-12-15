package server

import (
	"bytes"
	"fmt"

	"github.com/SierraSoftworks/girder"
	"github.com/SierraSoftworks/girder/errors"
	"github.com/SierraSoftworks/inki/crypto"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/openpgp"
)

var router = mux.NewRouter()

func init() {
	router.NotFoundHandler = notFoundHandler
	router.StrictSlash(true)
}

var notFoundHandler = girder.NewHandler(func(c *girder.Context) (interface{}, error) {
	log.WithFields(log.Fields{
		"url":        c.Request.URL,
		"method":     c.Request.Method,
		"user-agent": c.Request.UserAgent(),
		"headers":    c.Request.Header,
	}).Info("Route Not Found")
	return nil, errors.NotFound()
})

// Router returns the registered router for the API
func Router() *mux.Router {
	return router
}

func init() {
	Router().
		Path("/v1/keys").
		Methods("GET").
		Handler(girder.NewHandler(getAllKeys)).
		Name("GET /keys")

	Router().
		Path("/v1/keys").
		Methods("POST").
		Handler(girder.NewHandler(addKey)).
		Name("POST /keys")

	Router().
		Path("/v1/user/{user}/keys").
		Methods("GET").
		Handler(girder.NewHandler(getKeysForUser)).
		Name("GET /user/{user}/keys")

	Router().
		Path("/v1/user/{user}/authorized_keys").
		Methods("GET").
		Handler(girder.NewHandler(getAuthorizedKeysForUser)).
		Name("GET /user/{user}/authorized_keys")

	Router().
		Path("/v1/user/{user}/key/{fingerprint}").
		Methods("GET").
		Handler(girder.NewHandler(getKeyForUser)).
		Name("GET /user/{user}/key/{fingerprint}")
}

func getAllKeys(c *girder.Context) (interface{}, error) {
	return GetAllKeys(), nil
}

func getKeysForUser(c *girder.Context) (interface{}, error) {
	return GetKeysBy(UserEquals(c.Vars["user"])), nil
}

func getAuthorizedKeysForUser(c *girder.Context) (interface{}, error) {
	keys := GetKeysBy(UserEquals(c.Vars["user"]))

	b := bytes.NewBuffer([]byte{})
	for _, k := range keys {
		if err := k.Validate(); err == nil {
			b.WriteString(fmt.Sprintf("%s\n", k.PublicKey))
		}
	}

	c.ResponseHeaders.Set("Content-Type", "text/plain")
	c.Formatter = &StringFormatter{}

	return b.String(), nil
}

func getKeyForUser(c *girder.Context) (interface{}, error) {
	k := GetKeyBy(UserEquals(c.Vars["user"]).And(FingerprintEquals(c.Vars["fingerprint"])))
	if k == nil {
		return nil, errors.NotFound()
	}

	return k, nil
}

func addKey(c *girder.Context) (interface{}, error) {
	d := bytes.NewBuffer([]byte{})
	d.ReadFrom(c.Request.Body)

	reqs, err := crypto.ReadRequests(d.Bytes())
	if err != nil {
		log.WithError(err).Warn("Failed to decode armored request data")
		return nil, errors.BadRequest()
	}

	keys := []crypto.Key{}
	for _, r := range reqs {
		var key crypto.Key
		err := r.DecodeJSON(&key)
		if err != nil {
			log.WithError(err).Warn("Failed to decode JSON in request body")
			return nil, errors.BadRequest()
		}

		log.WithFields(log.Fields{
			"user":   key.User,
			"expire": key.Expires,
			"key":    key.PublicKey,
		}).Debug("Decoded key information")

		if err := key.Validate(); err != nil {
			log.WithError(err).Warn("Key data was not in a valid format, or has expired")
			return nil, errors.BadRequest()
		}

		user := GetConfig().GetUser(key.User)
		if user == nil {
			log.WithField("user", key.User).Warn("No configuration entry for this user")
			return nil, errors.NotAllowed()
		}

		kr, err := user.GetKeyRing()
		if err != nil {
			log.WithError(err).Warn("Could not load user's keyring")
			return nil, errors.ServerError()
		}

		s := bytes.NewBuffer([]byte{})
		s.ReadFrom(r.Signature.Body)

		signer, err := openpgp.CheckDetachedSignature(kr, bytes.NewBuffer(r.Payload), s)
		if err != nil {
			log.WithError(err).Warn("Failed to check request signature")
			return nil, errors.Unauthorized()
		}

		if signer == nil {
			log.Warn("No signatory found for the request")
			return nil, errors.Unauthorized()
		}

		log.WithFields(log.Fields{
			"user":   key.User,
			"key":    key.PublicKey,
			"expire": key.Expires,
		}).Debug("Accepted new key")
		keys = append(keys, key)
	}

	for _, k := range keys {
		AddKey(&k)
	}

	return keys, nil
}
