package auth_handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sme-backend/model"
	"sme-backend/src/core/config"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
	"sme-backend/src/core/middlewares"
	"sme-backend/src/enums/auth_types"
	"sme-backend/src/enums/user_types"
	"sme-backend/src/services/auth_service"
	file_upload_service "sme-backend/src/services/upload_service"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/utils"
)

func AdminSignIn(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	payload := &SignInDto{}

	if err := helpers.ParseBody(request, payload, true); err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Provide valid inputs", err)
		return
	}

	var auth model.Auth
	if err := db.Where("phone_number = ?", payload.PhoneNumber).First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err_msg := "user does not exists"
			helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
			return
		}
		helpers.HandleError(response, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	if auth.UserType != user_types.ADMIN {
		err_msg := "You are not authenticated to access this platform"
		helpers.HandleError(response, http.StatusUnauthorized, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	} else {
		otp, err := auth_service.GenerateOTPAndSend(auth.PhoneNumber)
		if err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
			return
		}

		if err := auth_service.CreatePhoneVerification(db, auth.PhoneNumber, otp); err != nil {
			err_msg := fmt.Sprintf("Unable to create a %s", auth.UserType)
			helpers.HandleError(response, http.StatusInternalServerError, err_msg, err)
			return
		}
	}

	success_msg := "OTP has been sent"
	output := map[string]any{"message": success_msg}
	helpers.HandleSuccess(response, http.StatusCreated, success_msg, output)
}

func SendOTP(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	_type := helpers.GetQueryParameter(request, "type")
	payload := &SignInDto{}

	if !auth_types.ContainsAuthType(_type) {
		err_msg := "Please provide a valid auth type"
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	}

	if err := helpers.ParseBody(request, payload, true); err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Provide valid inputs", err)
		return
	}

	var auth model.Auth
	if _type == auth_types.SIGNIN {
		if err := db.Where("phone_number = ?", payload.PhoneNumber).First(&auth).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err_msg := "user does not exists"
				helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
				return
			}
		}
	}

	if auth.ID != "" && auth.VerifiedAt == nil && auth.UserType == user_types.ENTREPRENEUR {
		err_msg := "Account verification in progress."
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	} else {
		testing_numbers := strings.Split(config.Config("TESTING_NUMBERS"), ",")
		if utils.Contains(testing_numbers, payload.PhoneNumber) {
			testing_otp := config.Config("TESTING_OTP")
			if err := auth_service.CreatePhoneVerification(db, payload.PhoneNumber, testing_otp); err != nil {
				err_msg := fmt.Sprintf("Unable to create a %s", auth.UserType)
				helpers.HandleError(response, http.StatusInternalServerError, err_msg, err)
				return
			}
		} else {
			otp, err := auth_service.GenerateOTPAndSend(payload.PhoneNumber)
			if err != nil {
				helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
				return
			}

			if err := auth_service.CreatePhoneVerification(db, payload.PhoneNumber, otp); err != nil {
				err_msg := fmt.Sprintf("Unable to create a %s", auth.UserType)
				helpers.HandleError(response, http.StatusInternalServerError, err_msg, err)
				return
			}
		}

	}

	success_msg := "OTP has been sent"
	output := map[string]any{"message": success_msg}
	helpers.HandleSuccess(response, http.StatusCreated, success_msg, output)
}

func SignUp(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	payload := &SignUpDto{}

	if err := helpers.ParseBody(request, payload, true); err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Provide valid inputs", err)
		return
	}

	if !user_types.ContainsUserType(payload.UserType) {
		err_msg := "Please provide a valid user type"
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	}

	if payload.UserType == user_types.ENTREPRENEUR {
		if payload.Email == nil || payload.Expertises == nil || payload.TotalWorkExperience == nil {
			err_msg := "Please provide a valid informations"
			helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
			return
		}
	}

	tx := db.Begin()
	if tx.Error != nil {
		err_msg := "Something went wrong. Please try again"
		helpers.HandleError(response, http.StatusInternalServerError, err_msg, tx.Error)
		return
	}

	if err := auth_service.CheckUserExistance(tx, payload.PhoneNumber); err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusBadRequest, err.Error(), err)
		return
	}

	auth := model.Auth{PhoneNumber: payload.PhoneNumber, UserType: payload.UserType}
	if payload.UserType == user_types.USER {
		now := time.Now()
		auth.VerifiedAt = &now
	}

	if err := tx.Create(&auth).Error; err != nil {
		tx.Rollback()
		err_msg := fmt.Sprintf("Unable to create a %s", auth.UserType)
		helpers.HandleError(response, http.StatusInternalServerError, err_msg, err)
		return
	}

	user := model.User{
		ID:                  auth.ID,
		Name:                payload.Name,
		Email:               payload.Email,
		Expertises:          payload.Expertises,
		TotalWorkExperience: payload.TotalWorkExperience,
		Documents:           payload.Documents,
		AadharNumber:        payload.AadharNumber,
		PanNumber:           payload.PanNumber,
	}

	if payload.UserType == user_types.USER {
		user.Verified = true
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		err_msg := fmt.Sprintf("Unable to create a %s", auth.UserType)
		helpers.HandleError(response, http.StatusInternalServerError, err_msg, err)
		return
	}

	var jwt_token *string
	if payload.UserType == user_types.USER && auth.ID != "" && auth.VerifiedAt != nil {
		token, err := auth_service.IssueJwtToken(auth.ID, auth.PhoneNumber, auth.UserType)
		if err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
			return
		}
		jwt_token = &token
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusInternalServerError, err.Error(), err)
		return
	}

	success_msg := "Account created successfully"
	if payload.UserType == user_types.ENTREPRENEUR {
		success_msg = "Account created and sent for verification."
	}

	output := map[string]any{"id": auth.ID, "token": jwt_token, "message": success_msg}
	helpers.HandleSuccess(response, http.StatusCreated, success_msg, output)
}

