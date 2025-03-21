package search_service

import (
	"fmt"
	"sme-backend/src/enums/user_types"
	"sme-backend/src/services/post_service"

	"gorm.io/gorm"
)

type SearchedServicesDto struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	Address          string  `json:"address"`
	Charge           float64 `json:"charge"`
	AdditionalCharge float64 `json:"additional_charge"`
	HomeAvailable    bool    `json:"home_available"`
	SalonAvailable   bool    `json:"salon_available"`
	URL              *string `json:"url"`
}

type SearchedUsersDto struct {
	ID       string `json:"id" gorm:"id"`
	Name     string `json:"name" gorm:"name"`
	PhotoUrl string `json:"photo_url" gorm:"photo_url"`
}

func SearchServices(db *gorm.DB, searched_services *[]SearchedServicesDto, search_query string) error {
	query := `SELECT s.id, s.title, s.description, s.address, s.charge, s.additional_charge, s.home_available, s.salon_available, MIN(sm.url) AS url
	FROM services s
	LEFT JOIN service_medias sm ON sm.service_id = s.id
	WHERE s.status = 'PUBLISHED' AND ( s.title ILIKE $1 OR s.description ILIKE $1 OR s.address ILIKE $1 )
	GROUP BY s.id
	LIMIT 20;`
	search_term := fmt.Sprintf("%%%s%%", search_query)
	return db.Raw(query, search_term).Scan(&searched_services).Error
}

func SearchPosts(db *gorm.DB, searched_posts *[]post_service.GetPostsDto, search_query string) error {
	return post_service.GetPosts(db, 0, 20, "", "", search_query, searched_posts)
}

func SearchUsers(db *gorm.DB, searched_users *[]SearchedUsersDto, search_query string) error {
	query := `SELECT u.id, u.photo_url, u.name
	FROM users u LEFT JOIN auth a ON a.id = u.id
	WHERE a.user_type = $1 AND u.verified AND (u.name ILIKE $2 OR EXISTS (SELECT 1 FROM unnest(u.expertises) AS expertise WHERE expertise ILIKE $2))
	GROUP BY u.id, u.name
	LIMIT 20;`

	search_term := fmt.Sprintf("%%%s%%", search_query)

	return db.Raw(query, user_types.ENTREPRENEUR, search_term).Scan(&searched_users).Error
}
