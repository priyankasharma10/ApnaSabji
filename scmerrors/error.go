package scmerrors

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type clientError struct {
	Err           string `json:"error"`
	MessageToUser string `json:"messageToUser"`
	DeveloperInfo string `json:"developerInfo"`
	StatusCode    int    `json:"statusCode"`
	IsClientError bool   `json:"isClientError"`
} // @name clientError

func RespondClientErr(resp http.ResponseWriter, err error, statusCode int, messageToUser, developerInfo string) {
	resp.WriteHeader(statusCode)

	clientErr := &clientError{
		MessageToUser: messageToUser,
		DeveloperInfo: developerInfo,
		Err:           err.Error(),
		StatusCode:    statusCode,
		IsClientError: true,
	}

	if err := json.NewEncoder(resp).Encode(clientErr); err != nil {
		logrus.Error(err)
	}
}
func RespondGenericServerErr(resp http.ResponseWriter, err error, developerInfo string) {
	resp.WriteHeader(http.StatusInternalServerError)
	clintErr := &clientError{
		DeveloperInfo: developerInfo,
		Err:           err.Error(),
		StatusCode:    http.StatusInternalServerError,
		IsClientError: false,
	}
	if err := json.NewEncoder(resp).Encode(clintErr); err != nil {
		logrus.Error(err)
	}
}
