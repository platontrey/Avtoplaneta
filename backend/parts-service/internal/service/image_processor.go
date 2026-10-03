package service

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

const (
	// MaxUploadFileSize - максимальный размер загружаемого исходного файла (30 МБ)
	MaxUploadFileSize = 30 * 1024 * 1024
	// MaxImageDimension - максимальный размер по длинной стороне (2048px)
	MaxImageDimension = 2048
	// JPEGCompressionQuality - качество сжатия JPEG (85%)
	JPEGCompressionQuality = 85
)

// imageProcessingSem ограничивает одновременную обработку тяжелых изображений (максимум 2 одновременно),
// чтобы предотвратить всплески потребления RAM на сервере при массовой загрузке.
var imageProcessingSem = make(chan struct{}, 2)

// OptimizeAndSaveImage декодирует изображение из src, применяет авто-ориентацию по EXIF,
// пропорционально масштабирует до MaxImageDimension (если исходник крупнее), накладывает
// на белый фон (при наличии прозрачности) и сохраняет как качественный Progressive JPEG.
func OptimizeAndSaveImage(src io.Reader, destinationPath string) error {
	imageProcessingSem <- struct{}{}
	defer func() { <-imageProcessingSem }()

	// 1. Декодируем изображение с авто-ориентацией по EXIF
	img, err := imaging.Decode(src, imaging.AutoOrientation(true))
	if err != nil {
		return fmt.Errorf("ошибка декодирования изображения: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 2. Масштабируем до MaxImageDimension, если исходник больше
	var processedImg image.Image = img
	if width > MaxImageDimension || height > MaxImageDimension {
		processedImg = imaging.Fit(img, MaxImageDimension, MaxImageDimension, imaging.Lanczos)
	}

	// 3. Если в исходнике была прозрачность (PNG / WebP), накладываем на белый фон
	finalImg := compositeOverWhite(processedImg)

	// 4. Убеждаемся, что целевая папка существует
	dir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию %s: %w", dir, err)
	}

	// 5. Записываем во временный файл и атомарно подменяем
	tmpFile, err := os.CreateTemp(dir, "upload_opt_*.jpg")
	if err != nil {
		return fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	jpegOptions := &jpeg.Options{Quality: JPEGCompressionQuality}
	if err := jpeg.Encode(tmpFile, finalImg, jpegOptions); err != nil {
		tmpFile.Close()
		return fmt.Errorf("ошибка кодирования JPEG: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия файла: %w", err)
	}

	// Перемещаем готовый файл на целевой путь
	if err := os.Rename(tmpPath, destinationPath); err != nil {
		data, readErr := os.ReadFile(tmpPath)
		if readErr != nil {
			return fmt.Errorf("ошибка перемещения файла: %w", err)
		}
		if writeErr := os.WriteFile(destinationPath, data, 0644); writeErr != nil {
			return fmt.Errorf("ошибка записи целевого файла: %w", writeErr)
		}
	}

	return nil
}

// OptimizeAndSaveImageBytes обрабатывает срез байтов (для gRPC streaming upload)
func OptimizeAndSaveImageBytes(data []byte, destinationPath string) error {
	return OptimizeAndSaveImage(bytes.NewReader(data), destinationPath)
}

// compositeOverWhite накладывает изображение на белый холст с альфа-блендингом для сохранения в JPEG без черных артефактов
func compositeOverWhite(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	bg := imaging.New(w, h, color.White)
	return imaging.Overlay(bg, src, image.Pt(0, 0), 1.0)
}
