package utils

import (
	"encoding/json"
)

// GenerateEmptyArray generates an empty array with length l to fill in some td in table/struct
func GenerateEmptyArray(l int) []string {
	var arr []string
	for i := 0; i < l; i++ {
		arr = append(arr, "")
	}
	return arr
}

// PrettyPrint to print struct in a readable way
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
