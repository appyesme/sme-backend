package post_service

import (
	"fmt"

	"gorm.io/gorm"
)

func GetPosts(tx *gorm.DB, page, limit int, profile_id, post_id, search_query string, posts *[]GetPostsDto) error {
	var query_conditions []any
	var query_values []any

	baseQuery := `SELECT
				p.*,
				json_build_object('id', u.id, 'name', u.name, 'photo_url', u.photo_url) AS author,
				COALESCE(json_agg(DISTINCT pm.*) FILTER (WHERE pm.id IS NOT NULL), '[]') AS medias
			FROM
				posts p
				JOIN users u ON u.id = p.created_by
				JOIN services s ON s.id = p.service_id
				LEFT JOIN post_medias AS pm ON pm.post_id = p.id
			WHERE p.status = 'PUBLISHED' %s
			GROUP BY p.id, u.id, u.photo_url
			ORDER BY p.created_at DESC
			LIMIT ? OFFSET ?;`

	/// START - This will joining with where condition.
	// 3rd [%s]
	if profile_id != "" {
		query_conditions = append(query_conditions, " AND p.created_by = ?")
		query_values = append(query_values, profile_id)
	}

	if post_id != "" {
		query_conditions = append(query_conditions, " AND p.id = ?")
		query_values = append(query_values, post_id)
	}

	if search_query != "" {
		query_conditions = append(query_conditions, "AND p.description ILIKE ?")
		search_term := fmt.Sprintf("%%%s%%", search_query) // "%search%"
		query_values = append(query_values, search_term)
	}

	if profile_id == "" && post_id == "" && search_query == "" {
		query_conditions = append(query_conditions, "")
	}
	/// END - This will joining with where condition.

	// Append limit and offset to query values
	query_values = append(query_values, limit, page*limit)

	finalQuery := fmt.Sprintf(baseQuery, query_conditions...)
	if err := tx.Raw(finalQuery, query_values...).Scan(&posts).Error; err != nil {
		return err
	}

	return nil
}