func VerifyOTP(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	payload := &VerifyOtpDto{}

	tx := db.Begin()
	if tx.Error != nil {
		err_msg := "Something went wrong. Please try again"
		helpers.HandleError(response, http.StatusInternalServerError, err_msg, tx.Error)
		return
	}

	defer tx.Rollback()

	if err := helpers.ParseBody(request, payload, true); err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Provide valid inputs", err)
		return
	}

	message, err := auth_service.VerifyOTP(tx, payload.PhoneNumber, payload.Otp)
	if err != nil {
		helpers.HandleError(response, http.StatusBadRequest, err.Error(), err)
		return
	}

	var auth model.Auth
	if err := tx.Where("phone_number = ?", payload.PhoneNumber).First(&auth).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
			return
		}
	}

	var jwt_token *string
	if auth.ID != "" && auth.VerifiedAt != nil {
		token, err := auth_service.IssueJwtToken(auth.ID, auth.PhoneNumber, auth.UserType)
		if err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
			return
		}
		jwt_token = &token
	}

	if err := tx.Commit().Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, err.Error(), err)
		return
	}

	result := map[string]any{"message": message, "token": jwt_token}
	helpers.HandleSuccess(response, http.StatusOK, message, result)
}

func UploadVerificationDocs(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_id := helpers.GetUrlParam(request, "user_id")

	storage_path := fmt.Sprintf("%s/verification-docs", user_id)
	uploaded_files, err := file_upload_service.UploadFiles(response, request, storage_path)

	err_msg := "Unable to upload verification docs. Contact support team"

	if err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, err_msg, err)
		return
	}

	docs_url := []string{}
	for _, file := range uploaded_files {
		docs_url = append(docs_url, file.Url)
	}

	docs := pq.StringArray(docs_url)
	user := &model.User{Documents: &docs}
	if err := db.Where("id = ?", user_id).Updates(user).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to upload verification docs. Contact support team", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Verification docs uploaded", uploaded_files)
}

func ResendOtp(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	payload := &SignInDto{}

	if err := helpers.ParseBody(request, payload, true); err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Provide valid inputs", err)
		return
	}

	var auth model.Auth
	if err := db.Where("phone_number = ?", payload.PhoneNumber).First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err_msg := "user does not exists"
			helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
			return
		}
	}

	otp, err := auth_service.GenerateOTPAndSend(auth.PhoneNumber)

	if err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if err := auth_service.CreatePhoneVerification(db, auth.PhoneNumber, otp); err != nil {
		err_msg := fmt.Sprintf("Unable to create a %s", auth.UserType)
		helpers.HandleError(response, http.StatusInternalServerError, err_msg, err)
		return
	}

	success_msg := "OTP has been sent"
	output := map[string]any{"message": success_msg}
	helpers.HandleSuccess(response, http.StatusCreated, success_msg, output)
}

func SignInByPass(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	payload := &SignInDto{}

	if err := helpers.ParseBody(request, payload, true); err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Provide valid inputs", err)
		return
	}

	var auth model.Auth
	if err := db.Where("phone_number = ?", payload.PhoneNumber).First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("user does not exists")
			helpers.HandleError(response, http.StatusNotFound, err.Error(), err)
			return
		}

		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	token, err := auth_service.IssueJwtToken(auth.ID, auth.PhoneNumber, auth.UserType)
	if err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	result := map[string]any{"token": token}
	helpers.HandleSuccess(response, http.StatusOK, "Signin success", result)
}

func VerifyJwtTokenExpiration(response http.ResponseWriter, request *http.Request) {
	token := helpers.GetQueryParameter(request, "token")

	if token == "" {
		err_msg := "Invalid JWT token"
		helpers.HandleError(response, http.StatusUnauthorized, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	}

	user_id, err := middlewares.VerifyJwtTokenExpiration(request, token)
	if err != nil {
		helpers.HandleError(response, http.StatusUnauthorized, "Invalid JWT token", err)
		return
	}

	result := map[string]any{"token": token, "user_id": user_id}
	helpers.HandleSuccess(response, http.StatusOK, "JWT verified", result)
}
