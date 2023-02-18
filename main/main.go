package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly"
)

type CwwatchInfo struct {
	GameName        string
	CwwatchStatus   int
	GameCrackStatus bool
	CrackGroup      string
}

var globalCwwatchInfo CwwatchInfo

const prefix string = "!cwbot"

func cwwatchWebScraper(gameName string, cwwatchInfo CwwatchInfo) {

	globalCwwatchInfo.GameName = gameName

	urlToCheck := "https://cwwatch.net/" + gameName

	c := colly.NewCollector(
		colly.AllowedDomains("cwwatch.net", "www.cwwatch.net", "https://cwwatch.net/"),
	)

	c.OnError(func(_ *colly.Response, err error) {
		if fmt.Sprintf("%s", err) == "Not Found" {
			globalCwwatchInfo.CwwatchStatus = 404
		}
	})

	c.OnHTML("div[class=status]", func(e *colly.HTMLElement) {
		crackStatus := e.Text
		if strings.ToUpper(crackStatus) == "CRACKED" {
			globalCwwatchInfo.GameCrackStatus = true
		}
	})

	c.OnHTML("div[id=GROUP]", func(h *colly.HTMLElement) {
		globalCwwatchInfo.CrackGroup = h.ChildText(".info")
	})

	c.Visit(urlToCheck)

}

func cwwatchInfoCheck(cwwatchInfo CwwatchInfo) string {
	log.Println(cwwatchInfo)
	if cwwatchInfo.CwwatchStatus == 404 {
		globalCwwatchInfo = CwwatchInfo{}
		return "Game not found. Please check the name or visit the cwwatch.net website."
	}

	if cwwatchInfo.GameCrackStatus {
		globalCwwatchInfo = CwwatchInfo{}
		return cwwatchInfo.GameName + " is cracked by " + cwwatchInfo.CrackGroup + "."
	} else {
		globalCwwatchInfo = CwwatchInfo{}
		return cwwatchInfo.GameName + " is not cracked yet."
	}

}

func main() {
	sess, err := discordgo.New("Bot MTA3NTcxODg2NjAxMDkwNjY1NA.GbEffU.0XCmi0gDN0BKSpWEFq_4DBuj5YSgnjezrVIp6U")
	if err != nil {
		log.Fatal(err)
	}

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		args := strings.Split(m.Content, " ")

		if args[0] != prefix {
			return
		}

		if len(args) == 1 && args[0] == prefix {
			s.ChannelMessageSend(m.ChannelID, "No Game input, please try again!")
		}

		if len(args) == 2 && args[1] != prefix {
			game := args[1]
			cwwatchWebScraper(game, globalCwwatchInfo)
			result := cwwatchInfoCheck(globalCwwatchInfo)
			s.ChannelMessageSend(m.ChannelID, result)
		}

		if len(args) > 2 {
			game := strings.Join(args[1:], "-")
			cwwatchWebScraper(game, globalCwwatchInfo)
			result := cwwatchInfoCheck(globalCwwatchInfo)
			s.ChannelMessageSend(m.ChannelID, result)
		}
	})

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println("the bot is online!")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
