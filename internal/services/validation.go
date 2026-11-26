package services

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

var (
	ErrRequiredFieldsEmpty = errors.New("lütfen tüm zorunlu alanları doldurunuz")
	ErrInvalidEmail        = errors.New("geçersiz e-posta adresi")
	ErrInvalidUsername     = errors.New("kullanıcı adı 3-50 karakter olmalı ve '@', '<', '>' karakterlerini içermemelidir")
)

// normalizeAndValidateEmail, e-postayı kırpar ve geçerli bir adres olup olmadığını denetler.
func normalizeAndValidateEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", ErrRequiredFieldsEmpty
	}
	addr, err := mail.ParseAddress(email)
	// "Ad <a@b.com>" gibi biçimler kabul edilmez; yalnızca çıplak adres geçerlidir
	if err != nil || addr.Address != email || strings.ContainsAny(email, "<>\"' ") {
		return "", ErrInvalidEmail
	}
	return email, nil
}

// normalizeAndValidateName, kullanıcı adını kırpar ve kurallara uygunluğunu denetler.
// '@' yasaktır: girişte '@' içeren tanımlayıcı e-posta olarak yorumlanır.
func normalizeAndValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrRequiredFieldsEmpty
	}
	n := utf8.RuneCountInString(name)
	if n < 3 || n > 50 || strings.ContainsAny(name, "@<>") {
		return "", ErrInvalidUsername
	}
	return name, nil
}
