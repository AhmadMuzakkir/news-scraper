package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/ahmadmuzakkir/scrapenews/store"
	"github.com/go-chi/chi"
)

type NewsHandler struct {
	Logger    *log.Logger
	newsStore store.NewsStore
}

func NewNewsHandler(newsStore store.NewsStore) *NewsHandler {
	return &NewsHandler{newsStore: newsStore}
}

func (n *NewsHandler) Routes() chi.Router {
	router := chi.NewRouter()

	router.Get("/get", n.get)
	return router
}

func (n *NewsHandler) get(w http.ResponseWriter, r *http.Request) {
	fromDatetimeStr := r.FormValue("from")
	untilDatetimeStr := r.FormValue("until")

	var fromDatetime = time.Now()
	var untilDatetime = time.Now().AddDate(0, 0, -1)

	var err error

	if fromDatetimeStr != "" {
		fromDatetime, err = time.Parse(time.RFC3339, fromDatetimeStr)
		if err != nil {
			fromDatetime = time.Now()
		}
	}

	if untilDatetimeStr != "" {
		untilDatetime, err = time.Parse(time.RFC3339, untilDatetimeStr)
		if err != nil {
			untilDatetime = time.Now().AddDate(0, 0, -1)
		}
	}

	list, err := n.newsStore.GetAll(fromDatetime, untilDatetime)
	if err != nil {
		n.logError("latest: %s", err)
		n.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
	}

	n.render(w, http.StatusOK, list)
}

func (n *NewsHandler) render(w http.ResponseWriter, status int, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		n.logError("marshal json: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonData)
}

func (n *NewsHandler) renderError(w http.ResponseWriter, status int, code, message string) {
	response := struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	response.Error.Code = code
	response.Error.Message = message
	n.render(w, status, response)
}

func (n *NewsHandler) logError(format string, a ...interface{}) {
	pc, _, _, _ := runtime.Caller(1)
	callerNameSplit := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	funcName := callerNameSplit[len(callerNameSplit)-1]
	n.Logger.Printf("ERROR: %s: %s", funcName, fmt.Sprintf(format, a...))
}
