package searches_handler

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"sme-backend/src/core/database"
	"sme-backend/src/core/helpers"
	search_service "sme-backend/src/services/searches_service"
)

func SearchServicesAndPostsAndUsers(response http.ResponseWriter, request *http.Request) {
	db := database.GetRlsContextDB(request)
	search_query := helpers.GetQueryParameter(request, "query")

	if search_query == "" || len(search_query) < 2 {
		err_msg := "Atleast 2 search characters required"
		helpers.HandleError(response, http.StatusBadRequest, err_msg, errors.New(strings.ToLower(err_msg)))
		return
	}

	var searched_services []search_service.SearchedServicesDto
	var searched_users []search_service.SearchedUsersDto
	// var searched_posts []post_service.GetPostsDto
	var wg sync.WaitGroup // WaitGroup to wait for both queries
	wg.Add(2)

	// Goroutine for searching posts
	// go func() {
	// 	defer wg.Done() // Mark this goroutine as done when finished
	// 	search_service.SearchPosts(db, &searched_posts, search_query)
	// }()

	// Goroutine for searching services
	go func() {
		defer wg.Done() // Mark this goroutine as done when finished
		search_service.SearchServices(db, &searched_services, search_query)
	}()

	// Goroutine for searching users
	go func() {
		defer wg.Done() // Mark this goroutine as done when finished
		search_service.SearchUsers(db, &searched_users, search_query)
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	resp := map[string]interface{}{"services": searched_services /* "posts": searched_posts, */, "users": searched_users}
	helpers.HandleSuccess(response, http.StatusOK, "Searched", resp)
}
