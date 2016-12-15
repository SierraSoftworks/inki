package server

import (
	"github.com/SierraSoftworks/girder"
	"github.com/SierraSoftworks/girder/errors"
	"github.com/SierraSoftworks/inki/crypto"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
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
}

func getAllKeys(c *girder.Context) (interface{}, error) {
	return GetAllKeys(), nil
}

func getKeysForUser(c *girder.Context) (interface{}, error) {
	return GetKeysBy(UserEquals(c.Vars["user"])), nil
}

func addKey(c *girder.Context) (interface{}, error) {
	var req crypto.Key
	if err := c.ReadBody(&req); err != nil {
		return nil, err
	}

    user := GetConfig().GetUser(req.User)
    

    req.IsValid()

	AddKey(&req)

	return req.Shorten(), nil
}
