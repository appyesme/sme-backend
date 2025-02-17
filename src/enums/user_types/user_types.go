package user_types

const (
	USER         string = "USER"
	ENTREPRENEUR string = "ENTREPRENEUR"
	ADMIN        string = "ADMIN"
)

var UserTypes = []string{USER, ENTREPRENEUR, ADMIN}

// ContainsUserType checks if the input string matches any valid UserType
func ContainsUserType(input string) bool {
	for _, userType := range UserTypes {
		if input == userType {
			return true
		}
	}
	return false
}
