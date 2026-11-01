package services_test

import (
	"bytes"
	"context"
	"database/sql"
	"mime/multipart"
	"net/textproto"
	"os"
	"strings"
	"testing"
	"time"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

// ============================================================================
// MOCK REPOSITORY
// ============================================================================

type mockListingRepo struct {
	listings map[int64]*models.Listing
	photos   map[int64][]models.ListingPhoto
	nextID   int64
	nextPID  int64
}

func newMockListingRepo() *mockListingRepo {
	return &mockListingRepo{
		listings: make(map[int64]*models.Listing),
		photos:   make(map[int64][]models.ListingPhoto),
		nextID:   1,
		nextPID:  1,
	}
}

func (m *mockListingRepo) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return nil, nil
}

func (m *mockListingRepo) CreateListing(ctx context.Context, tx *sql.Tx, input models.ListingCreateInput, userID int64) (*models.Listing, error) {
	id := m.nextID
	m.nextID++

	l := &models.Listing{
		ID:           id,
		UserID:       userID,
		Brand:        input.Brand,
		Model:        input.Model,
		Year:         input.Year,
		Mileage:      input.Mileage,
		FuelType:     input.FuelType,
		Transmission: input.Transmission,
		City:         input.City,
		Description:  input.Description,
		DamageStatus: input.DamageStatus,
		EngineVolume: input.EngineVolume,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.listings[id] = l
	return l, nil
}

func (m *mockListingRepo) AddPhoto(ctx context.Context, tx *sql.Tx, listingID int64, photoURL string, order int) (*models.ListingPhoto, error) {
	pid := m.nextPID
	m.nextPID++

	p := models.ListingPhoto{
		ID:           pid,
		ListingID:    listingID,
		PhotoURL:     photoURL,
		DisplayOrder: order,
		CreatedAt:    time.Now(),
	}
	m.photos[listingID] = append(m.photos[listingID], p)
	return &p, nil
}

func (m *mockListingRepo) GetListingByID(ctx context.Context, id int64) (*models.Listing, error) {
	l, ok := m.listings[id]
	if !ok {
		return nil, services.ErrListingNotFound
	}
	l.Photos = m.photos[id]
	return l, nil
}

func (m *mockListingRepo) ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error) {
	var items []*models.Listing
	for _, l := range m.listings {
		items = append(items, l)
	}
	return &models.Page[*models.Listing]{
		Items:      items,
		Total:      len(items),
		Page:       1,
		PerPage:    20,
		TotalPages: 1,
	}, nil
}

func (m *mockListingRepo) GetListingsByUserID(ctx context.Context, userID int64) ([]*models.Listing, error) {
	var items []*models.Listing
	for _, l := range m.listings {
		if l.UserID == userID {
			items = append(items, l)
		}
	}
	return items, nil
}

func (m *mockListingRepo) UpdateListing(ctx context.Context, id int64, input models.ListingCreateInput) (*models.Listing, error) {
	l, ok := m.listings[id]
	if !ok {
		return nil, services.ErrListingNotFound
	}
	l.Brand = input.Brand
	l.Model = input.Model
	l.Year = input.Year
	l.Mileage = input.Mileage
	l.FuelType = input.FuelType
	l.Transmission = input.Transmission
	l.City = input.City
	l.DamageStatus = input.DamageStatus
	l.EngineVolume = input.EngineVolume
	l.Description = input.Description
	l.UpdatedAt = time.Now()
	return l, nil
}

func (m *mockListingRepo) DeleteListing(ctx context.Context, id int64) error {
	delete(m.listings, id)
	delete(m.photos, id)
	return nil
}

func (m *mockListingRepo) GetPhotosByListingID(ctx context.Context, listingID int64) ([]models.ListingPhoto, error) {
	return m.photos[listingID], nil
}

func (m *mockListingRepo) DeletePhoto(ctx context.Context, photoID int64) (*models.ListingPhoto, error) {
	return nil, nil
}

func (m *mockListingRepo) CountPhotosByListingID(ctx context.Context, listingID int64) (int, error) {
	return len(m.photos[listingID]), nil
}

