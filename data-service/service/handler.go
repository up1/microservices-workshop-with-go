package service

import (
	"demo/db"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/sirupsen/logrus"
	"github.com/up1/microservices-workshop-with-go/common/util"
)

type Handler struct {
	dbClient        db.DbClient
	myIP            string
	isHealthy       bool
}

func NewHandler(dbClient db.DbClient) *Handler {
	myIP, err := util.ResolveIPFromHostsFile()
	if err != nil {
		myIP = util.GetIP()
	}
	return &Handler{ dbClient: dbClient, myIP: myIP, isHealthy: true}
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	var accountID = chi.URLParam(r, "accountId")

	account, err := h.dbClient.InquiryAccount(r.Context(), accountID)

	if err == nil {
		data, _ := json.Marshal(account)
		writeJSONResponse(w, http.StatusOK, data)
	} else {
		logrus.Errorf("Error reading accountID '%v' from DB: %v", accountID, err.Error())
		if err.Error() != "" {
			writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		} else {
			writeJSONResponse(w, http.StatusNotFound, []byte("Account not found"))
		}
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	dbUp := true
	if dbUp && h.isHealthy {
		data, _ := json.Marshal(healthCheckResponse{Status: "UP"})
		writeJSONResponse(w, http.StatusOK, data)
	} else {
		data, _ := json.Marshal(healthCheckResponse{Status: "Database unaccessible"})
		writeJSONResponse(w, http.StatusServiceUnavailable, data)
	}
}

func (h *Handler) SetHealthyState(w http.ResponseWriter, r *http.Request) {
	var state, err = strconv.ParseBool(chi.URLParam(r, "state"))
	if err != nil {
		logrus.Errorln("Invalid request to SetHealthyState, allowed values are true or false")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.isHealthy = state
	w.WriteHeader(http.StatusOK)
}

func writeJSONResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Connection", "close")
	w.WriteHeader(status)
	w.Write(data)
}

type healthCheckResponse struct {
	Status string `json:"status"`
}
