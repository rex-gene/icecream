package utils

func IsStringValid(str string) bool {
	for _, c := range str {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
			return false
		}
	}

	return true
}
