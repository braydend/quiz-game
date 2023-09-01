package main

import (
	"os"

	"github.com/braydend/quiz-game/src/game"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}

	return port
}

func main() {
	app := fiber.New()
	games := make(map[string]game.Game)

	app.Static("/", "./web/dist/")

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	/*
		Joins connection pool
		TODO: put a room/game ID in the url as a param
	*/
	app.Get("/ws/join/:id", websocket.New(func(c *websocket.Conn) {
		id := c.Params("id")

		game, ok := games[id]

		if ok {
			game.AddClient(c)
		}
	}))

	app.Get("/games/create/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		games[id] = game.NewGame()
		return nil
	})

	app.Get("/games", func(c *fiber.Ctx) error {
		return c.JSON(games)
	})

	// Check whether a game exists
	app.Get("/games/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		_, ok := games[id]
		type resp struct {
			Exists bool `json:"exists"`
		}
		return c.JSON(resp{Exists: ok})
	})

	app.Listen(getPort())
}
