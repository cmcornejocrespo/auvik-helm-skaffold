package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

var DEFAULT_LISTEN_PORT = 80

func hello(c echo.Context) error {
	return c.String(http.StatusOK, fmt.Sprintf("Message: %s", os.Getenv("MESSAGE")))
}

func main() {
	var listenPort int
	listenPortEnv, ok := os.LookupEnv("LISTEN_PORT")
	if ok {
		if parsedPort, err := strconv.ParseInt(listenPortEnv, 10, 32); err != nil {
			log.Fatalf("Invalid listen port: %d\n", listenPort)
		} else {
			listenPort = int(parsedPort)
		}
	} else {
		listenPort = DEFAULT_LISTEN_PORT
	}

	e := echo.New()

	e.GET("/", hello)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", listenPort)))
}
