package controller

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/viswals_backend_task/pkg/models"
	database "github.com/viswals_backend_task/pkg/postgres"
	"go.uber.org/zap"
)

// GetAllUsers fetches all users from the service and returns them as a JSON response.
func (c *Controller) GetAllUsers(ctx *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// Extract query parameters
	name := ctx.Query("user_name", "")
	email := ctx.Query("email", "")
	
	users, err := c.UserService.GetAllUsers(ctxWithTimeout,name,email)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed to get all users"})
	}
	return ctx.Status(fiber.StatusOK).JSON(users)
}

// GetAllUsersSSE streams user data using Server-Sent Events (SSE).
func (c *Controller) GetAllUsersSSE(ctx *fiber.Ctx) error {
	limit := int64(10)  // Default limit for pagination
	offset := int64(1)
	var err error

	// Parse limit from query parameters
	if ctx.Query("limit") != "" {
		limit, err = strconv.ParseInt(ctx.Query("limit"), 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid limit parameter"})
		}
	}

	// Set necessary headers for SSE
	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Access-Control-Allow-Origin", "*")

	// Stream data to client using a buffered writer
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		var isLastData bool
		for {
			data, err := c.UserService.GetAllUsersSSE(context.Background(), limit, offset*limit)
			if err != nil {
				if errors.Is(err, database.ErrNoData) {
					isLastData = true // No more data available
				}
			}

			_, err = fmt.Fprintf(w, "data: %s\n\n", string(data))
			if err != nil {
				// Failed to write response, possibly due to client disconnecting
				return
			}
			w.Flush() // Send data to client immediately
			time.Sleep(2 * time.Second) // Prevent flooding the client

			if isLastData {
				break
			}
			offset++
		}

		// Indicate the end of the stream
		_, _ = fmt.Fprint(w, "data: END\n\n")
		w.Flush()
	})

	return nil
}

// GetUser retrieves a user by ID from the service.
func (c *Controller) GetUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "user id is not provided or empty"})
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	user, err := c.UserService.GetUser(ctxWithTimeout, id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return ctx.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{"message": "deadline exceeded, please try again later"})
		}
		if errors.Is(err, database.ErrNoData) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
		}
		c.logger.Error("failed to get user", zap.Error(err), zap.String("id", id))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(user)
}

// CreateUser parses the request body and creates a new user in the database.
func (c *Controller) CreateUser(ctx *fiber.Ctx) error {
	var user models.UserDetails

	// Parse request body into user struct
	if err := ctx.BodyParser(&user); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "failed to parse request body"})
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	err := c.UserService.CreateUser(ctxWithTimeout, &user)
	if err != nil {
		if errors.Is(err, database.ErrDuplicate) {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "user already exists"})
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return ctx.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{"message": "request timeout, please try again later"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "internal server error"})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "user created successfully"})
}

// DeleteUser removes a user by ID from the database.
func (c *Controller) DeleteUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "user id is required"})
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	err := c.UserService.DeleteUser(ctxWithTimeout, id)
	if err != nil {
		if errors.Is(err, database.ErrNoData) {
			return ctx.Status(fiber.StatusNoContent).JSON(fiber.Map{"message": "user not found or already deleted"})
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return ctx.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{"message": "request timeout, please try again later"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "internal server error"})
	}

	return ctx.Status(fiber.StatusNoContent).Send(nil)
}
