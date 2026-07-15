package attachments

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"

	"agentmanager/modules/apperr"
	"agentmanager/modules/crypto"
	"agentmanager/modules/encryption"
)

func (s *Service) Upload(c echo.Context) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return apperr.JSON(c, http.StatusBadRequest, ErrUpload)
	}
	if fh.Size > s.cfg.MaxSize {
		return apperr.JSON(c, http.StatusRequestEntityTooLarge, ErrTooLarge)
	}
	src, err := fh.Open()
	if err != nil {
		return apperr.JSON(c, http.StatusBadRequest, ErrUpload)
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return apperr.Internal(c, err)
	}
	enc, err := encryption.EncryptBytes(data, s.cfg.MasterKey)
	if err != nil {
		return apperr.Internal(c, err)
	}
	id := crypto.RandomHex(16)
	if err := os.WriteFile(filepath.Join(s.cfg.StoragePath, id+".enc"), enc, 0o600); err != nil {
		return apperr.Internal(c, err)
	}

	ct := fh.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	rec := Record{ID: id, Filename: fh.Filename, ContentType: ct, Size: fh.Size, CreatedAt: time.Now().Unix()}
	if err := s.insert(rec); err != nil {
		return apperr.Internal(c, err)
	}
	return c.JSON(http.StatusOK, echo.Map{"id": id, "filename": rec.Filename, "content_type": ct, "size": rec.Size})
}

func (s *Service) Download(c echo.Context) error {
	id := c.Param("id")
	rec, ok := s.get(id)
	if !ok {
		return apperr.JSON(c, http.StatusNotFound, ErrNotFound)
	}
	enc, err := os.ReadFile(filepath.Join(s.cfg.StoragePath, id+".enc"))
	if err != nil {
		return apperr.JSON(c, http.StatusNotFound, ErrNotFound)
	}
	data, err := encryption.DecryptBytes(enc, s.cfg.MasterKey)
	if err != nil {
		return apperr.Internal(c, err)
	}
	c.Response().Header().Set("Content-Disposition", "inline; filename=\""+rec.Filename+"\"")
	return c.Blob(http.StatusOK, rec.ContentType, data)
}
