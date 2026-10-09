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
	locationresponsehandler "slack-bot/location-response-handler"
	routeresponsehandler "slack-bot/route-response-handler"
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
			log.Printf("debug: socket mode event: %s", evt.Type)

			switch evt.Type {

			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("debug: events API payload had unexpected type %T", evt.Data)
					continue
				}

				client.Ack(*evt.Request)

				event := eventsAPIEvent.InnerEvent
				log.Printf("debug: inner event type: %s", event.Type)

				switch ev := event.Data.(type) {
				case *slackevents.MessageEvent:
					handleMessage(api, apiClient, ev)
				default:
					log.Printf("debug: ignoring inner event data of type %T", ev)
				}
			}
		}
	}()

	log.Println("Bot running")
	client.Run()
}

func handleMessage(api *slack.Client, apiClient *apihelper.Client, event *slackevents.MessageEvent) {
	log.Printf("debug: message channel=%s user=%s subtype=%q bot_id=%q text=%q",
		event.Channel, event.User, event.SubType, event.BotID, event.Text)

	// Ignore messages from bots
	if event.BotID != "" {
		log.Println("debug: ignoring message from a bot")
		return
	}

	text := strings.TrimSpace(event.Text)

	switch {
	case strings.HasPrefix(text, "!nearme"):
		log.Println("debug: matched !nearme")
		handleNearMe(api, apiClient, event.Channel, text)
	case strings.HasPrefix(text, "!route"):
		log.Println("debug: matched !route")
		handleRoute(api, apiClient, event.Channel, text)
	default:
		log.Println("debug: message is not a command, ignoring")
	}
}

func newReply(api *slack.Client, channel string) func(string) {
	return func(msg string) {
		if _, _, err := api.PostMessage(channel, slack.MsgOptionText(msg, false)); err != nil {
			log.Printf("posting message: %v", err)
		}
	}
}

func handleNearMe(api *slack.Client, apiClient *apihelper.Client, channel, text string) {
	reply := newReply(api, channel)

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

	msg, err := locationresponsehandler.FormatLocationInfo(data)
	if err != nil {
		log.Printf("formatting: %v", err)
		reply("Sorry, couldn't read the bus info response.")
		return
	}

	reply(msg)
}

func handleRoute(api *slack.Client, apiClient *apihelper.Client, channel, text string) {
	reply := newReply(api, channel)

	if _, _, _, _, err := apihelper.ParseRoute(text); err != nil {
		reply("Usage: `!route (<lat>, <lon>) <operator> <line>`")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	data, err := apiClient.Route(ctx, text)
	if err != nil {
		log.Printf("route: %v", err)
		reply("Sorry, couldn't fetch route info right now.")
		return
	}

	msg, err := routeresponsehandler.FormatRouteInfo(data)
	if err != nil {
		log.Printf("formatting: %v", err)
		reply("Sorry, couldn't read the route info response.")
		return
	}

	reply(msg)
}
