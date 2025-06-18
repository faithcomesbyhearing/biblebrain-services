package controller

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	connection_service "biblebrain-services/service/connection"
	copyright_service "biblebrain-services/service/copyright"

	"github.com/gin-gonic/gin"
)

const (
	FormatPDF  = "pdf"
	FormatJSON = "json"
	ModeAudio  = "audio"
	ModeVideo  = "video"
	ModeText   = "text"
)

// Errors for validation.
var ErrProductsRequired = errors.New("products are required")

var ErrInvalidMode = errors.New("invalid mode")

var ErrInvalidFormat = errors.New("invalid format")

type CopyrightRequest struct {
	// Add fields as needed for the request
	Products []string `binding:"required"  form:"productCode"`
	Format   string   `binding:"omitempty" form:"format"`
	Mode     string   `binding:"omitempty" form:"mode"`
}

func (c *CopyrightRequest) Validate() error {
	// Implement validation logic if needed
	if len(c.Products) == 0 {
		return ErrProductsRequired
	}

	if c.Format != "json" && c.Format != "pdf" {
		return fmt.Errorf("%w: %q, only 'pdf' or 'json' is supported", ErrInvalidFormat, c.Format)
	}

	if c.Mode != "audio" && c.Mode != "video" && c.Mode != "text" {
		return fmt.Errorf("%w: %q, only 'audio', 'video', or 'text' are supported", ErrInvalidMode, c.Mode)
	}

	return nil
}

// GET api/copyright.
func Get(gctx *gin.Context) {
	var req CopyrightRequest
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request 2 data"})

		return
	}
	// Validate the request data if needed
	if err := req.Validate(); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	ctx := gctx.Request.Context()
	sqlCon := connection_service.GetBibleBrainDB(ctx)

	defer sqlCon.Close()

	cser := copyright_service.New(sqlCon)
	packageRequest := copyright_service.Package{
		Products: req.Products,
	}
	// Create the copyright PDF
	copyrights, err := cser.GetCopyrightBy(ctx, packageRequest.Products, req.Mode)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get copyrights: %v", err)})

		return
	}
	// If no copyrights are found, return a 404 error
	if len(copyrights) == 0 {
		gctx.JSON(http.StatusNotFound, gin.H{"error": "No copyrights found for the provided products"})

		return
	}

	switch req.Format {
	case FormatPDF:
		pdf, err := cser.StreamCopyright(ctx, copyrights, req.Mode)
		if err != nil {
			slog.Error("Failed to stream copyright PDF", "error", err)
			gctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

			return
		}

		defer pdf.Close()

		// Tell the client it’s a PDF
		gctx.Header("Content-Type", "application/pdf")
		// Optionally suggest a filename:
		gctx.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s.pdf"`, packageRequest.ID()))

		if _, err := io.Copy(gctx.Writer, pdf); err != nil {
			// This is a streaming error, so we return an error response
			gctx.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": fmt.Sprintf("streaming PDF failed: %v", err)},
			)

			return
		}
	case FormatJSON:
		// If the format is JSON, return the copyrights as JSON
		gctx.JSON(http.StatusOK, copyrights)
	default:
		gctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format specified"})

		return
	}
}
