package admins_handler

import (
	"errors"
	"fmt"
	"net/http"
	"sme-backend/model"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
	"sme-backend/src/enums/appointment_status"
	"sme-backend/src/enums/payment_status"
	"sme-backend/src/enums/user_types"
	"strings"
	"time"

	"gorm.io/gorm"
)

func GetPlatformUsers(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	users := make([]UserDetails, 0)

	page := helpers.GetQueryPageParam(request)
	limit := helpers.GetQueryLimitParam(request)
	search_query := helpers.GetQueryParameter(request, "search_query")

	var query_conditions []string
	var query_conditions_vars []any

	query := `SELECT a.phone_number, a.user_type, a.verified_at, u.* FROM auth a LEFT JOIN users u ON u.id = a.id`

	// WHERE conditions - START
	// return only verified users.
	query_conditions = append(query_conditions, "u.verified = true")

	if search_query != "" {
		query_conditions = append(query_conditions, "(u.name ILIKE ? OR u.email ILIKE ?)")
		search_term := fmt.Sprintf("%%%s%%", search_query)
		query_conditions_vars = append(query_conditions_vars, search_term, search_term)
	}

	// WHERE conditions - END

	query = query + " WHERE " + strings.Join(query_conditions, " AND ")

	if err := db.Raw(query, query_conditions_vars...).Limit(limit).Offset(page * limit).Scan(&users).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to get platform users", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Platform users fecthed", users)
}

func GetNewEntrepreneursJoiningRequests(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	users_detail := make([]UserDetails, 0)

	query := `SELECT * FROM auth a LEFT JOIN users u ON u.id = a.id WHERE a.user_type = ? AND NOT u.verified`
	if err := db.Raw(query, user_types.ENTREPRENEUR).Scan(&users_detail).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to get entrepreneurs joining requests", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Entrepreneurs joining requests fetched", users_detail)
}

func ApproveOrRejectEntrepreneur(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)

	joining_requestor_id := helpers.GetUrlParam(request, "joining_requestor_id")
	approval_status := helpers.GetQueryBoolParam(request, "approval_status")

	if approval_status == nil {
		err := errors.New("application approval status not valid")
		helpers.HandleError(response, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// START TRANSACTION
	tx := db.Begin()
	if err := tx.Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	defer tx.Rollback()

	var auth model.Auth
	if err := tx.Where("id = ?", joining_requestor_id).First(&auth).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	current_time := helpers.GetCurrentTime()
	auth.VerifiedAt = &current_time
	if err := tx.Where("id = ?", joining_requestor_id).Updates(&auth).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	user_verified := map[string]any{"verified": *approval_status}
	if err := tx.Model(model.User{}).Where("id = ?", joining_requestor_id).Updates(&user_verified).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	// END TRANSACTION

	helpers.HandleSuccess(response, http.StatusOK, "Account approved", auth)
}

func GetPaymentClearanceEntrepreneurs(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)

	query := `SELECT a.id, a.phone_number, u.name, MAX(lpc.updated_at) AS last_cleared_at, to_json(ba.*) AS bank_account
			FROM
			users u
			JOIN auth a ON a.id = u.id
			LEFT JOIN bank_accounts ba ON ba.created_by = a.id
			LEFT JOIN last_payment_cleared lpc ON lpc.entrepreneur_id = a.id  AND lpc.updated_at < current_date - interval '5 days'
			WHERE a.user_type = ? AND u.verified = TRUE
			GROUP BY a.id, ba.id, u.name ORDER BY last_cleared_at ASC;`

	var data []PaymentClearanceEntrepreneursDTO
	if err := db.Raw(query, user_types.ENTREPRENEUR).Scan(&data).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Account approved", data)
}

