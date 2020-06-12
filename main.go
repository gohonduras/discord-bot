package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gohonduras/discord-bot/hackernews"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"mvdan.cc/xurls/v2"
)

var (
	log = logrus.WithField("prefix", "main")
)

// Message handler implements several functions which will
// be in charge of responding to discord messages the bot
// observes.
type messageHandler struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// The init function runs on package initialization, helping us setup
// some useful globals such as a logging formatter.
func init() {
	formatter := new(prefixed.TextFormatter)
	formatter.TimestampFormat = "2020-01-01 07:12:23"
	formatter.FullTimestamp = true
	logrus.SetFormatter(formatter)
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatalf("Expected DISCORD_TOKEN env var, provided none")
	}
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		log.Fatalf("Could not initialize discord session: %v", err)
	}

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	// We initialize a new context with a cancelation function, useful
	// for cleanup of every possible goroutine on SIGTERM.
	ctx, cancel := context.WithCancel(context.Background())
	handler := &messageHandler{
		ctx:    ctx,
		cancel: cancel,
	}

	// Go hacker news handler.
	dg.AddHandler(handler.hackerNewsHandler)
	dg.AddHandler(handler.hyperlinkCompilerHandler)

	// Wait here until SIGTERM or another interruption signal is received.
	log.Println("Bot is now running, press ctrl-c to exit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session and cancel the global
	// context gracefully.
	cancel()
	if err := dg.Close(); err != nil {
		log.Fatalf("Could not gracefully stop discord session: %v", err)
	}
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func (mh *messageHandler) hackerNewsHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself.
	if m.Author.ID == s.State.User.ID {
		return
	}
	commandPrefix := "!hackernews"
	if !strings.Contains(m.Content, commandPrefix) {
		return
	}
	searchQuery := strings.TrimSpace(m.Content[len(commandPrefix):])
	if searchQuery == "" {
		return
	}
	hnClient := hackernews.NewAPIClient()
	res, err := hnClient.Search(mh.ctx, searchQuery)
	if err != nil {
		log.Errorf("Could not search hacker news API: %v", err)
		return
	}
	if _, err := s.ChannelMessageSend(m.ChannelID, res.String()); err != nil {
		log.Errorf("Could not send message over channel: %v", err)
	}
}

var data = [][]string{{"Line1", "Hello Readers of"}, {"Line2", "golangcode.com"}}

func (mh *messageHandler) hyperlinkCompilerHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Check for command prefix
	commandPrefix := "!links"
	if !strings.Contains(m.Content, commandPrefix) {
		return
	}
	// Read last 100 channel messages (ChannelMessages hard limit 😢)
	messages, err := s.ChannelMessages(m.ChannelID, 100, "", "", "")
	links := make(map[string]bool)
	messageCount := 0
	linkCount := 0
	for _, message := range messages {
		urls := xurls.Strict().FindAllString(message.Content, -1)
		if len(urls) > 0 {
			for _, url := range urls {
				if _, ok := links[url]; ok {
					//log.Println("Duplicated link")
				} else {
					links[url] = true
					log.Println("Added link %v", url)
					linkCount++
				}

			}
		}
		messageCount++
	}
	log.Println("Recorded %d links from %d messages", linkCount, messageCount)

	if err != nil {
		log.Errorf("Could not search hacker news API: %v", err)
		return
	}

	compiledLinks := new(bytes.Buffer)
	for key, _ := range links {
		fmt.Fprintf(compiledLinks, "%s\n", key)
	}

	//Cant' send more than 2000 per message. TODO: Split long messages and send them accordingly.
	if len(compiledLinks.String()) > 2000 {
		if _, err := s.ChannelMessageSend(m.ChannelID, "That's a lot of links (length over 2000 characters)"); err != nil {
			log.Errorf("Could not send message over channel: %v", err)
		}
	} else {
		if _, err := s.ChannelMessageSend(m.ChannelID, compiledLinks.String()); err != nil {
			log.Errorf("Could not send message over channel: %v", err)
		}
	}

}
