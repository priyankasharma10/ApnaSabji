package providers

import "github.com/vijaygniit/ApnaSabji/models"

type DBHelperProvider interface {
	CreateNewUser(newUserRequest *models.CreateNewUserRequest, userID int) (*int, error)
	IsUserAlreadyExists(emailID string) (isUserExist bool, user models.UserData, err error)
	UpdateSession(sessionId string) error
	FetchUserData(userID int) (models.FetchUserData, error)
	FetchUserSessionData(userID int) ([]models.FetchUserSessionsData, error)
	IsPhoneNumberAlreadyExist(mobilenumber string) (bool, error)
	GetUserInfoByEmail(email string) (models.GetUserDataByEmail, error)
	LogInUserUsingEmail(loginReq models.EmailAndOTP) (userID int, message string, err error)
	StartNewSession(userID int, request *models.CreateSessionRequest) (string, error)
	GenerateAndStoreOTP(email string, mobilenumber string)( models.GenerateAndStoreOTP)
	// EndSession(sessionId string) error
}
