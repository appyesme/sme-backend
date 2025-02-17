package helpers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sme-backend/src/enums/context_keys"
	"sme-backend/src/enums/user_types"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
}

var Validator = validator.New()

func GetUserType(request *http.Request) string {
	user_type := request.Context().Value(context_keys.USER_TYPE)
	if user_type == user_types.ENTREPRENEUR {
		return user_types.ENTREPRENEUR
	}
	return user_types.USER
}

func GetUserIdFromJwtToken(request *http.Request) string {
	user_id := request.Context().Value(context_keys.USER_ID)
	if user_id == nil {
		return ""
	}
	return user_id.(string)
}

func GetCurrentTime() time.Time {
	location, _ := time.LoadLocation("Asia/Kolkata")
	return time.Now().In(location)
}

func ParseBody(request *http.Request, data interface{}, validate bool) error {
	err := json.NewDecoder(request.Body).Decode(&data)
	if err == nil && validate {
		return Validator.Struct(data)
	}
	return err
}

func UnmarshalJSON(data []byte, target interface{}) error {
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, target)
}

func HandleSuccess(response http.ResponseWriter, statusCode int, message string, data interface{}) {
	payload := Response{Status: "success", Message: message, Data: data, Error: nil}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)
	json.NewEncoder(response).Encode(payload)
}

func HandleError(response http.ResponseWriter, statusCode int, message string, err error) {
	payload := Response{Status: "error", Message: message, Data: nil, Error: err.Error()}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)
	json.NewEncoder(response).Encode(payload)
}

func GetUrlParam(request *http.Request, key string) string {
	return chi.URLParam(request, key)
}

func GetQueryParameter(request *http.Request, key string) string {
	return request.URL.Query().Get(key)
}

func GetQueryPageParam(request *http.Request) int {
	page, err := strconv.Atoi(GetQueryParameter(request, "page"))
	if err != nil {
		page = 0
	}
	return page
}

func GetQueryLimitParam(request *http.Request) int {
	limit, err := strconv.Atoi(GetQueryParameter(request, "limit"))
	if err != nil {
		limit = 15
	}
	return limit
}

func GetQueryBoolParam(request *http.Request, key string) *bool {
	param := request.URL.Query().Get(key)
	var boolean bool
	if param == "true" {
		boolean = true
	} else if param == "false" {
		boolean = false
	}
	return &boolean
}

func GetQueryArray(request *http.Request, key string) ([]string, error) {
	values, ok := request.URL.Query()[key]
	if !ok {
		return nil, errors.New("invalid query parameter")
	}
	return values, nil
}
func ToRawMessage(data interface{}) *json.RawMessage {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	rawMsg := json.RawMessage(dataJSON)
	return &rawMsg
}
