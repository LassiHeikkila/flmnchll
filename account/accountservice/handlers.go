package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/LassiHeikkila/flmnchll/account/accountdb"
	"github.com/LassiHeikkila/flmnchll/helpers/httputils"
)

const (
	maxSignupBodyLen = 1024

	FormKeyUsername = "username"
	FormKeyPassword = "password"
)

func SignupHandler(w http.ResponseWriter, req *http.Request) {
	// POST, read body
	// Response will be JSON containing possible error message or id of newly created account

	// Sign up form should contain:
	// - username (unique)
	// - password

	// Don't want or need email, phone, personal info
	req.ParseForm()

	username := req.Form.Get(FormKeyUsername)
	password := req.Form.Get(FormKeyPassword)

	taken, err := accountdb.UsernameTaken(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(databaseError))
		return
	}
	if taken {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(usernameTaken))
		return
	}

	pwHash := HashPassword(password)

	id, err := accountdb.CreateAccount(
		accountdb.Account{
			Username:     username,
			PasswordHash: pwHash,
		},
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte(fmt.Sprintf(accountCreationSuccessFmt, id)))
}

func LoginHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	username := req.Form.Get(FormKeyUsername)
	password := req.Form.Get(FormKeyPassword)

	a, err := accountdb.GetAccountByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if !PasswordEqualsHash(password, a.PasswordHash) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	token, err := accountdb.CreateNewTokenForUserID(a.UserID, time.Now().Add(24*time.Hour))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(token))
}

func AccountInfoHandler(w http.ResponseWriter, req *http.Request) {
	// GET
	// account id in URL variables
	id := mux.Vars(req)["id"]

	if !tokenMatchesUserId(httputils.GetAuthToken(req), id) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(forbiddenError))
		return
	}

	a, err := accountdb.GetAccount(id)
	if errors.Is(err, accountdb.ErrAccountNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(accountWithIdNotFound))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(databaseError))
		return
	}

	e := json.NewEncoder(w)
	_ = e.Encode(a)
}

func AccountLookupHandler(w http.ResponseWriter, req *http.Request) {
	// GET
	// account token in URL variables
	token := mux.Vars(req)["token"]

	userId, err := accountdb.AuthenticateToken(token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(databaseError))
		return
	}

	a, err := accountdb.GetAccount(userId)
	if errors.Is(err, accountdb.ErrAccountNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(accountWithIdNotFound))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(databaseError))
		return
	}

	e := json.NewEncoder(w)
	_ = e.Encode(a)
}

func AccountInfoUpdateHandler(w http.ResponseWriter, req *http.Request) {
	// PUT
	// account id in URL variables
	id := mux.Vars(req)["id"]

	if !tokenMatchesUserId(httputils.GetAuthToken(req), id) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(forbiddenError))
		return
	}

	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(unimplementedError))

	// TODO: implement account detail updates
}

func AccountInfoDeleteHandler(w http.ResponseWriter, req *http.Request) {
	// DELETE
	// account id in URL variables
	id := mux.Vars(req)["id"]

	if !tokenMatchesUserId(httputils.GetAuthToken(req), id) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(forbiddenError))
		return
	}

	err := accountdb.DeleteAccount(id)
	if errors.Is(err, accountdb.ErrAccountNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(accountWithIdNotFound))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(databaseError))
		return
	}

	_, _ = w.Write([]byte(genericOK))
}

func TokenAuthenticateHandler(w http.ResponseWriter, req *http.Request) {
	// GET, check token from header

	// Auth middleware handles the check,
	// so if we're here, we're authenticated
	_, _ = w.Write([]byte(genericOK))
}

func TokenDeauthenticateHandler(w http.ResponseWriter, req *http.Request) {
	// POST, read body
	// body contains tokens to deauthenticate, one per line
	// also accept wildcard "*" to deauthenticate every token for the user
	authToken := httputils.GetAuthToken(req)
	// ignore error, auth middleware has already done the work once, we just need to get the account id again
	userId, _ := accountdb.AuthenticateToken(authToken)

	defer req.Body.Close()
	err := invalidateTokens(req.Body, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(genericErrorFmt, err.Error())))
		return
	}

	_, _ = w.Write([]byte(genericOK))
}
