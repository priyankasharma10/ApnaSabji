package authProvider

import (
	"github.com/sirupsen/logrus"
	"github.com/vijaygniit/ApnaSabji/models"
	"github.com/volatiletech/null"

	//"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)


var secret = []byte("supersecretkey")

type JWTClaim struct {
	Platform  string      `json:"platform"`
	ModelName null.String `json:"modelName"`
	OSVersion null.String `json:"osVersion"`
	DeviceID  null.String `json:"deviceId"`
	Username  string      `json:"username"`
	Email     string      `json:"email"`
	UUIDToken string      `json:"token"`
	jwt.StandardClaims
}

func GenerateJWT(devClaims map[string]interface{}) (tokenString string, err error) {
	// var userInfo models.GetUserDataByEmail
	var userSessionData models.CreateSessionRequest
	var ok bool
	var UUIDToken string
	userInfo, ok := devClaims["userInfo"].(models.GetUserDataByEmail)
	if !ok {
		logrus.Error("GenerateJWT:  error getting values out of the devClaims map 1")
	}
	UUIDToken, ok = devClaims["UUIDToken"].(string)
	if !ok {
		logrus.Error("GenerateJWT:  error getting values out of the devClaims map 2")
	}
	userSessionData, ok = devClaims["UserSession"].(models.CreateSessionRequest)
	if !ok {
		logrus.Error("GenerateJWT:  error getting values out of the devClaims map 3")
	}
	UserIDString := strconv.Itoa(userInfo.UserID)
	expirationTime := time.Now().Add(1 * time.Hour)

	claims := &jwt.MapClaims{
		"iss": UserIDString,
		"exp": time.Now().Add(time.Hour).Unix(),
		"data": map[string]string{
			"id":        UserIDString,
			"name":      userInfo.Fullname,
			"modelName": userSessionData.ModelName,
			"platform":  userSessionData.Platform,
			"oSVersion": userSessionData.OSVersion,
			"deviceId":  userSessionData.DeviceID,
			"email":     userInfo.Email,
			"uuidToken": UUIDToken,
			"expiresAt": expirationTime.String(),
			"issuer":    UserIDString,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(secret)
	return
}
