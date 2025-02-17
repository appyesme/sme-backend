package webhooks

type WebhookBunnyVideoStatusBody struct {
	VideoLibraryId int    `json:"videoLibraryId"`
	VideoGuid      string `json:"VideoGuid"`
	Status         int    `json:"Status"`
}
