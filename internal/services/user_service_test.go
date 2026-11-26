package services

import (
	"context"
	"errors"
	"testing"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
)

func newTestUserService(t *testing.T) (*UserService, *mockUserRepo) {
	t.Helper()
	repo := newMockUserRepo()
	hash, err := auth.HashPassword("EskiSifre1!")
	if err != nil {
		t.Fatal(err)
	}
	for i, name := range []string{"Bir", "Iki", "Uc", "Dort"} {
		u := &models.User{ID: int64(i + 1), Name: name, Email: name + "@test.com", PasswordHash: hash}
		repo.users[u.ID] = u
	}
	repo.nextID = 5
	return NewUserService(repo), repo
}

func TestUpdateProfile_Validation(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	if _, err := svc.UpdateProfile(ctx, 1, "Yeni Ad", "gecersiz"); !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("ErrInvalidEmail bekleniyordu, alınan: %v", err)
	}
	if _, err := svc.UpdateProfile(ctx, 1, "a@b", "bir@test.com"); !errors.Is(err, ErrInvalidUsername) {
		t.Errorf("ErrInvalidUsername bekleniyordu, alınan: %v", err)
	}
	user, err := svc.UpdateProfile(ctx, 1, "  YeniAd  ", "yeni@test.com")
	if err != nil || user.Name != "YeniAd" {
		t.Errorf("Geçerli profil güncellenemedi: %+v, %v", user, err)
	}
}

func TestUpdatePassword(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	if err := svc.UpdatePassword(ctx, 1, "YanlisEski!", "YeniSifre1!", "YeniSifre1!"); !errors.Is(err, ErrWrongPassword) {
		t.Errorf("ErrWrongPassword bekleniyordu, alınan: %v", err)
	}
	if err := svc.UpdatePassword(ctx, 1, "EskiSifre1!", "YeniSifre1!", "YeniSifre1!"); err != nil {
		t.Errorf("Şifre güncellenemedi: %v", err)
	}
}

func TestReportFraud(t *testing.T) {
	svc, repo := newTestUserService(t)
	ctx := context.Background()

	if _, _, err := svc.ReportFraud(ctx, 1, 1, ""); !errors.Is(err, ErrCannotReportSelf) {
		t.Errorf("ErrCannotReportSelf bekleniyordu, alınan: %v", err)
	}
	if _, _, err := svc.ReportFraud(ctx, 1, 999, ""); !errors.Is(err, ErrUserNotFound) {
		t.Errorf("ErrUserNotFound bekleniyordu, alınan: %v", err)
	}

	for reporter := int64(1); reporter <= 2; reporter++ {
		_, deleted, err := svc.ReportFraud(ctx, reporter, 4, "spam")
		if err != nil || deleted {
			t.Fatalf("İlk şikayetlerde hesap silinmemeli: deleted=%v err=%v", deleted, err)
		}
	}
	count, deleted, err := svc.ReportFraud(ctx, 3, 4, "spam")
	if err != nil || !deleted || count != FraudReportsForDelete {
		t.Fatalf("%d. şikayette hesap silinmeliydi: count=%d deleted=%v err=%v", FraudReportsForDelete, count, deleted, err)
	}
	if _, ok := repo.users[4]; ok {
		t.Error("Kullanıcı silinmiş olmalıydı")
	}
}
