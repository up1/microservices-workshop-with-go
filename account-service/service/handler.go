package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/up1/microservices-workshop-with-go/common/circuitbreaker"
	"github.com/up1/microservices-workshop-with-go/common/messaging"
	"github.com/up1/microservices-workshop-with-go/common/tracing"

	"github.com/sirupsen/logrus"
	"github.com/up1/microservices-workshop-with-go/common/util"
)

type Handler struct {
	messagingClient messaging.IMessagingClient
	client          *http.Client
	myIP            string
	isHealthy       bool
}

func NewHandler(messagingClient messaging.IMessagingClient, client *http.Client) *Handler {
	myIP, err := util.ResolveIPFromHostsFile()
	if err != nil {
		myIP = util.GetIP()
	}
	return &Handler{messagingClient: messagingClient, client: client, myIP: myIP, isHealthy: true}
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {

	// Read the 'accountId' path parameter from the mux map
	var accountID = chi.URLParam(r, "accountId")
	if accountID == "" {
		writeJSONResponse(w, http.StatusBadRequest, []byte("accountId parameter is missing"))
		return
	}

	// Call data service
	account, err := h.getAccount(r.Context(), accountID)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	if account.ID == "" {
		writeJSONResponse(w, http.StatusNotFound, []byte("Account  '"+accountID+"' not found"))
		return
	}

	account.ServedBy = h.myIP

	// Call image sercvice
    account, err = h.getImage(r.Context(), accountID)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	if account.ID == "" {
		writeJSONResponse(w, http.StatusNotFound, []byte("Account  '"+accountID+"' not found"))
		return
	}

	// Notify to report service
	h.notifyToReportService(r.Context(), account)

	data, _ := json.Marshal(account)
	writeJSONResponse(w, http.StatusOK, data)
}

func (h *Handler) getAccount(ctx context.Context, accountID string) (Account, error) {
	// Start a new opentracing child span
	child := tracing.StartSpanFromContextWithLogEvent(ctx, "getAccountData", "getAccount send")
	defer tracing.CloseSpan(child, "getAccount receive")

	// Create the http request and pass it to the circuit breaker
	req, err := http.NewRequest("GET", "http://data-service:8787/accounts/"+accountID, nil)
	body, err := circuitbreaker.PerformHTTPRequestCircuitBreaker(tracing.UpdateContext(ctx, child), "account-to-data", req)
	if err == nil {
		accountData := AccountData{}
		_ = json.Unmarshal(body, &accountData)
		return toAccount(accountData), nil
	}
	logrus.Errorf("Error: %v\n", err.Error())
	return Account{}, err
}

func (h *Handler) getImage(ctx context.Context, accountID string) (Account, error) {
	// Start a new opentracing child span
	child := tracing.StartSpanFromContextWithLogEvent(ctx, "getAccountData", "getAccount send")
	defer tracing.CloseSpan(child, "getAccount receive")

	// Create the http request and pass it to the circuit breaker
	req, err := http.NewRequest("GET", "http://image-service:7777/accounts/"+accountID, nil)
	body, err := circuitbreaker.PerformHTTPRequestCircuitBreaker(tracing.UpdateContext(ctx, child), "account-to-image", req)
	if err == nil {
		accountData := AccountData{}
		_ = json.Unmarshal(body, &accountData)
		return toAccount(accountData), nil
	}
	logrus.Errorf("Error: %v\n", err.Error())
	return Account{}, err
}

func toAccount(accountData AccountData) Account {
	return Account{
		ID: accountData.ID, Name: accountData.Name,
	}
}

func (h *Handler) notifyToReportService(ctx context.Context, account Account) {
	go func(account Account) {
		rn := ReportNotification{AccountID: account.ID, ReadAt: time.Now().UTC().String()}
		data, _ := json.Marshal(rn)
		logrus.Infof("Notifying account to report %v\n", account.ID)
		err := h.messagingClient.PublishOnQueueWithContext(ctx, data, "report_queue")
		if err != nil {
			logrus.Errorln(err.Error())
		}
		tracing.LogEventToOngoingSpan(ctx, "Sent account message")
	}(account)
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
	// Read the 'accountId' path parameter from the mux map
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
