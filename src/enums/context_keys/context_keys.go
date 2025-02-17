package context_keys

type ContextKeys string

var (
	DB        ContextKeys = "db"
	USER_ID   ContextKeys = "user_id"
	USER_TYPE ContextKeys = "user_type"
)
