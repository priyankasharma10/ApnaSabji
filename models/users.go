package models

import (
	"time"

	"github.com/unidoc/timestamp"
	"github.com/volatiletech/null"
)

type UserInfo struct {
	UserID       int    `json:"userId" db:"id"`
	Fullname     string `json:"fullname" db:"fullname"`
	Email        string `json:"email" db:"email"`
	Mobilenumber string `json:"mobilenumber" db:"mobilenumber"`
}

type FetchUserData struct {
	UserID       int    `json:"userId" db:"id"`
	Fullname     string `json:"name" db:"fullname"`
	Email        string `json:"email" db:"email"`
	Mobilenumber string `json:"phone" db:"mobilenumber"`
}

type UserContextData struct {
	UserID       int    `json:"userId" db:"id"`
	SessionID    string `json:"sessionID" db:"token"`
	Fullname     string `json:"name" db:"fullname"`
	Email        string `json:"email" db:"email"`
	Mobilenumber string `json:"phone" db:"mobilenumber"`
}

type FetchUserSessionsData struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"userId" db:"user_id"`
	UUIDToken string    `json:"UUIDToken" db:"token"`
	EndTime   time.Time `json:"endTime" db:"end_time"`
}

type UserData struct {
	UserID int `json:"userId" db:"id"`
}

type CreateNewUserRequest struct {
	Fullname       string              `json:"name" db:"name"`
	Email          null.String         `json:"email" db:"email"`
	Mobilenumber   null.String         `json:"mobilenumber" db:"mobilenumber"`
	CreatedAt      time.Time           `db:"created_at"`
	UpdatedAt      timestamp.Timestamp `db:"updated_at"`
	OTP            int                 `json:"OTP"`
	OTPExpiryTime  int                 `db:"OTPExpiry"`
	UpdatedbyAdmin bool                `db:"UpdatedbyAdmin"`
	Deleted        bool                `db:"Deleted"`
	DeletedbyAdmin bool                `db:"DeletedbyAdmin"`
	CheckActive    bool                `db:"CheckActive"`
}

type GetUserDataByEmail struct {
	UserID       int    `json:"userId" db:"id"`
	Fullname     string `json:"name" db:"name"`
	Email        string `json:"email" db:"email"`
	Mobilenumber string `json:"phone" db:"phone"`
}

type EmailAndOTP struct {
	OTP   string `json:"otp"`
	Email string `json:"email"`
}

type CreateSessionRequest struct {
	Platform  string `json:"platform"`
	ModelName string `json:"modelName"`
	OSVersion string `json:"osVersion"`
	DeviceID  string `json:"deviceId"`
}

type AuthLoginRequest struct {
	Platform  string      `json:"platform"`
	ModelName null.String `json:"modelName"`
	OSVersion null.String `json:"osVersion"`
	DeviceID  null.String `json:"deviceId"`
	Email     string      `json:"email"`
	OTP       string      `json:"otp"`
}

type GenerateAndStoreOTP struct {
	Email        string `json:"email" db:"email"`
	Mobilenumber string `json:"phone" db:"phone"`
	OTP          string `json:"otp"`
}
