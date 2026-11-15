package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Drom integration
func getDromDialogs(c *gin.Context) {
	dialogs, err := svc.GetDromBriefs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"dialogs": dialogs,
	})
}

func getDromMessages(c *gin.Context) {
	dialogID := c.Query("dialog_id")
	if dialogID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dialog_id parameter is required"})
		return
	}
	messages, dialogMeta, err := svc.GetDromMessages(c.Request.Context(), dialogID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"dialog":   dialogMeta,
	})
}

func sendDromMessage(c *gin.Context) {
	var req struct {
		DialogID string `json:"dialog_id"`
		Content  string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.DialogID == "" || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dialog_id and content required"})
		return
	}

	msg, err := svc.SendDromMessage(c.Request.Context(), req.DialogID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

func getDromAuthStatus(c *gin.Context) {
	status, err := svc.GetDromAuthStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

func loginDrom(c *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		Code     string `json:"code,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if strings.TrimSpace(req.Login) == "" || strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Логин и пароль обязательны"})
		return
	}

	res, err := svc.LoginDrom(c.Request.Context(), req.Login, req.Password, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func saveDromSession(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пустые данные сессии"})
		return
	}

	if err := svc.SaveDromSession(c.Request.Context(), body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Сессия успешно сохранена",
	})
}
