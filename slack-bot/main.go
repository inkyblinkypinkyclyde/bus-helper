package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	apihelper "slack-bot/api-helper"
	responseformatter "slack-bot/response-formatter"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("loading .env: %v", err)
	}

	api := slack.New(
		os.Getenv("SLACK_BOT_TOKEN"),
		slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
	)

	client := socketmode.New(api)
	apiClient := apihelper.NewClient(os.Getenv("API_BASE_URL"))

	go func() {
		for evt := range client.Events {
			switch evt.Type {

			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					continue
				}

				client.Ack(*evt.Request)

				event := eventsAPIEvent.InnerEvent

				switch ev := event.Data.(type) {
				case *slackevents.MessageEvent:
					handleMessage(api, apiClient, ev)
				}
			}
		}
	}()

	log.Println("Bot running")
	client.Run()
}

func handleMessage(api *slack.Client, apiClient *apihelper.Client, event *slackevents.MessageEvent) {
	// Ignore messages from bots
	if event.BotID != "" {
		return
	}

	text := strings.TrimSpace(event.Text)

	if !strings.HasPrefix(text, "!nearme") {
		return
	}

	reply := func(msg string) {
		if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText(msg, false)); err != nil {
			log.Printf("posting message: %v", err)
		}
	}

	if _, _, _, err := apihelper.ParseNearMe(text); err != nil {
		reply("Usage: `!nearme (<lat>, <long>) [distance in metres]`")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	data, err := apiClient.NearMe(ctx, text)
	if err != nil {
		log.Printf("nearme: %v", err)
		reply("Sorry, couldn't fetch bus info right now.")
		return
	}

	msg, err := responseformatter.FormatLocationInfo(data)
	if err != nil {
		log.Printf("formatting: %v", err)
		reply("Sorry, couldn't read the bus info response.")
		return
	}

	reply(msg)
}
