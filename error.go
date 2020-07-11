package main

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	// UNEXPECT_ERROR
	UNEXPECT_ERROR = "11000"
	// CHANNEL_ERROR
	CHANNEL_ERROR = "11001"
	// PARAMS_ERROR
	PARAMS_ERROR = "11002"
)

type Response struct {
	Status string                   `json:"status"`
	Data   map[string][]interface{} `json:"data"`
	Error  map[string]interface{}   `json:"error"`
}

// ErrorRespone 直接回傳api錯誤信息
func ErrorRespone(err error, errorCode string, w http.ResponseWriter, statusCode int) bool {
	if err != nil {
		var res Response
		res.Status = "error"
		res.Data = map[string][]interface{}{}
		res.Error = map[string]interface{}{}

		res.Error["message"] = err.Error()
		res.Error["error_code"] = errorCode

		w.WriteHeader(statusCode)
		log.Println(err)

		json.NewEncoder(w).Encode(res)

		return true
	}

	return false
}
