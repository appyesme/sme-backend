package auth_types

const (
	SIGNIN string = "SIGNIN"
	SIGNUP string = "SIGNUP"
)

var authTypes = []string{SIGNIN, SIGNUP}

func ContainsAuthType(input string) bool {
	for _, userType := range authTypes {
		if input == userType {
			return true
		}
	}
	return false
}
