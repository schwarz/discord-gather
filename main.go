package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/logutils"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

const VERSION = "0.2.0"

var (
	Queue = struct {
		sync.RWMutex
		m map[string]string
	}{m: make(map[string]string)}
	Needed int
)

func init() {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	var err error
	Needed, err = strconv.Atoi(os.Getenv("GATHER_NEEDED"))
	if err != nil {
		log.Fatal("[ERROR] Couldn't parse GATHER_NEEDED env variable.")
	}
}

func shuffle(arr []string) {
	rand.Seed(int64(time.Now().Nanosecond()))
	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i)
		arr[i], arr[j] = arr[j], arr[i]
	}
}

func main() {
	var err error
	dg, err := discordgo.New(os.Getenv("GATHER_DISCORD_TOKEN"))
	if err != nil {
		log.Fatal("[ERROR] Couldn't create discord session,", err)
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(presenceUpdate)

	err = dg.Open()
	if err != nil {
		log.Fatal("[ERROR] Couldn't open discord websocket,", err)
	}

	log.Println("[INFO] Bot is now running.")
	// Quit on signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.ChannelID == os.Getenv("GATHER_CHANNELID") && !m.Author.Bot {
		Queue.Lock()
		defer Queue.Unlock()
		switch {
		case strings.HasPrefix(m.Content, "!add"):
			if _, dup := Queue.m[m.Author.ID]; dup {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Player already added (%d/%d).", len(Queue.m), Needed))
				return
			}

			Queue.m[m.Author.ID] = m.Author.Username
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Player added (%d/%d).", len(Queue.m), Needed))
			if len(Queue.m) == Needed {
				var keys []string
				for k := range Queue.m {
					keys = append(keys, k)
				}
				shuffle(keys)
				msg := ""
				for i, _ := range keys {
					if i%(Needed/2) == 0 {
						msg = msg + fmt.Sprintf("**Team %d:**\n", 1+i/(Needed/2))
					}
					msg = msg + fmt.Sprintf("<@%s>\n", keys[i])
				}
				s.ChannelMessageSend(m.ChannelID, msg)
				Queue.m = make(map[string]string)
			}
		case strings.HasPrefix(m.Content, "!remove"):
			if _, ok := Queue.m[m.Author.ID]; ok {
				delete(Queue.m, m.Author.ID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Player removed (%d/%d).", len(Queue.m), Needed))
			}
		case strings.HasPrefix(m.Content, "!status"):
			players := ""
			first := true
			for _, v := range Queue.m {
				if !first {
					players = fmt.Sprintf("%s, %s", players, v)
				} else {
					players = v
					first = false
				}
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s (%d/%d).", players, len(Queue.m), Needed))
		case strings.HasPrefix(m.Content, "!rules"):
			s.ChannelMessageSend(m.ChannelID, os.Getenv("GATHER_RULES"))
		case strings.HasPrefix(m.Content, "!help"):
			s.ChannelMessageSend(m.ChannelID, os.Getenv("GATHER_HELP"))
		}
	}
}

func presenceUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	if p.Status == "offline" {
		Queue.Lock()
		defer Queue.Unlock()
		if _, ok := Queue.m[p.User.ID]; ok {
			delete(Queue.m, p.User.ID)
			s.ChannelMessageSend(os.Getenv("GATHER_CHANNELID"), fmt.Sprintf("Player <@%s> removed for going offline (%d/%d).", p.User.ID, len(Queue.m), Needed))
		}
	}
}
