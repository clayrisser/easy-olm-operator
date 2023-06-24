package util

// helper method to check if a string exists in an array of strings
func ContainsString(array []string, str string) bool {
	for _, a := range array {
		if a == str {
			return true
		}
	}
	return false
}
