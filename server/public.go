package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/ttacon/libphonenumber"
	"github.com/vijaygniit/ApnaSabji/models"
	"github.com/vijaygniit/ApnaSabji/providers/authProvider"
	"github.com/vijaygniit/ApnaSabji/scmerrors"
	"github.com/vijaygniit/ApnaSabji/utils"
	"github.com/volatiletech/null"
)

func (srv *Server) register(resp http.ResponseWriter, req *http.Request) {
	// Log the start of the registration process
	log.Println("Registration process started")

	var newUserReq models.CreateNewUserRequest
	uc := srv.MiddlewareProvider.UserFromContext(req.Context())

	// Log the request details
	log.Printf("Received registration request: %+v\n", req)

	if err := json.NewDecoder(req.Body).Decode(&newUserReq); err != nil {
		// Log the error and details when decoding fails
		log.Printf("Error decoding request body: %v\n", err)
		scmerrors.RespondClientErr(resp, err, http.StatusBadRequest, "Error creating user", "Error parsing request")
		return
	}

	// Check if the email is empty
	if newUserReq.Email.String == "" {
		log.Println("Email is empty")
		scmerrors.RespondClientErr(resp, errors.New("email cannot be empty"), http.StatusBadRequest, "Email cannot be empty", "Email cannot be empty")
		return
	}

	// Trim and check if the name is empty
	name := strings.TrimSpace(newUserReq.Fullname)
	if name == "" {
		log.Println("Name is empty")
		scmerrors.RespondClientErr(resp, errors.New("name cannot be empty"), http.StatusBadRequest, "Name cannot be empty", "Name cannot be empty")
		return
	}

	// Check if the user already exists
	isUserExist, _, err := srv.DBHelper.IsUserAlreadyExists(newUserReq.Email.String)
	if err != nil {
		// Log the error when checking user existence
		log.Printf("Error checking user existence: %v\n", err)
		scmerrors.RespondGenericServerErr(resp, err, "Error in processing request")
		return
	}

	if isUserExist {
		// Log the error when a user already exists
		log.Println("User already exists with the provided email")
		scmerrors.RespondClientErr(resp, errors.New("error creating user"), http.StatusBadRequest, "This email is already linked with one of our accounts. Please use a different email address", "Unable to create a user with a duplicate email address")
		return
	}

	// Lowercase the email
	newUserReq.Email.String = strings.ToLower(newUserReq.Email.String)

	// Check if the name is empty again
	if newUserReq.Fullname == "" {
		log.Println("Name is empty")
		scmerrors.RespondClientErr(resp, errors.New("name cannot be empty"), http.StatusBadRequest, "Name cannot be empty", "Name cannot be empty")
		return
	}

	// Check if the mobile number is empty
	if newUserReq.Mobilenumber.IsZero() {
		log.Println("Phone number is empty")
		scmerrors.RespondClientErr(resp, errors.New("phone number cannot be empty"), http.StatusBadRequest, "Phone number cannot be empty", "Phone number cannot be empty")
		return
	}

	// Validate and format mobile number
	num, err := libphonenumber.Parse(newUserReq.Mobilenumber.String, "IN")
	if err != nil || !libphonenumber.IsValidNumber(num) {
		log.Println("Invalid phone number")
		scmerrors.RespondClientErr(resp, errors.New("invalid phone number"), http.StatusBadRequest, "Invalid phone number", "Invalid phone number")
		return
	}

	uncleanPhoneNumber := newUserReq.Mobilenumber.String

	// Extract the phone number from the format "+91 XXXXX"
	if strings.Count(uncleanPhoneNumber, "+") == 2 {
		uncleanPhoneNumber = uncleanPhoneNumber[strings.LastIndex(uncleanPhoneNumber, "+")+1:]
	}

	phone := strings.ReplaceAll(uncleanPhoneNumber, " ", "")

	num, err = libphonenumber.Parse(phone, "IN")
	if err != nil {
		log.Printf("Error parsing phone number: %v\n", err)
		scmerrors.RespondClientErr(resp, err, http.StatusBadRequest, "Phone Number not in a correct format", "Invalid format for phone number")
		return
	}

	isValidNumber := libphonenumber.IsValidNumber(num)

	if !isValidNumber {
		log.Println("Invalid phone number")
		scmerrors.RespondClientErr(resp, errors.New("invalid phone number"), http.StatusBadRequest, "Invalid phone number", "Invalid phone number")
		return
	}

	phoneNumber := libphonenumber.Format(num, libphonenumber.E164)
	newUserReq.Mobilenumber = null.StringFrom(phoneNumber)

	// Check if the mobile number already exists
	isMobileAlreadyExist, err := srv.DBHelper.IsPhoneNumberAlreadyExist(phoneNumber)
	if err != nil {
		// Log the error when checking mobile number existence
		log.Printf("Error checking mobile number existence: %v\n", err)
		scmerrors.RespondGenericServerErr(resp, err, "Unable to create user")
		return
	}

	if isMobileAlreadyExist {
		// Log the error when a mobile number already exists
		log.Println("Mobile number already exists")
		scmerrors.RespondClientErr(resp, errors.New("mobile number already exists"), http.StatusBadRequest, "This phone number is already linked with one of our accounts. Please use a different phone number", "Unable to create a user")
		return
	}

	// Creating user in the database
	userID, err := srv.DBHelper.CreateNewUser(&newUserReq, uc.UserID)
	if err != nil {
		// Log the error when creating a new user in the database
		log.Printf("Error creating new user in the database: %v\n", err)
		scmerrors.RespondGenericServerErr(resp, err, "Error registering new user")
		return
	}

	// Log the successful registration
	log.Printf("User registered successfully with ID: %v\n", userID)

	utils.EncodeJSONBody(resp, http.StatusCreated, map[string]interface{}{
		"message": "success",
		"userId":  userID,
	})
}

