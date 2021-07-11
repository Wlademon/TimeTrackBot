package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	defTime "time"

	"github.com/Wlademon/vkBot/time"

	"github.com/Wlademon/vkBot/file/cache"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

const DefRepeatTime = int(defTime.Hour / defTime.Second)
const DefMinTime = defTime.Hour * 20

var Recipients map[int64]PoolCommand

func main() {
	initEnv()
	time.InitTime("Europe/Moscow")
	cache.InitCache("cache")
	tempo := initTempo()
	bot, err := initBot()
	if err != nil {
		panic(err)
	}
	Recipients := initRecipients()
	go scheduleSender(bot, tempo, Recipients)
}

func cacheRecipients(recipients map[int64]PoolCommand) {
	marshal, err := json.Marshal(recipients)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = cache.CreateForever("_", string(marshal)).Set()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func initRecipients() map[int64]PoolCommand {
	var recipients map[int64]PoolCommand
	if exist, struc := cache.Get("_"); exist {
		err := json.Unmarshal([]byte(struc), &recipients)
		if err != nil {
			fmt.Println(err)
		}
	}
	return recipients
}

func initTempo() *Tempo {
	tempo := new(Tempo)
	tempo.SetToken(os.Getenv("TEMPO_TOKEN"))
	err := tempo.SetUrl(os.Getenv("TEMPO_URL"))
	if err != nil {
		panic(err)
	}

	return tempo
}

func scheduleSender(bot *tgbotapi.BotAPI, tempo *Tempo, recipients map[int64]*PoolCommand, pool PoolCommandHandlers, now defTime.Time) {
	for {
		for chatId, recipient := range recipients {
			f := func(Entity *CommandEntity) bool {
				commandEntity := *Entity
				message, isTrue := ExecuteCommand(commandEntity.Command(), pool, recipient, tempo)
				if isTrue {
					msg := tgbotapi.NewMessage(chatId, message)
					_, _ = bot.Send(msg)

					return true
				}

				return false
			}
			recipient.Each(f, now)
		}
		defTime.Sleep(defTime.Minute)
	}
}

func initEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
}

func initBot() (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_ACCESS_TOKEN"))
	if err != nil {
		return nil, err
	}
	bot.Debug = false

	return bot, err
}

const (
	StartSelf Command = "/startSelf"
	Start     Command = "/start"
	Get       Command = "/get"        // получить свое стреканое время
	GetSetter Command = "/getSetter"  // установить время отправки
	GetSetMin Command = "/getSetMin"  // установить минимальные часы для себя
	All       Command = "/all"        // получить свое стреканое время
	AllSetter Command = "/allSetter"  // установить время отправки
	AllSetMin Command = "/allSetMin"  // установить минимальные часы для всех
	NameAll   Command = "/nameIdAll"  // получить все имена и id
	NameSelf  Command = "/nameIdSelf" // получить свое имя и id
)

func inArrayString(num []string, an string) int {
	exist := false
	iter := 0
	for i, n := range num {
		if n == an {
			iter = i
			exist = true
			break
		}
	}
	if exist {
		return iter
	}
	return -1
}

type Iterator struct {
	Values map[int64]*PoolCommand
	index  uint
}

func (i *Iterator) Init(v map[int64]*PoolCommand) {
	i.Values = v
}

func (i *Iterator) Each(f func(command *PoolCommand)) {
	for index, elem := range i.Values {
		f(elem)
		i.Values[index] = elem
	}
}
