package users_handler

import (
	"fmt"
	"net/http"

	"sme-backend/model"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
	"sme-backend/src/enums/user_types"
	file_upload_service "sme-backend/src/services/upload_service"

	"gorm.io/gorm/clause"
)

func GetUserAppointments(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_id := helpers.GetUserIdFromJwtToken(request)
	user_type := helpers.GetUserType(request)
	date := helpers.GetQueryParameter(request, "date")

	if date == "" {
		helpers.HandleError(response, http.StatusBadRequest, "Date is required", fmt.Errorf("missing date"))
		return
	}

	var appointments []GetUserAppointmentsDto
	var query_values []any
	query := `select
				ap.*,
				to_jsonb(s.*) as service,
				to_jsonb(p.*) as payment
				%s
			from appointments ap
				left join payments p on p.appointment_id = ap.id AND ap.status = 'INITIATED'
				left join services s on s.id = ap.service_id
				%s
			where %s ORDER BY ap.updated_at ASC`

	if user_type == user_types.ENTREPRENEUR {
		candidate_query_select := `, json_build_object('id', u.id, 'email', u.email, 'name', u.name, 'photo_url', u.photo_url, 'phone_number', a.phone_number) as candidate`
		candidate_query_joins := `left join users u on u.id = ap.created_by left join auth a on a.id = u.id`
		query = fmt.Sprintf(query, candidate_query_select, candidate_query_joins, "ap.appointment_date = ? AND s.created_by = ?")
		query_values = append(query_values, date, user_id)
	} else {
		query = fmt.Sprintf(query, "", "", "ap.appointment_date = ? AND ap.created_by = ?")
		query_values = append(query_values, date, user_id)
	}

	if err := db.Raw(query, query_values...).Scan(&appointments).Error; err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Unable to fetch appointments", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Appointments fetched", appointments)
}

func GetEntrepreneurAppointments(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_id := helpers.GetUserIdFromJwtToken(request)
	date := helpers.GetQueryParameter(request, "date")

	var appointments []GetUserAppointmentsDto
	query := `select
					ap.*,
					to_jsonb(s.*) as service,
					to_jsonb(p.*) as payment
				from
					appointments ap
					join services s on s.id = ap.service_id
					left join payments p on p.appointment_id = ap.id AND ap.status = 'INITIATED'
				where ap.date = ? AND s.created_by = ?`

	if err := db.Raw(query, date, user_id).Scan(&appointments).Error; err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Unable to fetch appointments", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Appointments fetched", appointments)
}

func GetUserDetailsByUserId(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	profile_id := helpers.GetUrlParam(request, "profile_id")
	user_id := helpers.GetUserIdFromJwtToken(request)

	var user_details UserDetailsDto

	var condition_vars []any
	query := `select
				u.id,
				a.phone_number,
				a.user_type,
				u.name,
				u.email,
				u.photo_url,
				u.total_work_experience,
				u.about,
				%s
				to_jsonb(u.expertises) as expertises
			from users u join auth a on a.id = u.id
			where u.id = ? group by u.id, a.phone_number, a.user_type`

	if user_id != "" {
		query = fmt.Sprintf(query, "EXISTS ( SELECT 1 FROM favourite_users fu WHERE fu.created_by = ? AND fu.profile_id = ?) AS favourited,")
		condition_vars = append(condition_vars, user_id, profile_id)
	} else {
		query = fmt.Sprintf(query, "")
	}

	condition_vars = append(condition_vars, profile_id)
	if err := db.Raw(query, condition_vars...).Scan(&user_details).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable fetch user details", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "User details fetched", user_details)
}

func UpdateUserDetails(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	var user_details_update UpdateUserDetailsDto
	if err := helpers.ParseBody(request, &user_details_update, false); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to parse the body", err)
		return
	}

	user_id := helpers.GetUserIdFromJwtToken(request)
	user := model.User{
		Name:                *user_details_update.Name,
		Email:               user_details_update.Email,
		TotalWorkExperience: user_details_update.TotalWorkExperience,
		Expertises:          user_details_update.Expertises,
		About:               user_details_update.About,
	}

	// Update user details
	if err := db.Where("id = ?", user_id).Updates(&user).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Couldn't update user details", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "User details updated", user_details_update)
}

func UploadPhoto(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_id := helpers.GetUserIdFromJwtToken(request)

	var user model.User
	if err := db.Where("id = ?", user_id).First(&user).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to upload photo", err)
		return
	}

	if user.PhotoURL != nil && *user.PhotoURL != "" {
		storage_path := file_upload_service.GetStoragePathByURL(*user.PhotoURL)
		if err := file_upload_service.DeleteFileFromStorage([]string{storage_path}); err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Unable to upload photo", err)
			return
		}
	}

	storage_path := fmt.Sprintf("%s/photos/", user_id)
	uploaded_files, err := file_upload_service.UploadFiles(response, request, storage_path)
	if err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to upload photo", err)
		return
	}

	update := model.User{PhotoURL: &uploaded_files[0].Url}
	if err := db.Where("id = ?", user_id).Updates(&update).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to upload photo", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusCreated, "Photo uploaded", uploaded_files[0].Url)
}

func GetFavouriteUsers(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_id := helpers.GetUserIdFromJwtToken(request)

	var data []map[string]any
	query := `SELECT u.id, u.photo_url, u.name FROM favourite_users fu JOIN users u on u.id = fu.profile_id WHERE fu.created_by = ?`
	if err := db.Raw(query, user_id).Scan(&data).Error; err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Unable to add to favourite", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusCreated, "Added to favourite", data)
}

func AddUserToFavourite(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	profile_id := helpers.GetUrlParam(request, "profile_id")
	user_id := helpers.GetUserIdFromJwtToken(request)

	favourite := model.FavouriteUser{CreatedBy: user_id, ProfileID: profile_id}
	if err := db.Create(&favourite).Error; err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Unable to add to favourite", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusCreated, "Added to favourite", favourite)
}

func RemoveUserFromFavourite(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	profile_id := helpers.GetUrlParam(request, "profile_id")
	user_id := helpers.GetUserIdFromJwtToken(request)

	var favourite model.FavouriteUser
	if err := db.Where("profile_id = ? AND created_by = ?", profile_id, user_id).Clauses(clause.Returning{}).Delete(&favourite).Error; err != nil {
		helpers.HandleError(response, http.StatusBadRequest, "Unable to remove from favourite", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Removed from favourite", favourite)
}