func GetClearingPaymentDetails(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	entrepreneur_id := helpers.GetQueryParameter(request, "entrepreneur_id")

	var service_ids []any
	if err := db.Model(&model.Service{}).Where("created_by = ?", entrepreneur_id).Select("id").Find(&service_ids).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	var payments []model.Payment

	query := `SELECT p.* FROM payments p
			LEFT JOIN appointments a ON a.id = p.appointment_id AND a.status = ?
			WHERE p.service_id IN (?) AND p.status = ? AND p.created_at < (CURRENT_DATE - INTERVAL '5 days');`

	if err := db.Raw(query, appointment_status.COMPLETED, service_ids, payment_status.PAID).Scan(&payments).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	var total_amount_to_clear float64
	var payment_ids []string
	if len(payments) > 0 {
		for _, payment := range payments {
			total_amount_to_clear = total_amount_to_clear + payment.Amount
			payment_ids = append(payment_ids, payment.ID)
		}
	}

	total_amount_to_clear = total_amount_to_clear / 100 // Convert to rupee amount

	result := map[string]any{
		"total_amount_to_clear": total_amount_to_clear,
		"payment_ids":           payment_ids,
	}

	helpers.HandleSuccess(response, http.StatusOK, "Account approved", result)
}

func MarkAsPaymentCleared(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	entrepreneur_id := helpers.GetQueryParameter(request, "entrepreneur_id")

	var payload_ids []any
	if err := helpers.ParseBody(request, &payload_ids, false); err != nil {
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

	if err := tx.Model(&model.Payment{}).Where("id IN (?)", payload_ids).Update("status", payment_status.CLEARED).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	var lastPaymentCleared model.LastPaymentCleared
	if err := tx.Where("entrepreneur_id = ?", entrepreneur_id).First(&lastPaymentCleared).Error; err != nil {
		if err == gorm.ErrRecordNotFound {

			lastPaymentCleared.CreatedBy = helpers.GetUserIdFromJwtToken(request)
			lastPaymentCleared.EntrepreneurID = entrepreneur_id

			if err := tx.Create(&lastPaymentCleared).Error; err != nil {
				helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
				return
			}
		} else {
			helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
			return
		}
	}

	if err := tx.Where("id = ? AND entrepreneur_id = ?", lastPaymentCleared.ID, entrepreneur_id).Updates(&lastPaymentCleared).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	// END TRANSACTION

	helpers.HandleSuccess(response, http.StatusOK, "Payment cleared", nil)
}

func GetWeeklyJoinedUsersCount(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)

	start_date := helpers.GetQueryParameter(request, "start_date")
	end_date := helpers.GetQueryParameter(request, "end_date")

	type DailyCount struct {
		Day   time.Time `json:"day"`
		Count int       `json:"count"`
	}

	var dailyCounts []DailyCount

	query := `SELECT DATE(verified_at) AS day, COUNT(*) AS count from auth a
				WHERE DATE(a.verified_at) >= DATE(?) AND DATE(a.verified_at) <= DATE(?)
				GROUP BY DATE(a.verified_at) ORDER BY DATE(a.verified_at)`

	if err := db.Raw(query, start_date, end_date).Scan(&dailyCounts).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Daily user join count", dailyCounts)
}

func GetDailyWiseEarnings(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	start_date := helpers.GetQueryParameter(request, "start_date")
	end_date := helpers.GetQueryParameter(request, "end_date")

	type DayWiseEarnings struct {
		Day      time.Time `json:"day"`
		Earnings float64   `json:"earnings"`
	}

	var earnings []DayWiseEarnings

	query := `SELECT DATE(updated_at) AS day, (SUM(p.amount) / 100) AS earnings from payments p
				WHERE p.status = ? AND DATE(p.updated_at) >= DATE(?) AND DATE(p.updated_at) <= DATE(?)
				GROUP BY DATE(p.updated_at) ORDER BY DATE(p.updated_at)`

	if err := db.Raw(query, payment_status.PAID, start_date, end_date).Scan(&earnings).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Earnings fetched", earnings)
}
