package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/LassiHeikkila/flmnchll/account/accountservice/accountclient"
	"github.com/LassiHeikkila/flmnchll/helpers/httputils"
	"github.com/LassiHeikkila/flmnchll/room/roomdb"
)

// Needed API endpoints:
// - creating a new room
//  - does not need to take any arguments
// - user joining a room
//  - get user id from auth token
//  - add user to room
//  - return room id
// - user get room details
// 	- return details if user is member in room
//   - check auth token
// - user get own current room (if any)
// - delete room
//  - only admin or room creator can do it
// - user leaving a room
//  - get user id from auth token
//  - add user to room
// - room having content selected

func CreateRoomHandler(w http.ResponseWriter, req *http.Request) {
	// select peer server for the room
	peerServer, err := SelectAvailablePeerServer()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(peerServerUnavailable))
		return
	}

	// check user is authenticated and get their id
	userID, err := accountclient.ValidateUserToken(httputils.GetAuthToken(req))
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(forbiddenError))
		return
	}

	// get user object from roomdb
	u, err := roomdb.GetUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(userWithIdNotFound))
		return
	}

	r, err := roomdb.CreateRoom(*u, peerServer.Address())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(databaseError))
		return
	}

	e := json.NewEncoder(w)
	e.Encode(NewOKResponseWithDetails(map[string]any{
		"roomID":  r.ID,
		"shortID": r.ShortID,
	}))
}

func DeleteRoomHandler(w http.ResponseWriter, req *http.Request) {
	// check user is authenticated and get their id
	userID, err := accountclient.ValidateUserToken(httputils.GetAuthToken(req))
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(forbiddenError))
		return
	}

	// get room id from query parameters
	// account id in URL variables
	roomID := mux.Vars(req)["id"]

	u, err := roomdb.GetUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(userWithIdNotFound))
		return
	}

	r, err := roomdb.GetRoom(roomID, true)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(roomWithIdNotFound))
		return
	}

	if r.Owner.ID != u.ID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(forbiddenError))
		return
	}

	err = roomdb.DeleteRoom(r.ID, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(databaseError))
		return
	}

	w.Write([]byte(genericOK))
}

func SetRoomSelectedContentHandler(w http.ResponseWriter, req *http.Request) {
	// for now, only the owner can select the content
	// check user is authenticated and get their id
	userID, err := accountclient.ValidateUserToken(httputils.GetAuthToken(req))
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(forbiddenError))
		return
	}

	// get room id from query parameters
	// account id in URL variables
	roomID := mux.Vars(req)["roomID"]
	contentID := mux.Vars(req)["contentID"]

	u, err := roomdb.GetUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(userWithIdNotFound))
		return
	}

	r, err := roomdb.GetRoom(roomID, true)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(roomWithIdNotFound))
		return
	}

	if r.Owner.ID != u.ID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(forbiddenError))
		return
	}

	r.SelectedContentId = contentID
	err = roomdb.UpdateRoom(*r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(databaseError))
		return
	}

	w.Write([]byte(genericOK))
}

func JoinRoomHandler(w http.ResponseWriter, req *http.Request) {
	// // check user is authenticated and get their id
	// userID, err := accountclient.ValidateUserToken(httputils.GetAuthToken(req))
	// if err != nil {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	w.Write([]byte(forbiddenError))
	// 	return
	// }

	// get room id from path
	// username also in path for now
	roomID := mux.Vars(req)["id"]
	username := mux.Vars(req)["username"]

	u, _ := roomdb.GetUserByName(username)
	if u == nil {
		u, _ = roomdb.CreateUser(roomdb.GenerateUUID(), username)
	}

	r, err := roomdb.GetRoom(roomID, true)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(roomWithIdNotFound))
		return
	}

	err = u.JoinRoom(r.ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(databaseError))
		return
	}

	w.Write([]byte(genericOK))
}

func LeaveRoomHandler(w http.ResponseWriter, req *http.Request) {
	// // check user is authenticated and get their id
	// userID, err := accountclient.ValidateUserToken(httputils.GetAuthToken(req))
	// if err != nil {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	w.Write([]byte(forbiddenError))
	// 	return
	// }
	username := mux.Vars(req)["username"]

	u, err := roomdb.GetUserByName(username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(userWithIdNotFound))
		return
	}

	err = u.LeaveRoom()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(databaseError))
		return
	}

	w.Write([]byte(genericOK))
}

func GetRoomDetailsHandler(w http.ResponseWriter, req *http.Request) {
	// // check user is authenticated and get their id
	// userID, err := accountclient.ValidateUserToken(httputils.GetAuthToken(req))
	// if err != nil {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	w.Write([]byte(forbiddenError))
	// 	return
	// }

	// get room id from query parameters
	// account id in URL variables
	roomID := mux.Vars(req)["id"]

	// check that user is in the room
	r, err := roomdb.GetRoom(roomID, true)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(roomWithIdNotFound))
		return
	}

	// userIsMember := func() bool {
	// 	for _, u := range r.Members {
	// 		if u.ID == userID {
	// 			return true
	// 		}
	// 	}
	// 	return false
	// }

	// userIsOwner := func() bool {
	// 	return r.Owner.ID == userID
	// }

	// if !userIsMember() || !userIsOwner() {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	w.Write([]byte(forbiddenError))
	// 	return
	// }

	// return room details
	e := json.NewEncoder(w)

	_ = e.Encode(roomToJSON(r))
}

func GetCurrentRoomHandler(w http.ResponseWriter, req *http.Request) {
	// check user is authenticated and get their id
	userID, err := accountclient.ValidateUserToken(httputils.GetAuthToken(req))
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(forbiddenError))
		return
	}

	// get user object
	u, err := roomdb.GetUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(userWithIdNotFound))
		return
	}

	// get the room based on room id
	roomID := u.RoomID
	if roomID == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(roomWithIdNotFound))
		return
	}

	r, err := roomdb.GetRoom(roomID, false)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(roomWithIdNotFound))
		return
	}

	// return room json
	e := json.NewEncoder(w)

	_ = e.Encode(roomToJSON(r))
}
