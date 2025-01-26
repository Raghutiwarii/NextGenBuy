package utils

import (
	"fmt"
	"regexp"
	"strings"
)

func IsValidPhoneNumber(phoneNumber string) bool {
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "(", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ")", "")

	match, err := regexp.MatchString("^[0-9]{10}$", phoneNumber)
	if err != nil {
		fmt.Println("Error in matching phone number regex:", err)
		return false
	}

	return match
}
