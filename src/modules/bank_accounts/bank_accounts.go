package bank_accounts_handler

import (
	"errors"
	"net/http"
	"sme-backend/model"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"

	"gorm.io/gorm"
)

func GetBankAccount(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	reqestor_id := helpers.GetUserIdFromJwtToken(request)

	var account model.BankAccount
	if err := db.Where("created_by = ?", reqestor_id).First(&account).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			helpers.HandleError(response, http.StatusInternalServerError, "Couldn't get account", err)
			return
		}
	}

	helpers.HandleSuccess(response, http.StatusOK, "Bank account fetched", account)
}

func CreateUpdateBankAccount(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	reqestor_id := helpers.GetUserIdFromJwtToken(request)

	var account model.BankAccount
	if err := helpers.ParseBody(request, &account, false); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to parse the body", err)
		return
	}

	account.CreatedBy = reqestor_id
	if account.ID != "" {
		if err := db.Where("id = ?", account.ID).Updates(&account).Error; err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Couldn't update account", err)
			return
		}
	} else {
		if err := db.Create(&account).Error; err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Couldn't create account", err)
			return
		}
	}

	helpers.HandleSuccess(response, http.StatusCreated, "Bank account saved", account)
}
