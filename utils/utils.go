package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

func EncodeJSONBody(resp http.ResponseWriter, statusCode int, data interface{}) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(statusCode)
	err := json.NewEncoder(resp).Encode(data)
	if err != nil {
		logrus.Errorf("EncodeJSONBody : Error encoding response %v", err)
	}
}
