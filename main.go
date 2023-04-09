package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

var lines chan string
var channel = flag.String("channel", "-1", "Channel to connect to")
var token = flag.String("token", "-1", "OAuth token to use")

func main() {
	flag.Parse()
	if *channel == "-1" {
		fmt.Println("Please specify a channel to connect to")
		os.Exit(1)
	}
	if *token == "-1" {
		fmt.Println("Please specify an OAuth token to use")
		os.Exit(1)
	}
	godotenv.Load()
	lines = make(chan string, 1000)
	server_endpoint := "irc.chat.twitch.tv:6667"
	nickname := "go-ttv"
	stream, err := net.Dial("tcp", server_endpoint)
	retires := 3
	if err != nil {
		for err != nil && retires > 0 {
			fmt.Println("Failed to connect to " + server_endpoint + ", retrying in 5 seconds")
			time.Sleep(5 * time.Second)
			stream, err = net.Dial("tcp", server_endpoint)
			retires--
		}
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Connected to " + server_endpoint)
	stream.Write([]byte("PASS " + *token + "\n"))
	stream.Write([]byte("NICK " + nickname + "\n"))
	stream.Write([]byte("JOIN " + "#" + *channel + "\n"))
	go processMessage()
	for {
		buf := make([]byte, 1024*8)
		size, err := stream.Read(buf)
		retires = 3
		if err != nil {
			for err != nil && retires > 0 {
				fmt.Println("Failed to read from " + server_endpoint + ", retrying in 5 seconds")
				time.Sleep(5 * time.Second)
				size, err = stream.Read(buf)
				retires--
			}
			if err != nil {
				panic(err)
			}
		}
		lines <- string(buf[:size])
	}
}

func processMessage() {
	timestampColor := color.New(color.FgBlack, color.BgWhite, color.Bold).SprintFunc()
	for {
		line := <-lines
		for _, line := range strings.Split(line, "\n") {
			now := time.Now()
			if line == "" {
				continue
			}
			firstExclamation := strings.Index(line, "!")
			if firstExclamation == -1 {
				continue
			}
			lastColon := strings.LastIndex(line, ":")
			if lastColon == -1 {
				continue
			}
			fmt.Printf(
				"%s %s %s\n",
				timestampColor(now.Format("15:04:05")),
				color.CyanString(line[1:firstExclamation]+":"),
				line[lastColon+1:],
			)
		}
	}
}
