/*

markov-bot - discord bot to save messages into a markov chain
Copyright (C) 2018 superwhiskers <whiskerdev@protonmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

*/

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mb-14/gomarkov"
	log "github.com/sirupsen/logrus"
	"github.com/superwhiskers/harmony"
)

type configType struct {
	Token   string `json:"token"`
	Prefix  string `json:"prefix"`
	Order   int    `json:"order"`
	Hemlock bool   `json:"hemlock"`
}

type hemlockContent struct {
	Content string `json:"content"`
	Rating  int    `json:"rating"`
}

var (
	config        configType
	handler       *harmony.CommandHandler
	chain         *gomarkov.Chain
	mentionRegex  *regexp.Regexp
	hemlockOutput []hemlockContent

	chainMutex = &sync.Mutex{}
	hemlockMutex = &sync.Mutex{}
)

func init() {

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
	})

}

func main() {

	runtime.GOMAXPROCS(100)

	logfile, err := os.OpenFile("markov.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {

		log.Warnf("unable to open logfile. falling back to stdout only. error: %v", err)

	} else {

		defer logfile.Close()
		log.SetOutput(io.MultiWriter(os.Stdout, logfile))

	}

	configByte, err := ioutil.ReadFile("config.json")
	if err != nil {

		log.Panicf("unable to read config file. error: %v", err)

	}

	err = json.Unmarshal(configByte, &config)
	if err != nil {

		log.Panicf("unable to parse config file. error: %v", err)

	}

	if config.Hemlock {

		hemlockMutex.Lock()
		hemlockOutputByte, err := ioutil.ReadFile("hemlock.json")
		err = json.Unmarshal(hemlockOutputByte, &hemlockOutput)
		if err != nil {

			hemlockOutput = []hemlockContent{}

		}
		hemlockMutex.Unlock()

	}

	chainMutex.Lock()
	chainByte, err := ioutil.ReadFile("model.json")
	if err != nil {

		chain = gomarkov.NewChain(config.Order)

	} else {

		err = json.Unmarshal(chainByte, &chain)
		if err != nil {

			log.Panicf("unable to parse model. error: %v", err)

		}

	}
	chainMutex.Unlock()

	dg, err := discordgo.New(fmt.Sprintf("Bot %s", config.Token))
	if err != nil {

		log.Panicf("unable to create a discordgo session object. error: %v", err)

	}

	mentionRegex = regexp.MustCompile("\\@everyone|\\@here")

	handler = harmony.New(config.Prefix, true)
	handler.OnMessageHandler = onMessage
	handler.AddCommand("help", false, help)
	handler.AddCommand("markov", false, markov)

	dg.AddHandler(handler.OnMessage)
	dg.AddHandler(onReady)

	err = dg.Open()
	if err != nil {

		log.Panicf("unable to open the discord session. error: %v", err)

	}

	log.Printf("press ctrl-c to stop the bot...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Printf("stopping the bot...")

	dg.Close()

	chainMutex.Lock()
	chainByte, err = json.Marshal(chain)
	if err != nil {

		log.Panicf("unable to convert the model to json. error: %v", err)

	}

	err = ioutil.WriteFile("model.json", chainByte, 0644)
	if err != nil {

		log.Panicf("unable to write the model to a file. error: %v", err)

	}
	chainMutex.Unlock()

	if config.Hemlock {

		hemlockMutex.Lock()
		hemlockOutputByte, err := json.Marshal(hemlockOutput)
		if err != nil {

			log.Panicf("unable to marshal json. error: %v", err)

		}

		err = ioutil.WriteFile("hemlock.json", hemlockOutputByte, 0644)
		if err != nil {

			log.Panicf("unable to write json to a file. error: %v", err)

		}
		hemlockMutex.Unlock()

	}

}

// updates the model json and hemlock json in the background
func backgroundModelUpdater() {

	var (
		chainByte         []byte
		hemlockOutputByte []byte
		err               error
	)

	for {

		if config.Hemlock {

			hemlockMutex.Lock()
			hemlockOutputByte, err = json.Marshal(hemlockOutput)
			if err != nil {

				log.Panicf("unable to marshal json. error: %v", err)

			}

			err = ioutil.WriteFile("hemlock.json", hemlockOutputByte, 0644)
			if err != nil {

				log.Panicf("unable to write json to a file. error: %v", err)

			}
			hemlockMutex.Unlock()

		}

		chainMutex.Lock()
		chainByte, err = json.Marshal(chain)
		if err != nil {

			log.Panicf("unable to convert the model to json. error: %v", err)

		}
		chainMutex.Unlock()

		err = ioutil.WriteFile("model.json", chainByte, 0644)
		if err != nil {

			log.Panicf("unable to write the model to a file. error: %v", err)

		}

		time.Sleep(10 * time.Second)

	}

}

// handles message create events
func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {

	if strings.HasPrefix(m.Content, config.Prefix) {

		return

	}

	if m.Author.Bot == true {

		return

	}

	content, err := m.ContentWithMoreMentionsReplaced(s)
	if err != nil {

		content = m.ContentWithMentionsReplaced()

	}

	if config.Hemlock {

		hemlockMutex.Lock()
		hemlockOutput = append(hemlockOutput, hemlockContent{
			Content: content,
			Rating:  -1,
		})
		hemlockMutex.Unlock()

	}

	chainMutex.Lock()
	chain.Add(strings.Split(content, " "))
	chainMutex.Unlock()

}

// handles the ready event
func onReady(s *discordgo.Session, r *discordgo.Ready) {

	time.Sleep(500 * time.Millisecond)

	log.Printf("logged in as %s on %d servers...", r.User.String(), len(r.Guilds))

	go backgroundModelUpdater()

}

// command that shows the help message
func help(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Title:       "markov-bot",
		Description: "the bot that generates messages that make no sense",
		Color:       0xFFF176,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "commands",
				Value:  `**help**: shows this message
				**markov [count]**: generate 'count' messages. if 'count' is not provided, it generates one. count is any whole number ranging from 1-5`,
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "built with ❤ by superwhiskers#3210",
		},
	})

	if err != nil {

		log.Errorf("unable to send message. error: %v", err)

	}

	return

}

