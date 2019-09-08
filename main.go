package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	tg "github.com/semog/go-bot-api"
	"k8s.io/klog"

	_ "github.com/semog/go-sqldb"
)

const probedbFilename = "probes.db"

func main() {
	token := flag.String("token", "Ask @BotFather", "telegram bot token")
	debug := flag.Bool("debug", false, "Show debug information")
	flag.Parse()

	klog.InitFlags(nil)
	if *token == "Ask @BotFather" {
		klog.Fatal("token flag required. Go ask @BotFather.")
		os.Exit(2)
	}

	klog.Info("Connecting...")
	tg.SetLogger(&klogAdapter{})
	bot, err := tg.NewBotAPI(*token)
	if err != nil {
		klog.Fatal(fmt.Fprintf(os.Stderr, "Could not connect to bot: %v\n", err))
		os.Exit(2)
	}

	bot.Debug = *debug

	if err := run(bot); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
}

var pollsToUpdateConstRate = make(chan int, 10)
var pollsToUpdate = newUniqueChan()
var pollsToDeleteConstRate = make(chan int, 10)
var pollsToDelete = newUniqueChan()

func newUniqueChan() *uniqueChan {
	return &uniqueChan{
		C:   make(chan int, 1000),
		ids: make(map[int]struct{})}
}

type uniqueChan struct {
	C   chan int
	ids map[int]struct{}
}

func (u *uniqueChan) enqueue(id int) {
	if _, ok := u.ids[id]; ok {
		klog.Infof("Update for poll #%d is already scheduled.\n", id)
		return
	}
	u.C <- id
	u.ids[id] = struct{}{}
}

func (u *uniqueChan) dequeue() int {
	id := <-u.C
	delete(u.ids, id)
	return id
}

func newTimer() func() {
	start := time.Now()
	return func() {
		klog.Infoln("This action took: ", time.Now().Sub(start))
	}
}

func run(bot *tg.BotAPI) error {
	// fill update channel with constant rate
	go func() {
		for {
			time.Sleep(400 * time.Millisecond)
			pollID := pollsToUpdate.dequeue()
			pollsToUpdateConstRate <- pollID
		}
	}()
	go func() {
		for {
			time.Sleep(400 * time.Millisecond)
			pollID := pollsToDelete.dequeue()
			pollsToDeleteConstRate <- pollID
		}
	}()

	var st Store = newSQLStore(probedbFilename)
	defer st.Close()

	klog.Infof("Authorized on account %s", bot.Self.UserName)

	u := tg.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("could not prepare update channel: %v", err)
	}

	for {
		select {
		case pollID := <-pollsToUpdateConstRate:
			err := updatePollMessages(bot, pollID, st)
			if err != nil {
				klog.Infof("Could not update poll #%d: %v", pollID, err)
			}
		case pollID := <-pollsToDeleteConstRate:
			err := deletePollMessages(bot, pollID, st)
			if err != nil {
				klog.Infof("Could not delete poll #%d messages: %v", pollID, err)
			}
		case update := <-updates:
			stopTimer := newTimer()
			defer stopTimer()

			// INLINE QUERIES
			if update.InlineQuery != nil {
				klog.Infof("InlineQuery from [%s]: %s", update.InlineQuery.From.UserName, update.InlineQuery.Query)

				err = st.SaveUser(update.InlineQuery.From)
				if err != nil {
					klog.Infof("could not save user: %v", err)
				}

				// TODO: Figure out what this is. Looks sketchy. :-(
				// if update.InlineQuery.From.ID == 3761925 {
				// 	err = handleInlineQueryAdmin(bot, update, st)
				// 	if err != nil {
				// 		klog.Infof("could not handle inline query: %v", err)
				// 	}
				// }

				err = handleInlineQuery(bot, update, st)
				if err != nil {
					klog.Infof("could not handle inline query: %v", err)
				}

				continue
			}

			// poll was inserted into a chat
			if update.ChosenInlineResult != nil {
				pollID, err := strconv.Atoi(update.ChosenInlineResult.ResultID)
				if err != nil {
					return fmt.Errorf("could not parse pollID: %v", err)
				}
				err = st.AddInlineMsgToPoll(pollID, update.ChosenInlineResult.InlineMessageID)
				if err != nil {
					return fmt.Errorf("could not add inline message to poll: %v", err)
				}
				continue
			}

			// CALLBACK QUERIES
			if update.CallbackQuery != nil {
				klog.Infof("CallbackQuery from [%s]: %s", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)

				err = st.SaveUser(update.CallbackQuery.From)
				if err != nil {
					klog.Infof("could not save user: %v", err)
				}

				err = handleCallbackQuery(bot, update, st)
				if err != nil {
					klog.Infof("could not handle callback query: %v", err)
				}

				continue
			}

			if update.Message == nil {
				continue
			}

			err = st.SaveUser(update.Message.From)
			if err != nil {
				klog.Infof("could not save user: %v", err)
			}

			// Messages
			klog.Infof("Message from [%s] %s", update.Message.From.UserName, update.Message.Text)

			// Conversations
			err = handleDialog(bot, update, st)
			if err != nil {
				klog.Infof("could not handle dialog: %v", err)
			}
		}
	}
}
