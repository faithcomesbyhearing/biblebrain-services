package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Get(gctx *gin.Context) {
	gctx.JSON(http.StatusOK, "AWS Lambda ops biblebrain-service is running!")
}
