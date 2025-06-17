package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	copyright_controller "biblebrain-services/cmd/httpserver/api/copyright/controller"
	util "biblebrain-services/util"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

func setupRouter() *ginadapter.GinLambda {
	// Set up logger
	logger := util.Logger(os.Getenv("LOG_LEVEL"))
	slog.SetDefault(logger)
	slog.Info("Initializing router")

	// Build Gin engine and routes
	gengine := gin.Default()
	api := gengine.Group("/api")
	{
		api.GET("/copyright", copyright_controller.Get)
	}

	gengine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "PAGE_NOT_FOUND",
			"message": "Page not found",
		})
	})

	return ginadapter.New(gengine)
}

func main() {
	// Create the adapter exactly once, in main()
	ginLambda := setupRouter()

	// Start Lambda with a closure that captures our adapter
	lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		slog.Info("Handling request", "path", req.Path, "params", req.PathParameters)

		return ginLambda.ProxyWithContext(ctx, req)
	})
}
