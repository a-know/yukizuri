package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	Token       = flag.String("token", "", "Bot authentication token")
	App         = flag.String("app", "", "Application ID")
	Guild       = flag.String("guild", "", "Guild ID")
	Openaitoken = flag.String("openaitoken", "", "OpenAI token")
	// botのメッセージを保持するスライス
	messagesForZuri []*ChatMessage
	OpenaiClient    *Client
	OpenaiError     error

	Ctx = context.Background()
)

func main() {
	flag.Parse()

	// OpenAIのクライアントを作成
	OpenaiClient, OpenaiError = NewClient(*Openaitoken)
	if OpenaiError != nil {
		fmt.Println(OpenaiError)
	}

	if *App == "" {
		log.Fatal("application id is not set")
	}
	discord, err := discordgo.New("Bot " + *Token)
	if err != nil {
		fmt.Println("ログインに失敗しました")
		fmt.Println(err)
	}
	if err != nil {
		panic(err)
	}

	//イベントハンドラを追加
	discord.AddHandler(onMessageCreate)
	err = discord.Open()
	if err != nil {
		fmt.Println(err)
	}
	// 直近の関数（main）の最後に実行される
	defer discord.Close()

	fmt.Println("Listening...")
	stopBot := make(chan os.Signal, 1)
	signal.Notify(stopBot, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-stopBot
	return
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	clientId := *App
	u := m.Author
	fmt.Printf("%20s > %s\n", u.Username, m.Content)
	if u.ID != clientId {
		if strings.Contains(m.Content, "ずりさん") && *App == "1263105133353504769" {
			// ユーザーの発言を受け取る
			userMessage := NewChatMessage(RoleUser, "おまえさん", "")
			userMessage.Text = m.Content
			messagesForZuri = append(messagesForZuri, userMessage)

			// 人格呼び出し
			personality := GetPersonality("zuri")

			// OpenAIのAPIを叩いて、AIの発言を作成
			message, err := OpenaiClient.Completion(Ctx, "おまえさん", personality, messagesForZuri)
			if err != nil {
				fmt.Println("error:", err.Error())
			}
			fmt.Printf("[%s]\n", message)

			// bot のメッセージを送信
			// sendMessage(s, m.ChannelID, u.Mention()+"なんか喋った!")
			sendMessage(s, m.ChannelID, message.Text)
		} else if strings.Contains(m.Content, "ユキさん") && *App == "1263397754118471740" {
			// ユーザーの発言を受け取る
			userMessage := NewChatMessage(RoleUser, m.Author.Username+"様", "")
			userMessage.Text = m.Content
			messagesForZuri = append(messagesForZuri, userMessage)

			// 人格呼び出し
			personality := GetPersonality("yuki")

			// OpenAIのAPIを叩いて、AIの発言を作成
			message, err := OpenaiClient.Completion(Ctx, m.Author.Username+"様", personality, messagesForZuri)
			if err != nil {
				fmt.Println("error:", err.Error())
			}
			fmt.Printf("[%s]\n", message)

			// bot のメッセージを送信
			// sendMessage(s, m.ChannelID, u.Mention()+"なんか喋った!")
			sendMessage(s, m.ChannelID, message.Text)
		}
	}

}

func sendMessage(s *discordgo.Session, channelID string, msg string) {
	_, err := s.ChannelMessageSend(channelID, msg)
	log.Println(">>> " + msg)
	if err != nil {
		log.Println("Error sending message: ", err)
	}
}
