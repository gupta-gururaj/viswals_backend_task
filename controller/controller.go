package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var (
	defaultTimeout  = 5 * time.Second
	defaultHttpPort = "8080"
)

type Controller struct {
	// UserService UserService
	logger      *zap.Logger
	HttpPort    string
}

// Option defines functional options for the controller
type Option func(*Controller)

// WithHttpPort sets a custom HTTP port for the Controller
func WithHttpPort(port string) Option {
	return func(c *Controller) {
		c.HttpPort = port
	}
}

// New creates a new Controller with optional configurations
func New(logger *zap.Logger, opts ...Option) *Controller {
	ctrl := &Controller{
		// UserService: userService,
		logger:      logger,
		HttpPort:    defaultHttpPort,
	}
	for _, opt := range opts {
		opt(ctrl)
	}
	return ctrl
}

// sendResponse sends a JSON response with status, message, and optional data
func (c *Controller) sendResponse(ctx *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return ctx.Status(statusCode).JSON(fiber.Map{
		"status_code": statusCode,
		"message":     message,
		"data":        data,
	})
}

// Start initializes and starts the Fiber server
func (c *Controller) Start() error {
	app := fiber.New(fiber.Config{
		ReadTimeout: defaultTimeout,
	})

	c.registerRoutes(app)

	addr := fmt.Sprintf(":%s", c.HttpPort)
	c.logger.Info("Starting Fiber server", zap.String("port", c.HttpPort))

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	return nil
}

// Ping is a health check endpoint to confirm the server is running
func (c *Controller) Ping(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"message": "pong",
	})
}

// registerRoutes sets up routes for HTTP endpoints
func (c *Controller) registerRoutes(app *fiber.App) {
	// app.Get("/users", c.GetAllUsers)
	// app.Get("/users/:id", c.GetUser)
	// app.Post("/users", c.CreateUser)
	// app.Delete("/users/:id", c.DeleteUser)
	// app.Get("/users/sse", c.GetAllUsersSSE)
	app.Static("/static", "./web")
}
