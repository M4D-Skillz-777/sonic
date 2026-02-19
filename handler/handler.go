package handler

import (
	"net/http"

	"audio-fingerprinting/config"
	"audio-fingerprinting/fingerprint"
	"audio-fingerprinting/matcher"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	matcher *matcher.Matcher
	config  *config.Config
	rdb     *redis.Client
}

func NewHandler(m *matcher.Matcher, cfg *config.Config, rdb *redis.Client) *Handler {
	return &Handler{matcher: m, config: cfg, rdb: rdb}
}

func SetupRoutes(r *gin.Engine, m *matcher.Matcher, cfg *config.Config) {
	h := NewHandler(m, cfg, nil)

	r.GET("/health", h.Health)
	r.POST("/fingerprint", h.Fingerprint)
	r.POST("/fingerprint/custom", h.FingerprintCustomFFT)
	r.POST("/recognize", h.Recognize)
	r.POST("/recognize/custom", h.RecognizeCustomFFT)
	r.DELETE("/fingerprint/:name", h.DeleteSong)
	r.GET("/songs", h.ListSongs)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (h *Handler) Fingerprint(c *gin.Context) {
	songName := c.PostForm("name")
	if songName == "" {
		songName = c.Query("name")
	}
	if songName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "song name is required"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	if header.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty file"})
		return
	}

	fp, err := fingerprint.DecodeAndFingerprint(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.matcher.StoreFingerprint(c.Request.Context(), songName, fp.Hashes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"song_name":   songName,
		"hashes":      len(fp.Hashes),
		"fft_impl":    "gofft",
		"fft_details": "Optimized in-place radix-2 FFT from gofft library",
	})
}

func (h *Handler) Recognize(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	if header.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty file"})
		return
	}

	fp, err := fingerprint.DecodeAndFingerprint(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	songName, similarity, err := h.matcher.FindMatch(c.Request.Context(), fp.Hashes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if songName == "" {
		c.JSON(http.StatusOK, gin.H{
			"song_name":  "",
			"confidence": 0,
			"message":    "no match found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"song_name":  songName,
		"confidence": similarity,
	})
}

func (h *Handler) DeleteSong(c *gin.Context) {
	songName := c.Param("name")
	if songName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "song name is required"})
		return
	}

	if err := h.matcher.DeleteSong(c.Request.Context(), songName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"song_name": songName,
		"deleted":   true,
	})
}

func (h *Handler) ListSongs(c *gin.Context) {
	songs, err := h.matcher.ListSongs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if songs == nil {
		songs = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"songs": songs,
	})
}

func (h *Handler) FingerprintCustomFFT(c *gin.Context) {
	songName := c.PostForm("name")
	if songName == "" {
		songName = c.Query("name")
	}
	if songName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "song name is required"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	if header.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty file"})
		return
	}

	fp, err := fingerprint.DecodeAndFingerprintCustomFFT(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.matcher.StoreFingerprint(c.Request.Context(), songName, fp.Hashes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"song_name":   songName,
		"hashes":      len(fp.Hashes),
		"fft_impl":    "custom_cooley_tukey",
		"fft_details": "In-place radix-2 Cooley-Tukey FFT algorithm",
	})
}

func (h *Handler) RecognizeCustomFFT(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	if header.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty file"})
		return
	}

	fp, err := fingerprint.DecodeAndFingerprintCustomFFT(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	songName, similarity, err := h.matcher.FindMatch(c.Request.Context(), fp.Hashes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if songName == "" {
		c.JSON(http.StatusOK, gin.H{
			"song_name":  "",
			"confidence": 0,
			"message":    "no match found",
			"fft_impl":   "custom_cooley_tukey",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"song_name":  songName,
		"confidence": similarity,
		"fft_impl":   "custom_cooley_tukey",
	})
}
