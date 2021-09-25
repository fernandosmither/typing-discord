package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"net/http"

	_ "net/http/pprof"
)

// Discord Bot token
var Token string

// Discord User credentials
var Email string
var Password string
var AuthenticationToken string

// General bot settings (READ ONLY)
var Settings struct {
	BotAsUser    *discordgo.User // Bot account
	isBotAccount bool
}
var KnownUsers map[string]*discordgo.User

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&Email, "e", "", "User Email")
	flag.StringVar(&Password, "p", "", "User Password")
	flag.StringVar(&AuthenticationToken, "a", "", "Authentication Token")
	flag.Parse()

	KnownUsers = make(map[string]*discordgo.User)
}

func main() {
	// Make sure we start with a token supplied or email
	// if len(Token) == 0 && (len(Email) == 0 || len(Password) == 0) {
	// 	flag.Usage()
	// 	fmt.Println("Provide bot token OR email/password")
	// 	return
	// }
	Token := os.Getenv("DISCORD_TOKEN")
	var session *discordgo.Session
	var err error

	// Initiate a new session using Bot Token or useremail/password for authentication
	Settings.isBotAccount = false
	if len(Token) != 0 {
		fmt.Println("Using bot token")
		session, err = discordgo.New("Bot " + Token)
		Settings.isBotAccount = true
	} else if len(AuthenticationToken) == 0 {
		fmt.Println("Using email/password authentication")
		session, err = discordgo.New(Email, Password)
	} else {
		fmt.Println("Using email/password/auth token authentication")
		session, err = discordgo.New(Email, Password, AuthenticationToken)
	}

	if err != nil {
		log.Fatalln("ERROR, Failed to create Discord session:", err)
	}

	// Open a websocket connection to Discord and begin listening
	err = session.Open()
	if err != nil {
		log.Fatalln("ERROR, Couldn't open websocket connection:", err)
	}

	// grab self-user info
	myUser, err := session.User("@me")
	if err != nil {
		log.Fatalln("ERROR, Couldn't get user:", err)
	}
	Settings.BotAsUser = myUser

	// Register the callback for events
	session.AddHandler(messageCreate)
	session.AddHandler(typingStarted)

	// Wait here until CTRL-C or other term signal is received
	log.Println("NOTICE, Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	session.Close()
}

func typingStarted(s *discordgo.Session, ts *discordgo.TypingStart) {
	fmt.Println("typing detected")
	if ts.UserID == Settings.BotAsUser.ID {
		return
	}

	s.ChannelTyping(ts.ChannelID)
}

func knownUser(s *discordgo.Session, uid string) *discordgo.User {
	aUser, known := KnownUsers[uid]
	if known {
		return aUser
	} else {
		aUser, err := s.User(uid)
		if err != nil {
			log.Fatalln("ERROR, Couldn't get user: "+uid, err)
		}
		KnownUsers[uid] = aUser
		return aUser
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("message detected: " + m.Content)
	s.ChannelTyping(m.ChannelID)
}
