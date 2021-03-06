package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var (
	commandPrefix string
	botName       string
	botID         string
	botKey        string
	tenorKey      string
)

func main() {
	commandPrefix, botKey, botName = getConfigVars() //a note for future me: the commandPrefix was originally out of scope of my other functions because I was using the := operator. This causes the variables to fall out of scope after closing curly bracket. The more you knoowwwww
	fmt.Printf("Initializing Polly with command prefix '%s' \n", commandPrefix)
	discord, err := discordgo.New("Bot " + botKey)
	checkErr("Error creating discord session", err)
	user, err := discord.User("@me")
	checkErr("Error retrieving bot account", err)

	botID = user.ID
	//handlers. There are many different types in the library, corresponding to each of these event types https://discordapp.com/developers/docs/topics/gateway#event-names
	discord.AddHandler(readyHandler)
	err = discord.Open()
	checkErr("Unable to open a connection to discord: ", err)

	defer discord.Close()

	//the following helps the program exit gracefully when ^C is used to quit it
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Printf("Signal received, %s \n", sig)
		done <- true
	}()

	fmt.Println("Polly successfully launched. Awaiting signal...")
	<-done

	fmt.Println("Quitting")
	disconnect(discord)
}

func checkErr(msg string, err error) {
	if err != nil {
		panic(fmt.Errorf("%s: %+v", msg, err))
	}
}

func readyHandler(discord *discordgo.Session, ready *discordgo.Ready) {
	err := discord.UpdateStatus(0, "with Markov Chains")
	if err != nil {
		panic(fmt.Errorf("Fatal error, could not update status: %s", err))
	}
	servers := discord.State.Guilds //returns an array of all servers the bot is added to
	fmt.Printf("I'm installed on %d servers. Nice! \n", len(servers))
	total := 0
	for _, g := range servers {
		s, _ := discord.Guild(g.ID)
		total += s.MemberCount
	}
	fmt.Printf("I touch the hearts of %d people. Pretty sweet :)", total)
}

func disconnect(discord *discordgo.Session) {
	//set status to offline (-1)
	err := discord.UpdateStatus(-1, "")
	if err != nil {
		panic(fmt.Errorf("Fatal error, could not update status: %s", err))
	}
	fmt.Println("Set Polly's status to idle. Goodbye.")
}

func getConfigVars() (string, string, string) {
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetDefault("BOT_NAME", "polly")
	err := viper.ReadInConfig()
	if err != nil {
		//panic(fmt.Errorf("Fatal error, check config file/environment variables: %s \n", err))
		//do nothing so environment variables work. The following if statements should catch any actual errors, so this can be commented out
	}
	prefix := viper.GetString("COMMAND_PREFIX")
	if prefix == "" || len(prefix) != 1 {
		panic(fmt.Errorf("Fatal error, check COMMAND_PREFIX environment variable"))
	}
	key := viper.GetString("BOT_KEY")
	if key == "" {
		panic(fmt.Errorf("Fatal error, check BOT_KEY environment variable"))
	}
	name := strings.ToLower(viper.GetString("BOT_NAME"))
	if name == "" {
		panic(fmt.Errorf("Empty BOT_NAME variable, check config"))
	}
	tenorKey := viper.GetString("TENOR_KEY")
	if tenorKey == "" {
		fmt.Printf("TENOR_KEY not specified. The bot will still run, but meme commands will not work.")
	}
	return prefix, key, name
}
