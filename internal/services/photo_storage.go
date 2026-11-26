package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	ErrPhotoTooLarge    = errors.New("fotoğraf dosyası 5 MB'dan büyük olamaz")
	ErrPhotoInvalidType = errors.New("yalnızca JPEG, PNG ve WebP fotoğraf yüklenebilir")
)

const MaxPhotoBytes = int64(5 * 1024 * 1024) // 5 MB

// allowedPhotoTypes, izin verilen MIME türlerini kaydedilecek dosya uzantısına eşler.
// Uzantı kullanıcının dosya adından DEĞİL, içerikten tespit edilen türden üretilir;
// böylece "resim.html" gibi bir dosya HTML olarak sunulamaz (XSS önlemi).
var allowedPhotoTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// PhotoStorage, ilan fotoğraflarının saklandığı ortamın soyutlamasıdır (DIP).
// Yerel disk yerine ileride nesne depolama (S3 vb.) kullanılabilir.
type PhotoStorage interface {
	// Validate, dosyanın boyut ve tür kurallarına uyduğunu denetler.
	Validate(fh *multipart.FileHeader) error
	// Save, dosyayı kaydeder ve herkese açık URL'ini döner.
	Save(fh *multipart.FileHeader) (url string, err error)
	// Delete, URL'i verilen fotoğrafı siler (bulunamazsa sessizce geçer).
	Delete(url string) error
}

// LocalPhotoStorage, fotoğrafları yerel diskte saklar ve /uploads/ altından sunar.
type LocalPhotoStorage struct {
	dir       string
	urlPrefix string
}

func NewLocalPhotoStorage(dir string) (*LocalPhotoStorage, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("uploads klasörü oluşturulamadı: %w", err)
	}
	return &LocalPhotoStorage{dir: dir, urlPrefix: "/uploads/"}, nil
}

// detectPhotoType, dosyanın ilk 512 baytından gerçek MIME türünü tespit eder.
func detectPhotoType(fh *multipart.FileHeader) (ext string, err error) {
	file, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("dosya okunamadı: %w", err)
	}

	mime := http.DetectContentType(buf[:n])
	ext, ok := allowedPhotoTypes[mime]
	if !ok {
		return "", fmt.Errorf("%w: %s (tespit edilen tip: %s)", ErrPhotoInvalidType, fh.Filename, mime)
	}
	return ext, nil
}

func (s *LocalPhotoStorage) Validate(fh *multipart.FileHeader) error {
	if fh.Size > MaxPhotoBytes {
		return fmt.Errorf("%w: %s (%d bayt)", ErrPhotoTooLarge, fh.Filename, fh.Size)
	}
	_, err := detectPhotoType(fh)
	return err
}

func (s *LocalPhotoStorage) Save(fh *multipart.FileHeader) (string, error) {
	ext, err := detectPhotoType(fh)
	if err != nil {
		return "", err
	}

	name, err := randomFilename(ext)
	if err != nil {
		return "", err
	}

	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.OpenFile(filepath.Join(s.dir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(dst, io.LimitReader(src, MaxPhotoBytes+1)); err != nil {
		dst.Close()
		_ = os.Remove(dst.Name())
		return "", err
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(dst.Name())
		return "", err
	}
	return s.urlPrefix + name, nil
}

func (s *LocalPhotoStorage) Delete(url string) error {
	if !strings.HasPrefix(url, s.urlPrefix) {
		return nil
	}
	// path.Base ile yalnızca dosya adı alınır; "../" ile klasör dışına çıkılamaz
	err := os.Remove(filepath.Join(s.dir, path.Base(url)))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func randomFilename(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b) + ext, nil
}
