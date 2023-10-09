package dbhelperprovider

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/vijaygniit/ApnaSabji/models"
)

func (dh *DBHelper) CreateNewUser(newUserRequest *models.CreateNewUserRequest, userID int) (*int, error) {
	var newUserID int

	SQL := `
		INSERT INTO users
		(fullname, email, mobilenumber, created_at, created_by)
		VALUES (trim($1), lower(trim($2)), $3, $4, $5)
		RETURNING id
	`

	args := []interface{}{
		newUserRequest.Fullname,
		newUserRequest.Email.String,
		newUserRequest.Mobilenumber.String,
		time.Now().UTC(),
		userID,
	}

	err := dh.DB.Get(&newUserID, SQL, args...)
	if err != nil {
		logrus.Errorf("CreateNewUser: error creating user %v", err)
		return nil, err
	}

	SQL = `
		INSERT INTO user_profiles
		(user_id)
		VALUES ($1)
	`

	_, err = dh.DB.Exec(SQL, newUserID)
	if err != nil {
		logrus.Errorf("CreateNewUser: error creating user profile %v", err)
		return nil, err
	}

	return &newUserID, nil
}

func (dh *DBHelper) IsPhoneNumberAlreadyExist(mobilenumber string) (bool, error) {
	// language=sql
	SQL := `SELECT count(*) > 0 
            FROM users
            WHERE archived_at IS NULL
            AND mobilenumber  = $1
            AND deactivated IS FALSE`

	var isPhoneAlreadyExist bool
	err := dh.DB.Get(&isPhoneAlreadyExist, SQL, mobilenumber)
	if err != nil {
		logrus.Errorf("IsPhoneNumberAlreadyExist: error getting whether phone exist: %v", err)
		return isPhoneAlreadyExist, err
	}

	return isPhoneAlreadyExist, nil
}

func (dh *DBHelper) IsUserAlreadyExists(emailID string) (isUserExist bool, user models.UserData, err error) {
	//	language=sql
	SQL := `SELECT id
			FROM users
			WHERE email = lower($1)
			  AND archived_at IS NULL
				AND deactivated IS FALSE `

	err = dh.DB.Get(&user, SQL, emailID)
	if err != nil && err != sql.ErrNoRows {
		logrus.Errorf("isEmailAlreadyExist: unable to get user from email %v", err)
		return false, user, err
	}

	if err == sql.ErrNoRows {
		return false, user, nil
	}

	return true, user, nil
}

func (dh *DBHelper) FetchUserSessionData(userID int) ([]models.FetchUserSessionsData, error) {
	SQL := `
		SELECT id, user_id, end_time, token
		FROM sessions
		WHERE user_id = $1
	`

	fetchUserSessionData := make([]models.FetchUserSessionsData, 0)
	err := dh.DB.Select(&fetchUserSessionData, SQL, userID)
	if err != nil {
		logrus.Errorf("FetchUserSessionData: error getting user session data from database: %v", err)
		return fetchUserSessionData, err
	}
	return fetchUserSessionData, nil
}

func (dh *DBHelper) UpdateSession(sessionID string) error {
	SQL := `
		UPDATE sessions
		SET end_time = $2
		WHERE token = $1
	`

	_, err := dh.DB.Exec(SQL, sessionID, time.Now().Add(1*time.Hour))
	if err != nil {
		logrus.Errorf("UpdateSession: error updating user session data in the database: %v", err)
		return err
	}
	return nil
}

func (dh *DBHelper) FetchUserData(userID int) (models.FetchUserData, error) {
	var fetchUserData models.FetchUserData

	SQL := `
		SELECT id, fullname, email, mobilenumber
		FROM users
		WHERE id = $1
	`

	err := dh.DB.Get(&fetchUserData, SQL, userID)
	if err != nil {
		logrus.Errorf("FetchUserData: error getting user data: %v", err)
		return fetchUserData, err
	}

	return fetchUserData, nil
}

func (dh *DBHelper) GetUserInfoByEmail(email string) (models.GetUserDataByEmail, error) {
	//language=sql
	SQL := `SELECT  users.id, users.fullname, mobilenumber
			FROM users `
	var getUserDataByEmail models.GetUserDataByEmail
	err := dh.DB.Get(&getUserDataByEmail, SQL, email)
	if err != nil && err != sql.ErrNoRows {
		logrus.Errorf("GetUserInfoByEmail: error getting user data: %v", err)
		return getUserDataByEmail, err
	}
	if err == sql.ErrNoRows {
		return getUserDataByEmail, errors.New("email does not exist")
	}
	return getUserDataByEmail, nil
}

func (dh *DBHelper) LogInUserUsingEmail(loginReq models.EmailAndOTP) (userID int, message string, err error) {
	// language=SQL
	SQL := `SELECT 	id,   
					otp
			FROM users
		WHERE email = $1
		AND deactivated IS FALSE
		AND archived_at IS NULL`

	var user = struct {
		ID  int    `db:"id"`
		OTP string `db:"otp"`
	}{}

	if err = dh.DB.Get(&user, SQL, loginReq.Email); err != nil && err != sql.ErrNoRows {
		logrus.Errorf("LogInUserUsingEmail: error while getting user %v", err)
		return userID, "error getting user", err
	}

	if user.OTP != loginReq.OTP {
		return userID, "OTP Not Correct", errors.New("OTP not matched")
	}

	return user.ID, "", nil
}

func (dh *DBHelper) StartNewSession(userID int, request *models.CreateSessionRequest) (string, error) {

	// language=sql
	SQL := `INSERT INTO sessions 
			(user_id, start_time,end_time, platform, model_name, os_version, device_id, token) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)	RETURNING token, id`

	args := []interface{}{
		userID,
		time.Now(),
		time.Now().Add(1 * time.Hour),
		request.Platform,
		request.ModelName,
		request.OSVersion,
		request.DeviceID,
		uuid.New(),
	}

	type sessionDetails struct {
		Token     string `db:"token"`
		SessionID int64  `db:"id"`
	}
	var session sessionDetails
	err := dh.DB.Get(&session, SQL, args...)
	if err != nil {
		logrus.Errorf("StartNewSession: error while starting new session: %v\n", err)
		return session.Token, err
	}

	return session.Token, nil
}

func (dh *DBHelper) GenerateAndStoreOTP(email string) (otp string, err error) {
	// Generate a random 6-digit OTP
	otp, err = generateRandomOTP(6)
	if err != nil {
		logrus.Errorf("GenerateAndStoreOTP: error generating OTP %v", err)
		return "", err
	}

	// Store the OTP in the database
	if err := dh.storeOTPInDatabase(email, otp); err != nil {
		logrus.Errorf("GenerateAndStoreOTP: error storing OTP in the database %v", err)
		return "", err
	}

	return otp, nil
}

func (dh *DBHelper) storeOTPInDatabase(email, otp string) error {
	// language=SQL
	SQL := `UPDATE users
            SET otp = $1
            WHERE email = $2
              AND deactivated IS FALSE
              AND archived_at IS NULL`

	_, err := dh.DB.Exec(SQL, otp, email)
	return err
}

func generateRandomOTP(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length should be a positive integer")
	}

	rand.Seed(time.Now().UnixNano())

	// Generate random bytes
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode random bytes to base64
	otp := base64.URLEncoding.EncodeToString(randomBytes)

	// Remove non-alphanumeric characters
	otp = removeNonAlphanumeric(otp)

	// Truncate or pad to the desired length
	otp = otp[:length]

	return otp, nil
}
func removeNonAlphanumeric(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, s)
}
