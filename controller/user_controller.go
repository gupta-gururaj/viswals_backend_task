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

func (c *Controller) GetAllUsers(ctx *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	users, err := c.UserService.GetAllUsers(ctxWithTimeout)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed to get all users"})
	}
	return ctx.Status(fiber.StatusOK).JSON(users)
}

func (c *Controller) GetAllUsersSSE(ctx *fiber.Ctx) error {
	limit := int64(10)
	offset := int64(1)
	var err error

	if ctx.Query("limit") != "" {
		limit, err = strconv.ParseInt(ctx.Query("limit"), 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid limit parameter"})
		}
	}

	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Access-Control-Allow-Origin", "*")

	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		var isLastData bool
		for {
			data, err := c.UserService.GetAllUsersSSE(context.Background(), limit, offset*limit)
			if err != nil {
				if errors.Is(err, database.ErrNoData) {
					isLastData = true
				}
			}

			_, err = fmt.Fprintf(w, "data: %s\n\n", string(data))
			if err != nil {
				// c.logger.Error("failed to write response", zap.Error(err))
				return
			}
			w.Flush() // Ensures data is sent to the client
			time.Sleep(2 * time.Second)

			if isLastData {
				break
			}
			offset++
		}

		_, _ = fmt.Fprint(w, "data: END\n\n")
		w.Flush()
	})

	return nil
}


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

func (c *Controller) CreateUser(ctx *fiber.Ctx) error {
	var user models.UserDetails

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

