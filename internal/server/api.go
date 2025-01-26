package server

import (
	"LiteCanary/internal/models"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Login endpoint
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(ctx *gin.Context) {
	loginRequest := &LoginRequest{}
	if bindJSON(ctx, loginRequest) {
		return
	}
	session, err := localOpts.Opts.Commander.Login(loginRequest.Username, loginRequest.Password)
	if generalError(ctx, err, http.StatusUnauthorized) {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"session": session})
}

// Password reset endpoint
type ResetPasswordRequest struct {
	Password string `json:"password"`
}

func ResetPassword(ctx *gin.Context) {
	resetPasswordRequest := &ResetPasswordRequest{}
	if bindJSON(ctx, resetPasswordRequest) {
		return
	}

	token := getToken(ctx)

	err := localOpts.Opts.Commander.ResetPassword(token, resetPasswordRequest.Password)
	if generalError(ctx, err, http.StatusInternalServerError) {
		return
	}

	ctx.Status(http.StatusOK)
}

// Registration endpoint
type RegistrationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Register(ctx *gin.Context) {
	registrationRequest := &RegistrationRequest{}
	if bindJSON(ctx, registrationRequest) {
		return
	}

	err := localOpts.Opts.Commander.AddUser(registrationRequest.Username, registrationRequest.Password)
	if generalError(ctx, err, http.StatusInternalServerError) {
		return
	}

	ctx.Status(http.StatusOK)
}

// Update canary endpoint
type UpdateCanaryRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Redirect string `json:"redirect"`
	Id       string `json:"id"`
}

func UpdateCanary(ctx *gin.Context) {
	updateCanaryRequest := &UpdateCanaryRequest{}
	if bindJSON(ctx, updateCanaryRequest) {
		return
	}
	token := getToken(ctx)

	err := localOpts.Opts.Commander.UpdateCanary(token, updateCanaryRequest.Id, updateCanaryRequest.Name, updateCanaryRequest.Type, updateCanaryRequest.Redirect)
	if generalError(ctx, err, http.StatusBadRequest) {
		return
	}

	ctx.Status(http.StatusOK)
}

// New canary endpoint
type NewCanaryRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Redirect string `json:"redirect"`
}

func NewCanary(ctx *gin.Context) {
	newCanaryRequest := &NewCanaryRequest{}
	if bindJSON(ctx, newCanaryRequest) {
		return
	}
	token := getToken(ctx)

	canary, err := localOpts.Opts.Commander.AddCanary(token, newCanaryRequest.Name, newCanaryRequest.Type, newCanaryRequest.Redirect)
	if generalError(ctx, err, http.StatusBadRequest) {
		return
	}

	ctx.JSON(200, gin.H{
		"Id":        canary.Id,
		"Name":      canary.Name,
		"CreatedAt": canary.CreatedAt,
		"Type":      canary.Type,
	})
}

// Get all existing canaries and their alert history
func GetCanaries(ctx *gin.Context) {
	token := getToken(ctx)

	canaries, err := localOpts.Opts.Commander.GetCanaries(token)
	if generalError(ctx, err, http.StatusBadRequest) {
		return
	}

	var filteredCanaries []models.FilteredCanary
	for _, canary := range canaries {
		triggerHistory, err := localOpts.Opts.Commander.GetTriggerHistory(token, canary.Id)
		if generalError(ctx, err, http.StatusInternalServerError) {
			return
		}
		filteredCanaries = append(filteredCanaries, filterCanary(&canary, triggerHistory))
	}

	ctx.JSON(200, gin.H{
		"canaries": filteredCanaries,
	})
}

// Wipe canary
func WipeCanary(ctx *gin.Context) {
	token := getToken(ctx)
	canaryId := ctx.Param("id")
	if !localOpts.Opts.Commander.ValidCanary(canaryId) {
		generalError(ctx, errors.New("id not found"), http.StatusNotFound)
		return
	}
	err := localOpts.Opts.Commander.WipeCanary(token, canaryId)
	if generalError(ctx, err, http.StatusBadRequest) {
		return
	}
}

// Delete a canary
func DeleteCanary(ctx *gin.Context) {
	token := getToken(ctx)
	canaryId := ctx.Param("id")
	if !localOpts.Opts.Commander.ValidCanary(canaryId) {
		generalError(ctx, errors.New("id not found"), http.StatusNotFound)
		return
	}
	err := localOpts.Opts.Commander.DeleteCanaryByID(token, canaryId)
	if generalError(ctx, err, http.StatusBadRequest) {
		return
	}
}

// Delete user
func DeleteUser(ctx *gin.Context) {
	token := getToken(ctx)

	err := localOpts.Opts.Commander.DeleteUser(token)
	if generalError(ctx, err, http.StatusBadRequest) {
		return
	}
}

// Trigger a canary
func TriggerCanary(ctx *gin.Context) {
	canaryId := ctx.Param("id")
	canary, err := localOpts.Opts.Commander.GetCanary(canaryId)
	if err != nil {
		generalError(ctx, errors.New("id not found"), http.StatusNotFound)
		return
	}

	err = localOpts.Opts.Commander.TriggerCanary(&models.TriggerEvent{
		Canaryid:         canaryId,
		Ip:               ctx.RemoteIP(),
		Useragent:        ctx.Request.UserAgent(),
		Timestamp:        time.Now(),
		Keyboardlanguage: ctx.Request.Header.Get("Accept-Language"),
	})
	if generalError(ctx, err, http.StatusBadRequest) {
		return
	}
	if !validType(canary.Type) {
		generalError(ctx, errors.New("invalid type"), http.StatusBadRequest)
		return
	}

	switch canary.Type {
	case "redirect":
		ctx.Redirect(301, canary.Redirect)
	case "text":
		ctx.String(200, "This is a test page.")
	case "image":
		pixel, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+P+/HgAFhAJ/wlseKgAAAABJRU5ErkJggg==")
		ctx.Data(http.StatusOK, "image/png", pixel)
	}
}

// Helpers
func validType(canaryType string) bool {
	accepted := []string{"image", "text", "redirect"}
	for _, accept := range accepted {
		if accept == canaryType {
			return true
		}
	}

	return false
}

func convertHistory(history *[]models.TriggerEvent) *[]models.LocalTriggerEvent {
	var convertedHistory []models.LocalTriggerEvent
	for _, event := range *history {
		convertedHistory = append(convertedHistory, models.LocalTriggerEvent{
			Timestamp:        event.Timestamp,
			Useragent:        event.Useragent,
			Keyboardlanguage: event.Keyboardlanguage,
			Ip:               event.Ip,
		})
	}
	return &convertedHistory
}

func filterCanary(canary *models.Canary, history *[]models.TriggerEvent) models.FilteredCanary {
	return models.FilteredCanary{
		Name:    canary.Name,
		Id:      canary.Id,
		Type:    canary.Type,
		History: convertHistory(history),
	}
}

func getToken(ctx *gin.Context) string {
	bearerSections := strings.Split(ctx.Request.Header.Get("Authorization"), " ")
	if len(bearerSections) != 2 {
		return ""
	}
	return bearerSections[1]
}

func generalError(ctx *gin.Context, err error, statusCode int) bool {
	if err == nil {
		return false
	}
	if localOpts.Opts.Debug {
		log.Printf("warning in %s: %s\n", ctx.Request.URL.Path, err.Error())
	}
	ctx.AbortWithStatus(statusCode)
	return true
}

func bindJSON(ctx *gin.Context, obj any) bool {
	if err := ctx.BindJSON(obj); err != nil {
		generalError(ctx, err, http.StatusInternalServerError)
		return true
	}
	return false
}