// command wrapper for generate()
func markov(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {

	if len(args) == 0 {

		text := generate()
		text = string(mentionRegex.ReplaceAllFunc([]byte(text), func(in []byte) []byte {

			return append([]byte("<at>"), in[1:]...)

		}))

		_, err := s.ChannelMessageSend(m.ChannelID, text)
		if err != nil {

			log.Errorf("unable to send message. error: %v", err)

		}

		return

	}

	times, err := strconv.Atoi(args[0])
	if err != nil {

		_, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:       "invalid command argument",
			Description: fmt.Sprintf("\"%s\" is not a whole number", args[0]),
			Color:       0xFFF176,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "built with ❤ by superwhiskers#3210",
			},
		})

		if err != nil {

			log.Errorf("unable to send message. error: %v", err)

		}

		return

	}

	if times > 5 {

		_, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:       "number of messages to generate too high",
			Description: fmt.Sprintf("%s is greater than 5", args[0]),
			Color:       0xFFF176,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "built with ❤ by superwhiskers#3210",
			},
		})

		if err != nil {

			log.Errorf("unable to send message. error: %v", err)

		}

		return

	}

	if times < 1 {

		_, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title: "number of messages to generate too low",
			Description: fmt.Sprintf("%s is less than 1", args[0]),
			Color: 0xFFF176,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "built with ❤ by superwhiskers#3210",
			},
		})

		if err != nil {

			log.Errorf("unable to send message. error: %v", err)

		}

		return

	}

	for i := 0; i < times; i++ {

		text := generate()
		text = string(mentionRegex.ReplaceAllFunc([]byte(text), func(in []byte) []byte {

			return append([]byte("<at>"), in[1:]...)

		}))

		_, err = s.ChannelMessageSend(m.ChannelID, text)
		if err != nil {

			log.Errorf("unable to send message. error: %v", err)

		}

	}

	return

}

// generates text using the chain
func generate() string {

	chainMutex.Lock()

	order := chain.Order
	tokens := make([]string, 0)

	for i := 0; i < order; i++ {

		tokens = append(tokens, gomarkov.StartToken)

	}

	for tokens[len(tokens)-1] != gomarkov.EndToken {

		next, _ := chain.Generate(tokens[(len(tokens) - order):])
		tokens = append(tokens, next)

	}

	chainMutex.Unlock()

	if strings.Join(tokens[order:len(tokens)-1], " ") == "" {

		return generate()

	}

	return strings.Join(tokens[order:len(tokens)-1], " ")

}
