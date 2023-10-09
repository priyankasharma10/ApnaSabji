package middlewareprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/vijaygniit/ApnaSabji/models"
	"github.com/vijaygniit/ApnaSabji/providers"
	"github.com/vijaygniit/ApnaSabji/scmerrors"
)

var secret = []byte("supersecretkey")

const (
	authorization = "Authorization"
	bearerScheme  = "bearer"
	space         = " "
	sessionHeader = "x-session-token"
	maxAge        = 300
	sessionClaims = "sessionToken"
	minimumTime   = 10
	userContext   = "userData"
)

type StructuredLogger struct{}

func NewStructuredLogger() *StructuredLogger {
	return &StructuredLogger{}
}

type Middleware struct {
	DBHelper providers.DBHelperProvider
}

func corsOptions() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Content-Length", "Host", "User-Agent", "Accept", "Accept-Encoding", "Connection"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           maxAge, // Maximum value not ignored by any of major browsers
	})
}

func NewMiddleware(dbHelper providers.DBHelperProvider) providers.MiddlewareProvider {
	return &Middleware{
		DBHelper: dbHelper,
	}
}

func (AM Middleware) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			tokenParts := strings.Split(r.Header.Get(authorization), space)
			if len(tokenParts) != 2 {
				scmerrors.RespondClientErr(w, errors.New("token not Bearer"), http.StatusUnauthorized, "Invalid token", "Invalid token")
				return
			}

			if !strings.EqualFold(tokenParts[0], bearerScheme) {
				scmerrors.RespondClientErr(w, errors.New("token not Bearer"), http.StatusUnauthorized, "Invalid token", "Invalid token")
				return
			}
			token = tokenParts[1]
			claims, err := GetClaimsFromToken(token)
			if err != nil {
				scmerrors.RespondClientErr(w, err, http.StatusUnauthorized, "GetClaimsFromToken :Invalid token", "Invalid token")
				return
			}

			SessionId, isClaimsVerified, err := AM.getUserDataFromClaims(claims)
			if err != nil {
				scmerrors.RespondClientErr(w, err, http.StatusUnauthorized, "getUserDataFromClaims: Invalid token", "Invalid token")
				return
			}

			if !isClaimsVerified {
				scmerrors.RespondClientErr(w, errors.New("invalid token"), http.StatusUnauthorized, "Invalid token", "Invalid token")
				return
			}
			err = AM.DBHelper.UpdateSession(SessionId)
			if err != nil {
				scmerrors.RespondClientErr(w, err, http.StatusUnauthorized, "UpdateSession: error updating sessions ", "UpdateSession error updating sessions ")
				return
			}

			issuer := claims["iss"].(string)
			userIDInt, err := strconv.Atoi(issuer)
			if err != nil {
				scmerrors.RespondClientErr(w, err, http.StatusUnauthorized, "UpdateSession: error updating sessions ", "UpdateSession error updating sessions ")
				return
			}
			UserData, err := AM.DBHelper.FetchUserData(userIDInt)
			if err != nil {
				scmerrors.RespondClientErr(w, err, http.StatusUnauthorized, "UpdateSession: error updating sessions ", "UpdateSession error updating sessions ")
				return
			}
			var userContextData models.UserContextData
			userContextData.UserID = userIDInt
			userContextData.Fullname = UserData.Fullname
			userContextData.Email = UserData.Email
			userContextData.Mobilenumber = UserData.Mobilenumber
			userContextData.SessionID = SessionId
			logrus.Info(userContextData)
			ctxWithUser := context.WithValue(r.Context(), models.UserContext, &userContextData)
			rWithUser := r.WithContext(ctxWithUser)
			next.ServeHTTP(w, rWithUser)

		})
	}
}

func (AM Middleware) UserFromContext(ctx context.Context) *models.UserContextData {
	return ctx.Value(models.UserContext).(*models.UserContextData)
}

func (AM Middleware) Default() chi.Middlewares {
	return chi.Chain(corsOptions().Handler)
}

func GetClaimsFromToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return jwt.MapClaims{}, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		return claims, nil
	}
	return jwt.MapClaims{}, err
}

func (AM Middleware) getUserDataFromClaims(claims jwt.MapClaims) (string, bool, error) {
	var validToken bool
	var standardClaims jwt.StandardClaims
	//var UserData models.FetchUserData
	//data := make(map[string]interface{})
	data := claims["data"].(map[string]interface{})

	var issuer, sessionID string

	issuer = claims["iss"].(string)
	//name = data["name"].(string)
	//role := data["role"].(string)
	//email := data["email"]
	UUIDToken := data["uuidToken"].(string)
	//deviceId := data["deviceId"].(string)
	//modelName := data["modelName"].(string)
	//osVersion := data["oSVersion"].(string)
	//platform := data["platform"].(string)

	UserIDInt, err := strconv.Atoi(issuer)
	if err != nil {
		logrus.Error("GetUserDataFromClaims: error converting userId string to integer ", err)
		return sessionID, validToken, errors.New(fmt.Sprintln("GetUserDataFromClaims: error converting userId string to integer & \n", err))
	}

	UserSessionsData, err := AM.DBHelper.FetchUserSessionData(UserIDInt)
	if err != nil {
		logrus.Error("GetUserDataFromClaims: error fetching user session  Data from database ", err)
		return sessionID, validToken, errors.New(fmt.Sprintln("GetUserDataFromClaims: error fetching user Data from database  & \n", err))
	}
	//fmt.Println("UserSessionsData ", UserSessionsData)
	//fmt.Println("UserIDInt ", UserIDInt)
	//fmt.Println("data ", data)
	//sessionId = UserSessionsData[0].UUIDToken

	var sessionEndTime time.Time
	for _, sessionData := range UserSessionsData {
		sessionID = sessionData.UUIDToken
		sessionEndTime = sessionData.EndTime
		if standardClaims.ExpiresAt < time.Now().Unix() && sessionID == UUIDToken && sessionEndTime.Unix() > time.Now().Unix() {
			return sessionID, true, nil
		}
		// else {
		//  	return sessionId, validToken, errors.New(fmt.Sprintln("invalid session id ", err))
		// }
	}
	return sessionID, validToken, errors.New(fmt.Sprintln("invalid session id or Session is expired", err))
}
