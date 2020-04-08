package api

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var (
	emailRegexp = regexp.MustCompile(`^[A-z0-9!#$%&'*+\/=?^_` + "`" + `{|}~-]+(?:\.[A-z0-9!#$%&'*+\/=?^_` + "`" + `{|}~-]+)*@(?:[A-z0-9](?:[A-z0-9-]*[A-z0-9])?\.)+[A-z0-9](?:[A-z0-9-]*[A-z0-9])?$`)
)

func hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func compare(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return false
	}

	return true
}

func isValidEmail(email string) bool {
	return emailRegexp.MatchString(email)
}
