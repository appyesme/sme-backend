package services_handler

import (
	"errors"
	"fmt"
	"net/http"
	"sme-backend/model"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
	file_upload_service "sme-backend/src/services/upload_service"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetServices(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)

	profile_id := helpers.GetQueryParameter(request, "profile_id")

	page, err_page := strconv.Atoi(helpers.GetQueryParameter(request, "page"))
	limit, err_limit := strconv.Atoi(helpers.GetQueryParameter(request, "limit"))

	var err error
	var msg string

	// Set default values for page and limit if they are not valid numbers
	if err_page != nil {
		if errors.Is(err_page, strconv.ErrSyntax) {
			page = 1
		} else {
			msg = "Query page parameter is malformed"
			err = errors.New(strings.ToLower(msg))
		}
	}

	if err_limit != nil {
		if errors.Is(err_limit, strconv.ErrSyntax) {
			limit = 15
		} else {
			msg = "Query limit parameter is malformed"
			err = errors.New(strings.ToLower(msg))
		}
	}

	if err != nil {
		helpers.HandleError(response, http.StatusBadRequest, msg, err)
		return
	}

	services := make([]GetServicesDto, 0)

	var query_params []any
	query := `SELECT
				s.*,
				json_build_object('id', u.id, 'name', u.name, 'photo_url', u.photo_url) AS author,
				COALESCE(json_agg(DISTINCT sm.*) FILTER (WHERE sm.id IS NOT NULL), '[]') AS medias
			FROM
				services s
				JOIN users u ON u.id = s.created_by
				LEFT JOIN service_medias AS sm ON sm.service_id = s.id
			WHERE s.status = 'PUBLISHED' %s
			GROUP BY s.id, u.id, u.photo_url
			ORDER BY s.title
			LIMIT ? OFFSET ?;`

	if profile_id != "" {
		query = fmt.Sprintf(query, " AND s.created_by = ?")
		query_params = append(query_params, profile_id)
	} else {
		query = fmt.Sprintf(query, "")
	}

	query_params = append(query_params, []any{limit, (page - 1) * limit}...)

	// Run the query using db.Raw with the query and values
	if err := db.Raw(query, query_params...).Scan(&services).Error; err != nil {
		helpers.HandleError(response, http.StatusNotFound, "Unable to get services", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Services fetched", services)
}

func GetServiceByID(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	service_id := helpers.GetUrlParam(request, "service_id")

	var service GetServicesDto
	query := `SELECT
				s.*,
				json_build_object('id', u.id, 'name', u.name, 'photo_url', u.photo_url) AS author,
				COALESCE(json_agg(DISTINCT sm.*) FILTER (WHERE sm.id IS NOT NULL), '[]') AS medias
			FROM
				services s
				JOIN users u ON u.id = s.created_by
				LEFT JOIN service_medias AS sm ON sm.service_id = s.id
			WHERE s.id = ? GROUP BY s.id, u.id, u.photo_url`

	// Run the query using db.Raw with the query and values
	if err := db.Raw(query, service_id).Scan(&service).Error; err != nil {
		helpers.HandleError(response, http.StatusNotFound, "Unable to get service", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Service fetched", service)
}

func GetServiceDays(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	service_id := helpers.GetUrlParam(request, "service_id")
	var services_days []model.ServiceDay

	if err := db.Where("service_id = ?", service_id).Find(&services_days).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to get service days", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Service days fetched", services_days)
}

func ChangeServiceDayEnableStatus(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)

	service_day_id := chi.URLParam(request, "service_day_id")

	var service_day model.ServiceDay
	if err := helpers.ParseBody(request, &service_day, false); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to parse the body", err)
		return
	}

	if err := db.Clauses(clause.Returning{}).
		Where("id = ?", service_day_id).
		Select("Enabled").
		Updates(&service_day).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Couldn't update service day status", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Service day status updated", service_day)
}

func CreateService(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)

	var service model.Service
	if err := helpers.ParseBody(request, &service, false); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to parse the body", err)
		return
	}

	service.CreatedBy = helpers.GetUserIdFromJwtToken(request)

	tx := db.Begin()
	if tx.Error != nil {
		err := errors.New("something went wrong. please try again")
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a service", err)
		return
	}

	if err := tx.Create(&service).Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusInternalServerError, "Couldn't create service", err)
		return
	}

	service_days := []model.ServiceDay{
		{CreatedBy: service.CreatedBy, Day: 1, ServiceID: service.ID},
		{CreatedBy: service.CreatedBy, Day: 2, ServiceID: service.ID},
		{CreatedBy: service.CreatedBy, Day: 3, ServiceID: service.ID},
		{CreatedBy: service.CreatedBy, Day: 4, ServiceID: service.ID},
		{CreatedBy: service.CreatedBy, Day: 5, ServiceID: service.ID},
		{CreatedBy: service.CreatedBy, Day: 6, ServiceID: service.ID},
		{CreatedBy: service.CreatedBy, Day: 7, ServiceID: service.ID},
	}

	if err := tx.Create(&service_days).Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusInternalServerError, "Couldn't create service", err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a service", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusCreated, "Service created", service)
}

