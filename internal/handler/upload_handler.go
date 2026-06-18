package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"ctm/pkg/i18n"
	"ctm/pkg/storage"
)

const maxUploadSize = 10 << 20 // 10 MB

var allowedExtensions = map[string]bool{
	// Images
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".webp": true, ".heic": true, ".heif": true, ".bmp": true,
	".tiff": true, ".tif": true, ".svg": true,
	// Documents
	".pdf": true, ".doc": true, ".docx": true,
	".xls": true, ".xlsx": true, ".csv": true, ".txt": true,
	".odt": true, ".ods": true, ".odp": true, ".rtf": true,
}

type UploadHandler struct {
	store storage.Storage
	base  string // base URL prefix for returned URLs
}

func NewUploadHandler(store storage.Storage, baseURL string) *UploadHandler {
	return &UploadHandler{store: store, base: strings.TrimRight(baseURL, "/")}
}

func (h *UploadHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	rg.POST("/upload", auth, h.Upload)
}

func (h *UploadHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "too large") {
			BadRequest(c, "FILE_TOO_LARGE", i18n.T(c, "FILE_TOO_LARGE"))
			return
		}
		BadRequest(c, "VALIDATION_ERROR", i18n.T(c, "VALIDATION_ERROR"))
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		BadRequest(c, "FILE_TYPE_NOT_ALLOWED", i18n.T(c, "FILE_TYPE_NOT_ALLOWED"))
		return
	}

	// determine subdirectory from optional "dir" form field (default: "uploads")
	dir := c.PostForm("dir")
	if dir == "" {
		dir = "uploads"
	}

	relPath, err := h.store.Save(dir, header.Filename, file)
	if err != nil {
		InternalError(c)
		return
	}

	url := h.base + relPath
	OKMsg(c, gin.H{"url": url, "name": header.Filename, "size": header.Size}, "FILE_UPLOADED")
}
