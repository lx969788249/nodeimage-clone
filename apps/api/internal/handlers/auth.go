package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nodeimage/api/internal/models"
	"nodeimage/api/internal/security"
	"nodeimage/api/internal/service"
)

type registerRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"displayName" binding:"required"`
	DeviceName  string `json:"deviceName"`
}

type authResponse struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	DeviceID     string        `json:"deviceId"`
	User         userResponse  `json:"user"`
}

type userResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	Status      string `json:"status"`
}

func (h HandlerSet) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.authService.Register(c.Request.Context(), service.RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sendAuthResponse(c, result)
}

type loginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
}

func (h HandlerSet) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.authService.Login(c.Request.Context(), service.LoginInput{
		Email:      req.Email,
		Password:   req.Password,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	})
	if err != nil {
		status := http.StatusUnauthorized
		if strings.Contains(err.Error(), "suspended") {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	sendAuthResponse(c, result)
}

type refreshRequest struct {
	UserID       string `json:"userId" binding:"required"`
	DeviceID     string `json:"deviceId" binding:"required"`
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (h HandlerSet) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.authService.Refresh(c.Request.Context(), service.RefreshInput{
		UserID:       req.UserID,
		DeviceID:     req.DeviceID,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	sendAuthResponse(c, result)
}

type logoutRequest struct {
	UserID   string `json:"userId" binding:"required"`
	DeviceID string `json:"deviceId" binding:"required"`
}

func (h HandlerSet) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.UserID, req.DeviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func sendAuthResponse(c *gin.Context, result service.AuthResult) {
	resp := authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		DeviceID:     result.DeviceID,
		User: userResponse{
			ID:          result.User.ID,
			Email:       result.User.Email,
			DisplayName: result.User.DisplayName,
			Role:        string(result.User.Role),
			Status:      string(result.User.Status),
		},
	}

	c.JSON(http.StatusOK, resp)
}

func (h HandlerSet) Me(c *gin.Context) {
	userVal, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, ok := userVal.(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_user"})
		return
	}

	resp := userResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
		Status:      string(user.Status),
	}

	c.JSON(http.StatusOK, gin.H{
		"user": resp,
	})
}

type sessionResponse struct {
	ID         string    `json:"id"`
	DeviceID   string    `json:"deviceId"`
	DeviceName string    `json:"deviceName"`
	IPAddress  string    `json:"ipAddress"`
	UserAgent  string    `json:"userAgent"`
	LastSeenAt time.Time `json:"lastSeenAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Current    bool      `json:"current"`
}

func (h HandlerSet) ListSessions(c *gin.Context) {
	userVal, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, ok := userVal.(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_user"})
		return
	}

	claimsVal, _ := c.Get("access_claims")
	claims, ok := claimsVal.(security.AccessClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_claims"})
		return
	}

	sessions, err := h.sessions.ListByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]sessionResponse, 0, len(sessions))
	for _, session := range sessions {
		resp = append(resp, sessionResponse{
			ID:         session.ID,
			DeviceID:   session.DeviceID,
			DeviceName: session.DeviceName,
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			LastSeenAt: session.LastSeenAt,
			ExpiresAt:  session.ExpiresAt,
			Current:    session.ID == claims.SessionID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": resp,
	})
}

func (h HandlerSet) RevokeSession(c *gin.Context) {
	userVal, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, ok := userVal.(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_user"})
		return
	}

	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deviceId required"})
		return
	}

	claimsVal, _ := c.Get("access_claims")
	claims, ok := claimsVal.(security.AccessClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_claims"})
		return
	}
	if claims.DeviceID == deviceID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot_revoke_current_device"})
		return
	}

	if err := h.sessions.DeleteByDevice(c.Request.Context(), user.ID, deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