func DeleteService(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)

	service_id := helpers.GetUrlParam(request, "service_id")

	tx := db.Begin()
	if tx.Error != nil {
		err := errors.New("something went wrong. please try again")
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a service", err)
		return
	}

	var service model.Service
	var storage_paths []string

	if err := tx.Model(model.ServiceMedia{}).Where("service_id = ?", service_id).Select("storage_path").Find(&storage_paths).Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a service", err)
		return
	}

	if len(storage_paths) > 0 {
		if err := file_upload_service.DeleteFileFromStorage(storage_paths); err != nil {
			tx.Rollback()
			helpers.HandleError(response, http.StatusConflict, "Unable to delete a service", err)
			return
		}
	}

	if err := tx.Clauses(clause.Returning{}).Delete(&service, "id = ?", service_id).Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a service", err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a service", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Service deleted", service)
}

func UpdateService(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)

	service_id := chi.URLParam(request, "service_id")

	service := model.Service{}
	if err := helpers.ParseBody(request, &service, false); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to parse the body", err)
		return
	}

	if err := db.Where("id = ?", service_id).
		Select("home_available", "title", "expertises", "charge", "additional_charge", "description", "address", "status", "salon_available").
		Updates(&service).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Couldn't update service", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Service updated", service)
}

func UploadServiceMedia(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	service_id := chi.URLParam(request, "service_id")
	user_id := helpers.GetUserIdFromJwtToken(request)

	storage_path := fmt.Sprintf("%s/services/%s/medias", user_id, service_id)
	uploaded_files, err := file_upload_service.UploadFiles(response, request, storage_path)

	if err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to upload service media", err)
		return
	}

	for _, file := range uploaded_files {
		service_media := &model.ServiceMedia{}
		service_media.CreatedBy = user_id
		service_media.ServiceID = service_id
		service_media.FileName = file.FileName
		service_media.URL = file.Url
		service_media.StoragePath = file.StoragePath

		if err := db.Create(service_media).Error; err != nil {
			continue
		}
	}

	helpers.HandleSuccess(response, http.StatusOK, "Service media uploaded", uploaded_files)
}

func DeleteServiceMedia(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	media_id := chi.URLParam(request, "media_id")

	tx := db.Begin()
	if tx.Error != nil {
		err := errors.New("something went wrong. please try again")
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a service", err)
		return
	}

	var service_media model.ServiceMedia

	if err := tx.Clauses(clause.Returning{}).Where("id = ?", media_id).Delete(&service_media).Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to delete a service media", err)
		return
	}

	if err := file_upload_service.DeleteFileFromStorage([]string{service_media.StoragePath}); err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to delete media", err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a service", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Service media deleted", service_media)
}

func GetServiceDayTimings(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	service_day_id := chi.URLParam(request, "service_day_id")

	var service_timings []model.ServiceTiming

	if err := db.Where("service_day_id = ?", service_day_id).Find(&service_timings).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Couldn't create service", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Service timing fetched", service_timings)
}

func AddUpdateServiceTimings(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	service_day_id := chi.URLParam(request, "service_day_id")

	var service_timings []*model.ServiceTiming
	if err := helpers.ParseBody(request, &service_timings, false); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to parse the body", err)
		return
	}

	for _, timing := range service_timings {
		timing.CreatedBy = helpers.GetUserIdFromJwtToken(request)
		timing.ServiceDayID = service_day_id
	}

	if err := db.Save(&service_timings).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Couldn't create service timings", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusCreated, "Service timings created", service_timings)
}

func DeleteServiceTimings(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	service_day_id := chi.URLParam(request, "service_day_id")
	service_timing_id := chi.URLParam(request, "service_timing_id")

	var service_timing model.ServiceTiming
	if err := db.Clauses(clause.Returning{}).Where("id = ? AND service_day_id = ?", service_timing_id, service_day_id).Delete(&service_timing).Error; err != nil {
		if err == gorm.ErrForeignKeyViolated {
			msg := "You cannot delete this timing due to timing is linked with appointment(s). You can modify it."
			helpers.HandleError(response, http.StatusInternalServerError, msg, err)
			return
		} else {
			helpers.HandleError(response, http.StatusInternalServerError, "Couldn't delete service timing", err)
			return
		}
	}

	helpers.HandleSuccess(response, http.StatusOK, "Service timing delete", service_timing)
}
