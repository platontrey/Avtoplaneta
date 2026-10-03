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

// DownloadAppHandler отдает APK-файл мобильного приложения
func (h *Handler) DownloadAppHandler(c *gin.Context) {
	apkPaths := []string{
		"./downloads/avtoplaneta-release.apk",
		"./downloads/app-release.apk",
		"backend/downloads/avtoplaneta-release.apk",
		"backend/downloads/app-release.apk",
		"/app/downloads/avtoplaneta-release.apk",
		"/app/downloads/app-release.apk",
		"../downloads/avtoplaneta-release.apk",
		"../downloads/app-release.apk",
	}

	for _, p := range apkPaths {
		if _, err := os.Stat(p); err == nil {
			c.Header("Content-Disposition", "attachment; filename=\"avtoplaneta-release.apk\"")
			c.Header("Content-Type", "application/vnd.android.package-archive")
			c.File(p)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "файл обновления APK не найден на сервере"})
}
