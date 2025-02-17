package post_service

import (
	"encoding/json"
	"sme-backend/model"
)

type GetPostsDto struct {
	model.Post
	Liked      bool            `json:"liked"`
	TotalLikes int32           `json:"total_likes"`
	Author     json.RawMessage `json:"author"`
	Medias     json.RawMessage `json:"medias"`
}
