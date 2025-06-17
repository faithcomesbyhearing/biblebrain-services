package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	connection_service "biblebrain-services/service/connection"
	copyright_service "biblebrain-services/service/copyright"

	"github.com/gin-gonic/gin"
)

var ErrProductsRequired = errors.New("products are required")

type CopyrightRequest struct {
	// Add fields as needed for the request
	Products []copyright_service.Product `binding:"required" json:"productList"`
}

func (c *CopyrightRequest) Validate() error {
	// Implement validation logic if needed
	if len(c.Products) == 0 {
		return ErrProductsRequired
	}

	return nil
}

// POST /copyright/create.
func Create(gctx *gin.Context) {
	var req CopyrightRequest
	if err := gctx.ShouldBindJSON(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})

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
	copyrights, err := cser.GetCopyrightBy(ctx, packageRequest.ProductCodes())
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get copyrights: %v", err)})

		return
	}
	// If no copyrights are found, return a 404 error
	if len(copyrights) == 0 {
		gctx.JSON(http.StatusNotFound, gin.H{"error": "No copyrights found for the provided products"})

		return
	}

	pdf, err := cser.StreamCopyright(ctx, copyrights, true)
	if err != nil {
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
}
