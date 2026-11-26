// ============================================================================
// KULLANICI HESAP SERVİSİ (USER SERVICE)
// ============================================================================
// Oturum açmış kullanıcının kendi hesabıyla ilgili işlemleri (profil ve şifre
// güncelleme) ile dolandırıcılık/spam şikayetlerini yürütür.
// Kimlik doğrulamadan ayrı tutulmuştur (Tek Sorumluluk İlkesi / SRP).
// ============================================================================

package services

import (
	"context"
	"errors"
	"strings"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

var (
	ErrUserNotFound     = repository.ErrUserNotFound
	ErrCannotReportSelf = repository.ErrCannotReportSelf
	ErrWrongPassword    = errors.New("mevcut şifreniz hatalı")
)

const (
	FraudReportsForDelete = 3 // Bu sayıda farklı kullanıcı şikayet edince hesap silinir
	defaultFraudReason    = "Dolandırıcılık veya spam ihlali şüphesi"
	maxFraudReasonRunes   = 500
)

// UserServiceInterface, kullanıcı hesap işlemlerinin sözleşmesidir.
type UserServiceInterface interface {
	UpdateProfile(ctx context.Context, id int64, name, email string) (*models.UserDTO, error)
	UpdatePassword(ctx context.Context, id int64, oldPassword, newPassword, confirmPassword string) error
	ReportFraud(ctx context.Context, reporterID, reportedUserID int64, reason string) (reportCount int, deleted bool, err error)
}

type UserService struct {
	userRepo repository.UserRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
	return &UserService{userRepo: userRepo}
}

// UpdateProfile, kullanıcının adını ve e-postasını günceller.
func (s *UserService) UpdateProfile(ctx context.Context, id int64, name, email string) (*models.UserDTO, error) {
	name, err := normalizeAndValidateName(name)
	if err != nil {
		return nil, err
	}
	email, err = normalizeAndValidateEmail(email)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateProfile(ctx, id, name, email); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user.ToDTO(), nil
}

// UpdatePassword, mevcut şifreyi doğrulayarak yeni şifre belirler.
func (s *UserService) UpdatePassword(ctx context.Context, id int64, oldPassword, newPassword, confirmPassword string) error {
	if oldPassword == "" || newPassword == "" {
		return ErrRequiredFieldsEmpty
	}
	if newPassword != confirmPassword {
		return ErrPasswordsDoNotMatch
	}
	if len(newPassword) < 8 || len(newPassword) > 72 {
		return ErrPasswordInvalid
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if !auth.CheckPasswordHash(oldPassword, user.PasswordHash) {
		return ErrWrongPassword
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(ctx, id, hash)
}

// ReportFraud, dolandırıcılık/spam şikayetini işler.
// FraudReportsForDelete sayısına ulaşıldığında kullanıcının hesabı silinir.
func (s *UserService) ReportFraud(ctx context.Context, reporterID, reportedUserID int64, reason string) (int, bool, error) {
	if reporterID == reportedUserID {
		return 0, false, ErrCannotReportSelf
	}

	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = defaultFraudReason
	}
	if r := []rune(reason); len(r) > maxFraudReasonRunes {
		reason = string(r[:maxFraudReasonRunes])
	}

	reportCount, err := s.userRepo.ReportUserFraud(ctx, reporterID, reportedUserID, reason)
	if err != nil {
		return 0, false, err
	}

	if reportCount >= FraudReportsForDelete {
		if err := s.userRepo.DeleteUser(ctx, reportedUserID); err != nil {
			return reportCount, false, err
		}
		return reportCount, true, nil
	}
	return reportCount, false, nil
}
