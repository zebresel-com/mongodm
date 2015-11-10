package mongodm

import (
	"fmt"
	"regexp"
)

func validateEmail(email string) bool {

	regex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	return regex.MatchString(email)
}

func validateRegexp(regex string, target string) bool {

	match, err := regexp.MatchString(regex, target)

	if err != nil {
		fmt.Println("%v", err)
	}

	return match
}
