package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GetAppVersionHandler возвращает конфигурацию версий для мобильного приложения
func (h *Handler) GetAppVersionHandler(c *gin.Context) {
	configPaths := []string{
		"./config/app_version.json",
		"backend/config/app_version.json",
		"/app/config/app_version.json",
		"../config/app_version.json",
	}

	for _, p := range configPaths {
		if data, err := os.ReadFile(p); err == nil {
			c.Data(http.StatusOK, "application/json; charset=utf-8", data)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "файл конфигурации версий не найден"})
}

// DownloadAppHandler отдает APK-файл мобильного приложения с поддержкой архитектур (arm64, armv7)
func (h *Handler) DownloadAppHandler(c *gin.Context) {
	abi := c.DefaultQuery("abi", "")
	var targetFiles []string

	if abi == "arm64" || abi == "arm64-v8a" {
		targetFiles = []string{
			"avtoplaneta-arm64-v8a.apk",
			"avtoplaneta-latest-arm64.apk",
			"avtoplaneta-arm64.apk",
		}
	} else if abi == "armv7" || abi == "armeabi-v7a" {
		targetFiles = []string{
			"avtoplaneta-armeabi-v7a.apk",
			"avtoplaneta-latest-armv7.apk",
			"avtoplaneta-armv7.apk",
		}
	}

	// Общие фоллбэки
	targetFiles = append(targetFiles,
		"avtoplaneta-release.apk",
		"avtoplaneta-latest.apk",
		"app-release.apk",
	)

	baseDirs := []string{
		"./downloads",
		"backend/downloads",
		"/app/downloads",
		"../downloads",
	}

	for _, dir := range baseDirs {
		for _, name := range targetFiles {
			p := dir + "/" + name
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				c.Header("Content-Disposition", "attachment; filename=\""+name+"\"")
				c.Header("Content-Type", "application/vnd.android.package-archive")
				c.Header("Content-Encoding", "identity")
				c.Header("Cache-Control", "no-transform, public, max-age=3600")
				c.Header("Accept-Ranges", "bytes")
				c.File(p)
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "файл обновления APK не найден на сервере"})
}
