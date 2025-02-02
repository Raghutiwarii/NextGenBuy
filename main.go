package main

import (
	"ecom/backend/controllers"
	"ecom/backend/database"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get environment variables
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8000" // Default to port 8000
	}

	// Database connection
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	database.DB = db // Ensure the global DB variable is set
	log.Println("Successfully connected to the database!")

	// Initialize Gin router
	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// endpoint
	r.POST("/register", controllers.OnBoardingCustomer)
	r.POST("/login", controllers.Login)

	r.POST("/merchant/register", controllers.OnBoardingMerchant)

	// Secure routes with JWT authentication middleware
	// secured := r.Group("/")

	// secured.Use(middleware.AuthMiddleware())
	// secured.GET("/user/profile", controllers.GetUserProfile)

	// Display banner in logs
	banner := `
	  ,------.  ,-----.  ,-----.  ,--.   ,--. 
	  |  .---' '  .--./ '  .-.  ' |   '.'   | 
	  |  '--,  |  |     |  | |  | |  |'.'|  | 
	  |  '--.' '  '--'\ '  '-'  ' |  |   |  | 
	  '------'  '-----'  '-----'  '--'   '--' 
	`
	log.Println(banner)

	// Log the routes
	for _, route := range r.Routes() {
		fmt.Println(route.Method, route.Path)
	}

	// Start the server
	log.Printf("Server is running on port %s", serverPort)
	if err := r.Run(fmt.Sprintf(":%s", serverPort)); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
