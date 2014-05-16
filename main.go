package main

// Main entry point for the app. Handles command-line options and starts the web listener

import (
	"flag"
	"fmt"
	"os"
)

var (
	httpPort    int
	botUsername string

	apiKey       string
	webhookToken string
)

func main() {
	// Parse command-line options
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: ./gopherbot -apiKey=xoxp-xxxx -webhookToken=xxxx -port=8002\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&apiKey, "apiKey", "", "Your Slack API key")
	flag.StringVar(&webhookToken, "webhookToken", "", "Your incoming webhook token")
	flag.IntVar(&httpPort, "port", 8002, "The HTTP port on which to listen")
	flag.StringVar(&botUsername, "botUsername", "gopherbot", "The name of the bot when it speaks")

	flag.Parse()

	if httpPort == 0 || apiKey == "" || webhookToken == "" {
		flag.Usage()
		os.Exit(2)
	}

	// Start the webserver
	StartServer(httpPort)
}
