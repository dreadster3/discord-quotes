package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dreadster3/discord-quotes/pkg/models"
	"github.com/joho/godotenv"
)

var quotesUrl string
var userId string
var discordToken string

func init() {
	flag.StringVar(&quotesUrl, "sourceUrl", "", "URL to get quotes from")
	flag.StringVar(&userId, "userId", "", "User ID to send message to")
	flag.StringVar(&discordToken, "discordToken", "", "Discord token")
	flag.Parse()
}

func fallbackEnv(value string, key string) (string, error) {
	if value == "" {
		token, exists := os.LookupEnv(key)
		if !exists {
			return "", errors.New(key + " not found in environment")
		}
		return token, nil
	}

	return value, nil
}

func _main() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	discordToken, err := fallbackEnv(discordToken, "DISCORD_TOKEN")
	if err != nil {
		return err
	}

	userId, err := fallbackEnv(userId, "USER_ID")
	if err != nil {
		return err
	}

	sourceUrl, err := fallbackEnv(quotesUrl, "QUOTES_URL")
	if err != nil {
		return err
	}

	if !strings.HasPrefix(discordToken, "Bot ") {
		discordToken = "Bot " + discordToken
	}

	session, err := discordgo.New(discordToken)
	if err != nil {
		return err
	}
	defer session.Close()
	session.Identify.Intents = discordgo.IntentDirectMessages

	channel, err := session.UserChannelCreate(userId)
	if err != nil {
		return err
	}

	response, err := http.Get(sourceUrl)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	quoteResponse := &models.QuoteResponse{}
	if err := json.Unmarshal(body, quoteResponse); err != nil {
		return err
	}

	if _, err := session.ChannelMessageSend(channel.ID, quoteResponse.Quote); err != nil {
		return err
	}

	fmt.Println("Message sent successfully")

	return nil
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
