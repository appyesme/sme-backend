package commissions_handler

import (
	"fmt"
	"net/http"
	"sme-backend/model"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
)

func GetCommissonsDetails(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	var commisson model.Commission

	if err := db.Find(&commisson).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Commisson fetched", commisson)
}

func UpdateCommissonsDetails(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	requestor_id := helpers.GetUserIdFromJwtToken(request)

	var commisson model.Commission

	if err := helpers.ParseBody(request, &commisson, false); err != nil {
		fmt.Println("Err 0", err)
		helpers.HandleError(response, http.StatusBadRequest, "Provide valid inputs", err)
		return
	}

	// START TRANSACTION
	tx := db.Begin()
	if err := tx.Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	defer tx.Rollback()

	if err := db.Exec("DELETE FROM commissions").Error; err != nil {
		fmt.Println("Err 1", err)
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	commisson.CreatedBy = requestor_id
	if err := tx.Create(&commisson).Error; err != nil {
		fmt.Println("Err 2", err)

		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	// END TRANSACTION

	helpers.HandleSuccess(response, http.StatusOK, "Commisson updated", commisson)
}
