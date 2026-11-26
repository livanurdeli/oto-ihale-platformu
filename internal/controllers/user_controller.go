package controllers

// ============================================================================
// KULLANICI HESAP DENETLEYİCİSİ (USER CONTROLLER)
// ============================================================================
// Kullanıcı paneli, profil/şifre güncelleme ve dolandırıcılık şikayeti uçlarını
// yönetir. İş mantığı UserService'tedir.
// ============================================================================

import (
	"fmt"
	"html/template"
	"net/http"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/services"
)

var profileErrors = []error{
	services.ErrRequiredFieldsEmpty,
	services.ErrInvalidEmail,
	services.ErrInvalidUsername,
	services.ErrEmailAlreadyTaken,
	services.ErrUsernameTaken,
	services.ErrPasswordInvalid,
	services.ErrPasswordsDoNotMatch,
	services.ErrWrongPassword,
}

type UserController struct {
	userService   services.UserServiceInterface
	dashboardView *template.Template
}

func NewUserController(userService services.UserServiceInterface) (*UserController, error) {
	view, err := LoadView("dashboard.html")
	if err != nil {
		return nil, err
	}
	return &UserController{userService: userService, dashboardView: view}, nil
}

// Dashboard, oturum açmış kullanıcının panelini gösterir.
func (c *UserController) Dashboard(w http.ResponseWriter, r *http.Request) {
	RenderView(w, http.StatusOK, c.dashboardView, NewViewData(r, "Kullanıcı Paneli", "dashboard"))
}

// UpdateProfilePost, kullanıcı adı ve e-posta güncellemesini işler.
func (c *UserController) UpdateProfilePost(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		middlewares.SetFlash(w, middlewares.FlashError, "Form verisi okunamadı")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if _, err := c.userService.UpdateProfile(r.Context(), user.ID, r.FormValue("name"), r.FormValue("email")); err != nil {
		middlewares.SetFlash(w, middlewares.FlashError, userFacingError(err, profileErrors...))
	} else {
		middlewares.SetFlash(w, middlewares.FlashSuccess, "Profil bilgileriniz başarıyla güncellendi.")
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// UpdatePasswordPost, şifre değişikliğini işler.
func (c *UserController) UpdatePasswordPost(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		middlewares.SetFlash(w, middlewares.FlashError, "Form verisi okunamadı")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	err := c.userService.UpdatePassword(r.Context(), user.ID,
		r.FormValue("old_password"), r.FormValue("new_password"), r.FormValue("confirm_password"))
	if err != nil {
		middlewares.SetFlash(w, middlewares.FlashError, userFacingError(err, profileErrors...))
	} else {
		middlewares.SetFlash(w, middlewares.FlashSuccess, "Şifreniz başarıyla güncellendi.")
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// ReportFraudPost, bir kullanıcıyı dolandırıcılık/spam gerekçesiyle şikayet eder.
// İşlem sonrası formdaki güvenli 'return_to' adresine dönülür.
func (c *UserController) ReportFraudPost(w http.ResponseWriter, r *http.Request) {
	reporter := middlewares.GetUserFromContext(r.Context())

	returnTo := "/listings"
	if err := r.ParseForm(); err == nil && isSafeRedirect(r.FormValue("return_to")) {
		returnTo = r.FormValue("return_to")
	}

	targetID, err := parseIDParam(r, "id")
	if err != nil {
		middlewares.SetFlash(w, middlewares.FlashError, "Geçersiz kullanıcı.")
		http.Redirect(w, r, returnTo, http.StatusSeeOther)
		return
	}

	count, deleted, err := c.userService.ReportFraud(r.Context(), reporter.ID, targetID, r.FormValue("reason"))
	switch {
	case err != nil:
		middlewares.SetFlash(w, middlewares.FlashError,
			userFacingError(err, services.ErrCannotReportSelf, services.ErrUserNotFound))
	case deleted:
		middlewares.SetFlash(w, middlewares.FlashSuccess, fmt.Sprintf(
			"Kullanıcı %d farklı kişi tarafından bildirildiği için hesabı ve ilanları sistemden silindi.",
			services.FraudReportsForDelete))
		returnTo = "/listings" // İlan da silindiği için ilan sayfasına dönülemez
	default:
		middlewares.SetFlash(w, middlewares.FlashSuccess, fmt.Sprintf(
			"Şikayetiniz alındı. Bu kullanıcı için toplam bildirim: %d/%d.", count, services.FraudReportsForDelete))
	}
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}
