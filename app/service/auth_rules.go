package service

import (
	"strings"
	"unicode"

	"latihan-fiber/app/model"
)

func ValidateRegister(req model.RegisterRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.Username) == "" {
		errs["username"] = "wajib diisi"
	}
	if strings.TrimSpace(req.Email) == "" {
		errs["email"] = "wajib diisi"
	}
	if message := checkPasswordStrength(req.Password); message != "" {
		errs["password"] = message
	}
	return errs
}

func ValidateLogin(req model.LoginRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.Username) == "" {
		errs["username"] = "wajib diisi"
	}
	if req.Password == "" {
		errs["password"] = "wajib diisi"
	}
	return errs
}

func checkPasswordStrength(password string) string {
	if len([]rune(password)) < 8 {
		return "minimal 8 karakter"
	}
	lower := strings.ToLower(password)
	common := map[string]bool{
		"password": true, "password123": true, "12345678": true,
		"qwerty123": true, "admin123": true,
	}
	if common[lower] {
		return "password terlalu umum"
	}
	var hasLetter, hasNumber bool
	for _, r := range password {
		hasLetter = hasLetter || unicode.IsLetter(r)
		hasNumber = hasNumber || unicode.IsNumber(r)
	}
	if !hasLetter || !hasNumber {
		return "harus mengandung huruf dan angka"
	}
	return ""
}
