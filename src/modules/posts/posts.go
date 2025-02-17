package posts_handler

import (
	"errors"
	"fmt"
	"net/http"
	"sme-backend/model"
	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
	"sme-backend/src/services/post_service"
	file_upload_service "sme-backend/src/services/upload_service"

	"github.com/go-chi/chi"
	"gorm.io/gorm/clause"
)

func GetPosts(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	profile_id := helpers.GetQueryParameter(request, "profile_id")

	page := helpers.GetQueryPageParam(request)
	limit := helpers.GetQueryLimitParam(request)

	posts := make([]post_service.GetPostsDto, 0)

	if err := post_service.GetPosts(db, page, limit, profile_id, "", "", &posts); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to get posts", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Posts fetched", posts)
}

func CreateUpdatePost(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	user_id := helpers.GetUserIdFromJwtToken(request)
	var post model.Post
	if err := helpers.ParseBody(request, &post, false); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to parse the body", err)
		return
	}

	post.CreatedBy = user_id
	if post.ID != "" {
		if err := db.Where("id = ?", post.ID).Updates(&post).Error; err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Couldn't update post", err)
			return
		}
	} else {
		if err := db.Create(&post).Error; err != nil {
			helpers.HandleError(response, http.StatusInternalServerError, "Couldn't create post", err)
			return
		}
	}

	var posts []post_service.GetPostsDto
	if err := post_service.GetPosts(db, 0, 1, "", post.ID, "", &posts); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to get posts", err)
		return
	}

	if posts != nil && len(posts) != 1 {
		helpers.HandleError(response, http.StatusInternalServerError, "Something went wrong", errors.New("something went wrong"))
		return
	}

	helpers.HandleSuccess(response, http.StatusCreated, "Post created", posts[0])
}

func DeletePost(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)

	post_id := helpers.GetUrlParam(request, "post_id")

	tx := db.Begin()
	if tx.Error != nil {
		err := errors.New("something went wrong. please try again")
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a post", err)
		return
	}

	var post model.Post
	var post_medias_storage_path []string

	if err := tx.Model(model.PostMedia{}).Where("post_id = ?", post_id).Select("storage_path").Find(&post_medias_storage_path).Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a post", err)
		return
	}

	// folder_path := fmt.Sprintf("public/%s/post-medias/%s", user_id, post_id)
	if err := file_upload_service.DeleteFileFromStorage(post_medias_storage_path); err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a post", err)
		return
	}

	if err := tx.Where("id = ?", post_id).Clauses(clause.Returning{}).Delete(&post).Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a post", err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		helpers.HandleError(response, http.StatusConflict, "Unable to delete a post", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Post deleted", post)
}

func DeletePostMedia(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	media_id := chi.URLParam(request, "media_id")

	var post_media model.PostMedia
	if err := db.Where("id = ?", media_id).First(&post_media).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to delete a post media", err)
		return
	}

	if err := file_upload_service.DeleteFileFromStorage([]string{post_media.StoragePath}); err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to delete media", err)
		return
	}

	if err := db.Clauses(clause.Returning{}).Where("id = ?", media_id).Delete(&post_media).Error; err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to delete a post media", err)
		return
	}

	helpers.HandleSuccess(response, http.StatusOK, "Post media deleted", post_media)
}

func UploadPostMedia(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	post_id := chi.URLParam(request, "post_id")
	user_id := helpers.GetUserIdFromJwtToken(request)

	storage_path := fmt.Sprintf("%s/posts/medias/%s", user_id, post_id)
	uploaded_files, err := file_upload_service.UploadFiles(response, request, storage_path)

	if err != nil {
		helpers.HandleError(response, http.StatusInternalServerError, "Unable to upload post media", err)
		return
	}

	for _, file := range uploaded_files {
		post_media := &model.PostMedia{}
		post_media.PostID = post_id
		post_media.CreatedBy = user_id
		post_media.FileName = file.FileName
		post_media.URL = file.Url
		post_media.StoragePath = file.StoragePath

		if err := db.Create(post_media).Error; err != nil {
			continue
		}
	}

	helpers.HandleSuccess(response, http.StatusOK, "Post media uploaded", uploaded_files)
}
