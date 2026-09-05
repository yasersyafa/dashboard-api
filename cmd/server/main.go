package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	finance "github.com/yasersyafa/dashboard-api/internal/finance/category"
	"github.com/yasersyafa/dashboard-api/internal/gen"
	"github.com/yasersyafa/dashboard-api/internal/task"
)

type API struct {
	*task.Handler
	*finance.CategoryHandler
}

func main() {
	ctx := context.Background()

	connectionString := os.Getenv("DABATASE_URL")
	if connectionString == "" {
		connectionString = "postgres://dashboard:dashboard@localhost:5434/dashboard" // example url
	}

	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		log.Fatalf("unable to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("unable to ping database: %v", err)
	}
	log.Println("database connected")

	// tasks
	taskRepo := task.NewPostgresRepository(pool)
	taskService := task.NewService(taskRepo)
	taskHandler := task.NewHandler(taskService)

	// finance - Category
	categoryRepo := finance.NewPostgresCategoryRepository(pool)
	categoryService := finance.NewCategoryService(categoryRepo)
	categoryHandler := finance.NewCategoryHandler(categoryService)

	api := &API {
		Handler: taskHandler,
		CategoryHandler: categoryHandler,
	}

	router := gin.Default()

	router.StaticFile("/docs", "./docs.html")
	router.StaticFile("/api/openapi.yaml", "./api/openapi.yaml")

	gen.RegisterHandlers(router, api)

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"data:": "Server is running",
		})
	})

	router.Run(":4321")
}