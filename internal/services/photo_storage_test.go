package services

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newFileHeader, verilen ad ve içerikle gerçek bir multipart dosya başlığı üretir.
func newFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("photos", filename)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fw.Write(content)
	_ = mw.Close()

	req := httptest.NewRequest("POST", "/", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatal(err)
	}
	return req.MultipartForm.File["photos"][0]
}

var jpegMagic = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}

func TestLocalPhotoStorage_ExtensionComesFromContent(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalPhotoStorage(dir)
	if err != nil {
		t.Fatal(err)
	}

	// JPEG başlığıyla başlayan ama .html uzantılı (XSS denemesi) dosya
	content := append(append([]byte{}, jpegMagic...), []byte("<script>alert(1)</script>")...)
	url, err := store.Save(newFileHeader(t, "saldiri.html", content))
	if err != nil {
		t.Fatalf("Kaydetme başarısız: %v", err)
	}
	if !strings.HasPrefix(url, "/uploads/") || filepath.Ext(url) != ".jpg" {
		t.Errorf("Uzantı içerikten (.jpg) üretilmeliydi, alınan: %s", url)
	}

	if err := store.Delete(url); err != nil {
		t.Fatalf("Silme başarısız: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.Base(url))); !errors.Is(err, os.ErrNotExist) {
		t.Error("Dosya silinmiş olmalıydı")
	}
}

func TestLocalPhotoStorage_RejectsNonImage(t *testing.T) {
	store, _ := NewLocalPhotoStorage(t.TempDir())
	err := store.Validate(newFileHeader(t, "foto.jpg", []byte("<html><body>merhaba</body></html>")))
	if !errors.Is(err, ErrPhotoInvalidType) {
		t.Errorf("ErrPhotoInvalidType bekleniyordu, alınan: %v", err)
	}
}

func TestLocalPhotoStorage_DeleteCannotEscapeDir(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "uploads")
	store, _ := NewLocalPhotoStorage(dir)
	outside := filepath.Join(parent, "gizli.txt")
	_ = os.WriteFile(outside, []byte("x"), 0o644)

	_ = store.Delete("/uploads/../gizli.txt")
	if _, err := os.Stat(outside); err != nil {
		t.Error("Klasör dışındaki dosya silinmemeliydi")
	}
}
