package auth_service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sme-backend/model"
	"sme-backend/src/core/config"
	"sme-backend/src/core/helpers"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func CheckUserExistance(tx *gorm.DB, phone_number string) error {
	var auth model.Auth
	if err := tx.Where("phone_number = ?", phone_number).First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		return err
	}

	return errors.New("user already exists")
}

func VerifyOTP(tx *gorm.DB, phoneNumber string, otp string) (string, error) {
	var message string
	var phoneVerification model.PhoneVerification
	// Find the OTP record for the provided phone number and validate the expiration
	if err := tx.Where("phone_number = ? AND otp_code = ?", phoneNumber, otp).First(&phoneVerification).Error; err != nil {
		return message, errors.New("invalid otp")
	}

	// Check if OTP has expired
	if time.Now().After(phoneVerification.OtpExpiresAt) {
		return message, errors.New("otp has expired")
	}

	if err := tx.Where("id = ?", phoneVerification.ID).Delete(model.PhoneVerification{}).Error; err != nil {
		return message, err
	}

	message = "OTP verified successfully"
	return message, nil
}

func IssueJwtToken(sub string, phone_number string, user_type string) (string, error) {
	claims := jwt.MapClaims{}
	claims["sub"] = sub
	claims["phone_number"] = phone_number
	claims["user_type"] = user_type
	claims["iat"] = time.Now().Unix()
	claims["exp"] = helpers.GetCurrentTime().Add(time.Hour * 24 * 30).Unix() // Token is valid for 30 days
	tokenClaimer := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := config.Config("JWT_SECRET")
	token, err := tokenClaimer.SignedString([]byte(jwtSecret))
	return token, err
}

func CreatePhoneVerification(db *gorm.DB, phone_number string, otp string) error {
	phoneVerification := model.PhoneVerification{
		PhoneNumber:  phone_number,
		OtpCode:      otp,
		OtpExpiresAt: time.Now().Add(5 * time.Minute),
	}

	var id string
	if err := db.Model(model.PhoneVerification{}).Select("id").First(&id, "phone_number = ?", phone_number).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	phoneVerification.ID = id
	return db.Where("phone_number = ?", phone_number).Save(&phoneVerification).Error
}

func GenerateOTPAndSend(phone_number string) (string, error) {
	otp, err := generateOTP()
	if err != nil {
		return otp, err
	}

	message := fmt.Sprintf("Your S - ME App Verification Code is %s. Don't share it with anyone.", otp)

	// Prepare the data for URL encoding
	data := url.Values{}
	data.Set("user", config.Config("PHONE_AUTH_USER"))
	data.Set("key", config.Config("PHONE_AUTH_KEY"))
	data.Set("senderid", config.Config("PHONE_AUTH_SENDER_ID"))
	data.Set("accusage", "1")
	data.Set("mobile", phone_number)
	data.Set("message", message)

	// URL encode the data
	apiURL := config.Config("PHONE_AUTH_API_URL") + data.Encode()

	// Create the HTTP POST request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return otp, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return otp, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return otp, errors.New("something went wrong")
	}

	return otp, nil
}

func generateOTP() (string, error) {
	number, err := rand.Int(rand.Reader, big.NewInt(10000)) // Range: 0-9999
	if err != nil {
		return "", err
	}
	otp := fmt.Sprintf("%04d", number.Int64()) // Zero-padded to 4 digits
	return otp, nil
}