// Generate a Random OTP

func (srv *Server) generateAndStoreOTP(email string) (string, error) {
	otp, err := generateRandomOTP(6)
	if err != nil {
		logrus.Error("Error generating OTP: ", err)
		return "", err
	}

	// Store the OTP in the database
	err = srv.DBHelper.StoreOTP(email, otp)
	if err != nil {
		logrus.Error("Error storing OTP in the database: ", err)
		return "", err
	}

	return otp, nil
}

func generateRandomOTP(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length should be a positive integer")
	}

	// Generate random bytes
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode random bytes to base64
	otp := base64.URLEncoding.EncodeToString(randomBytes)

	// Remove non-alphanumeric characters
	otp = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r
		}
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, otp)

	// Truncate or pad to the desired length
	otp = otp[:length]

	return otp, nil
}

// LoginWithEmailOtp

func (srv *Server) loginWithEmailOTP(resp http.ResponseWriter, req *http.Request) {
	var token string
	var authLoginRequest models.AuthLoginRequest
	err := json.NewDecoder(req.Body).Decode(&authLoginRequest)
	if err != nil {
		logrus.Error("loginWithEmail/Mobilenumber and otp: unable to decode request body ", err)
		return
	}

	if authLoginRequest.OTP == "" {
		scmerrors.RespondClientErr(resp, errors.New("otp can not be empty"), http.StatusBadRequest, "Empty otp!", "otp field can not be empty")
		return
	}

	if authLoginRequest.Email == "" {
		scmerrors.RespondClientErr(resp, errors.New("email can not be empty"), http.StatusBadRequest, "Please enter email to login", "email can not be empty")
		return
	}

	// Log the received authentication request
	logrus.Infof("Received login request with Email: %s, OTP: %s", authLoginRequest.Email, authLoginRequest.OTP)

	UserDataByEmail, err := srv.DBHelper.GetUserInfoByEmail(authLoginRequest.Email)
	if err != nil {
		logrus.Error("Error getting user info by email: ", err)
		scmerrors.RespondClientErr(resp, err, http.StatusBadRequest, "error getting user info", "error getting user info")
		return
	}

	loginReq := models.EmailAndOTP{
		Email: authLoginRequest.Email,
		OTP:   authLoginRequest.OTP,
	}
	loginReq.Email = strings.ToLower(loginReq.Email)

	createUserSession := models.CreateSessionRequest{
		Platform:  authLoginRequest.Platform,
		ModelName: authLoginRequest.ModelName.String,
		OSVersion: authLoginRequest.OSVersion.String,
		DeviceID:  authLoginRequest.DeviceID.String,
	}

	userID, errorMessage, err := srv.DBHelper.LogInUserUsingEmail(loginReq)
	if err != nil {
		logrus.Error("Error logging in user with email: ", err)
		scmerrors.RespondClientErr(resp, err, http.StatusInternalServerError, errorMessage, errorMessage)
		return
	}

	UUIDToken, err := srv.DBHelper.StartNewSession(userID, &createUserSession)
	if err != nil {
		logrus.Error("Error creating session: ", err)
		scmerrors.RespondGenericServerErr(resp, err, "error in creating session")
		return
	}

	userInfo, err := srv.DBHelper.FetchUserData(userID)
	if err != nil {
		logrus.Error("Error getting user info: ", err)
		scmerrors.RespondGenericServerErr(resp, err, "error in getting user info")
		return
	}

	devClaims := make(map[string]interface{})
	devClaims["UUIDToken"] = UUIDToken
	devClaims["userInfo"] = UserDataByEmail
	devClaims["UserSession"] = createUserSession

	token, err = authProvider.GenerateJWT(devClaims)
	if err != nil {
		logrus.Error("Error generating JWT: ", err)
		scmerrors.RespondClientErr(resp, err, http.StatusInternalServerError, "error while login", "error while login")
		return
	}

	utils.EncodeJSONBody(resp, http.StatusOK, map[string]interface{}{
		"userInfo": userInfo,
		"token":    token,
	})
}