// ============================================================================
// 1. DOĞRULAMA (VALIDATION) TESTLERİ
// ============================================================================

func TestValidateListing_Valid(t *testing.T) {
	currentYear := time.Now().Year()
	engineVol := 1598

	validInput := models.ListingCreateInput{
		Brand:        "Toyota",
		Model:        "Corolla",
		Year:         currentYear - 2,
		Mileage:      45000,
		FuelType:     "benzin",
		Transmission: "otomatik",
		City:         "Ankara",
		DamageStatus: models.DamageUnknown,
		EngineVolume: &engineVol,
		Description:  "Yetkili servis bakımlı",
	}

	if err := services.ValidateListing(validInput); err != nil {
		t.Fatalf("Geçerli ilan doğrulanamadı: %v", err)
	}
}

func TestValidateListing_InvalidFields(t *testing.T) {
	currentYear := time.Now().Year()

	tests := []struct {
		name          string
		mutate        func(i *models.ListingCreateInput)
		expectedError string
	}{
		{
			name: "Boş marka",
			mutate: func(i *models.ListingCreateInput) {
				i.Brand = ""
			},
			expectedError: "marka",
		},
		{
			name: "Boş model",
			mutate: func(i *models.ListingCreateInput) {
				i.Model = ""
			},
			expectedError: "model",
		},
		{
			name: "1950'den eski yıl",
			mutate: func(i *models.ListingCreateInput) {
				i.Year = 1949
			},
			expectedError: "yıl",
		},
		{
			name: "Gelecek yıl",
			mutate: func(i *models.ListingCreateInput) {
				i.Year = currentYear + 2
			},
			expectedError: "yıl",
		},
		{
			name: "Negatif kilometre",
			mutate: func(i *models.ListingCreateInput) {
				i.Mileage = -5
			},
			expectedError: "kilometre",
		},
		{
			name: "Geçersiz yakıt türü",
			mutate: func(i *models.ListingCreateInput) {
				i.FuelType = "nukleer"
			},
			expectedError: "yakıt",
		},
		{
			name: "Geçersiz vites türü",
			mutate: func(i *models.ListingCreateInput) {
				i.Transmission = "roket"
			},
			expectedError: "vites",
		},
		{
			name: "Boş şehir",
			mutate: func(i *models.ListingCreateInput) {
				i.City = ""
			},
			expectedError: "şehir",
		},
		{
			name: "Geçersiz hasar durumu",
			mutate: func(i *models.ListingCreateInput) {
				i.DamageStatus = "pert_arac_tanimsiz"
			},
			expectedError: "hasar durumu",
		},
		{
			name: "Çok uzun açıklama (>5000 karakter)",
			mutate: func(i *models.ListingCreateInput) {
				i.Description = strings.Repeat("A", 5001)
			},
			expectedError: "açıklama",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := models.ListingCreateInput{
				Brand:        "Renault",
				Model:        "Megane",
				Year:         2020,
				Mileage:      50000,
				FuelType:     "dizel",
				Transmission: "manuel",
				City:         "İzmir",
				DamageStatus: models.DamageUnknown,
			}
			tt.mutate(&input)

			err := services.ValidateListing(input)
			if err == nil {
				t.Fatalf("Hata bekleniyordu ancak nil döndü")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.expectedError)) {
				t.Errorf("Beklenen hata ifadesi '%s', gelen: '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

// ============================================================================
// 2. YETKİLENDİRME (AUTHORIZATION) TESTLERİ
// ============================================================================

func TestCanModifyListing(t *testing.T) {
	owner := &models.UserDTO{ID: 10, Email: "owner@test.com", Role: models.RoleUser}
	stranger := &models.UserDTO{ID: 20, Email: "stranger@test.com", Role: models.RoleUser}
	admin := &models.UserDTO{ID: 99, Email: "admin@test.com", Role: models.RoleAdmin}

	listing := &models.Listing{
		ID:     1,
		UserID: 10, // Sahip ID: 10
		Brand:  "Honda",
		Model:  "Civic",
	}

	// 1. İlan sahibi düzenleyebilir/silebilir
	if !services.CanModifyListing(owner, listing) {
		t.Errorf("İlan sahibi (ID 10) kendi ilanını değiştirebilmeliydi")
	}

	// 2. Başka kullanıcı düzenleyemez/silemez (403)
	if services.CanModifyListing(stranger, listing) {
		t.Errorf("Başka bir kullanıcı (ID 20) ilanı değiştirememeliydi")
	}

	// 3. Admin başkasının ilanını düzenleyebilir/silebilir
	if !services.CanModifyListing(admin, listing) {
		t.Errorf("Yönetici (Admin) ilanı değiştirebilmeliydi")
	}

	// 4. Giriş yapmamış kullanıcı (nil) değiştiremez
	if services.CanModifyListing(nil, listing) {
		t.Errorf("Giriş yapmamış anonim kullanıcı (nil) ilanı değiştirememeliydi")
	}
}

func TestListingService_UpdateAndDeleteAuthorization(t *testing.T) {
	repo := newMockListingRepo()
	svc := services.NewListingService(repo)
	ctx := context.Background()

	// İlanı oluşturalım (Sahip ID: 10)
	listing, err := repo.CreateListing(ctx, nil, models.ListingCreateInput{
		Brand:        "Ford",
		Model:        "Focus",
		Year:         2018,
		Mileage:      70000,
		FuelType:     "benzin",
		Transmission: "manuel",
		City:         "Bursa",
		DamageStatus: models.DamageUnknown,
	}, 10)
	if err != nil {
		t.Fatalf("İlan oluşturulamadı: %v", err)
	}

	owner := &models.UserDTO{ID: 10, Role: models.RoleUser}
	stranger := &models.UserDTO{ID: 20, Role: models.RoleUser}
	admin := &models.UserDTO{ID: 99, Role: models.RoleAdmin}

	updateInput := models.ListingCreateInput{
		Brand:        "Ford",
		Model:        "Focus Titanium",
		Year:         2018,
		Mileage:      75000,
		FuelType:     "benzin",
		Transmission: "manuel",
		City:         "Bursa",
		DamageStatus: models.DamageUnknown,
	}

	// Başkası güncellemeye çalıştığında ErrListingForbidden (403) dönmeli
	_, err = svc.UpdateListing(ctx, stranger, listing.ID, updateInput)
	if err != services.ErrListingForbidden {
		t.Errorf("Yabancı kullanıcı güncellemesinde ErrListingForbidden bekleniyordu, gelen: %v", err)
	}

	// Sahip başarıyla güncelleyebilmeli
	updated, err := svc.UpdateListing(ctx, owner, listing.ID, updateInput)
	if err != nil {
		t.Fatalf("Sahip güncelleme yapamadı: %v", err)
	}
	if updated.Model != "Focus Titanium" {
		t.Errorf("Model güncellenemedi: %s", updated.Model)
	}

	// Başkası silmeye çalıştığında ErrListingForbidden (403) dönmeli
	err = svc.DeleteListing(ctx, stranger, listing.ID, os.TempDir())
	if err != services.ErrListingForbidden {
		t.Errorf("Yabancı kullanıcı silmesinde ErrListingForbidden bekleniyordu, gelen: %v", err)
	}

	// Admin başkasının ilanını silebilmeli
	err = svc.DeleteListing(ctx, admin, listing.ID, os.TempDir())
	if err != nil {
		t.Fatalf("Admin ilanı silemedi: %v", err)
	}
}

// ============================================================================
// 3. FOTOĞRAF DOĞRULAMA (MIME / BOYUT) TESTLERİ
// ============================================================================

func TestPhotoValidation_MimeAndContent(t *testing.T) {
	// 1. Geçerli JPEG içeriği
	jpegBytes := services.MakeTestJPEGBytes()
	if !services.IsValidPhotoContent(jpegBytes) {
		t.Errorf("Geçerli JPEG içeriği onaylanamadı")
	}

	// 2. Geçerli PNG içeriği
	pngBytes := services.MakeTestPNGBytes()
	if !services.IsValidPhotoContent(pngBytes) {
		t.Errorf("Geçerli PNG içeriği onaylanamadı")
	}

	// 3. Yanıltıcı dosya (Uzantısı jpg ama içeriği zararlı text / script / html)
	fakeJpg := []byte("<html><body><h1>Bu bir fotoğraf değil bir HTML belgesidir!</h1></body></html>")
	if services.IsValidPhotoContent(fakeJpg) {
		t.Errorf("HTML içeriği taşıyan sahte JPG reddedilmeliydi!")
	}

	// 4. Rastgele binary / exe
	fakeExe := []byte("MZ\x90\x00\x03\x00\x00\x00\x04\x00\x00\x00\xff\xff\x00\x00")
	if services.IsValidPhotoContent(fakeExe) {
		t.Errorf("Executable dosya içeriği reddedilmeliydi!")
	}

	// 5. Boş dosya
	if services.IsValidPhotoContent([]byte{}) {
		t.Errorf("Boş dosya reddedilmeliydi!")
	}
}

func TestPhotoUpload_SizeAndTypeErrors(t *testing.T) {
	repo := newMockListingRepo()
	svc := services.NewListingService(repo)
	ctx := context.Background()
	user := &models.UserDTO{ID: 1, Role: models.RoleUser}
	tmpDir := os.TempDir()

	validInput := models.ListingCreateInput{
		Brand:        "Toyota",
		Model:        "Yaris",
		Year:         2021,
		Mileage:      30000,
		FuelType:     "benzin",
		Transmission: "otomatik",
		City:         "Ankara",
		DamageStatus: models.DamageUnknown,
	}

	// 1. Yanlış türde dosya yükleme (Uzantı .jpg ancak içerik metin)
	fakeHeader := createMultipartFileHeader("hacker.jpg", []byte("plain text content pretending to be image"))
	_, err := svc.CreateListingWithPhotos(ctx, user, validInput, []*multipart.FileHeader{fakeHeader}, tmpDir)
	if err == nil {
		t.Fatalf("Sahte JPEG içeriği reddedilmeliydi ancak onaylandı")
	}
	if !strings.Contains(err.Error(), "JPEG") && !strings.Contains(err.Error(), "türü") {
		t.Errorf("Beklenen tür hatası alınamadı: %v", err)
	}

	// 2. Çok büyük dosya (> 5 MB)
	bigBytes := make([]byte, 6*1024*1024) // 6 MB
	copy(bigBytes, services.MakeTestJPEGBytes())
	bigHeader := createMultipartFileHeader("huge.jpg", bigBytes)
	_, err = svc.CreateListingWithPhotos(ctx, user, validInput, []*multipart.FileHeader{bigHeader}, tmpDir)
	if err == nil {
		t.Fatalf("6 MB dosya boyutu reddedilmeliydi ancak onaylandı")
	}
	if !strings.Contains(err.Error(), "5 MB") {
		t.Errorf("Beklenen dosya boyutu hatası alınamadı: %v", err)
	}

	// 3. Geçerli fotoğraf ile yükleme
	validHeader := createMultipartFileHeader("car.jpg", services.MakeTestJPEGBytes())
	created, err := svc.CreateListingWithPhotos(ctx, user, validInput, []*multipart.FileHeader{validHeader}, tmpDir)
	if err != nil {
		t.Fatalf("Geçerli fotoğraf ile ilan oluşturulamadı: %v", err)
	}
	if len(created.Photos) != 1 {
		t.Errorf("1 fotoğraf bekleniyordu, gelen: %d", len(created.Photos))
	}
}

// createMultipartFileHeader, testler için in-memory multipart.FileHeader simüle eder.
func createMultipartFileHeader(filename string, content []byte) *multipart.FileHeader {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="photos"; filename="`+filename+`"`)
	part, _ := writer.CreatePart(h)
	_, _ = part.Write(content)
	_ = writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(int64(len(content) + 1024))
	return form.File["photos"][0]
}
