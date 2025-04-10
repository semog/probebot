package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	tg "github.com/semog/go-bot-api/v5"
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

var pollsToUpdateConstRate = make(chan int, 200)
var pollsToUpdate = newUniqueChan()
var pollsToDeleteConstRate = make(chan int, 200)
var pollsToDelete = newUniqueChan()

func newUniqueChan() *uniqueChan {
	return &uniqueChan{
		C:   make(chan int, 10000),
		ids: make(map[int]struct{})}
}

type uniqueChan struct {
	C   chan int
	ids map[int]struct{}
}

func (u *uniqueChan) enqueue(id int) {
	u.ids[id] = struct{}{}
	u.C <- id
}

func (u *uniqueChan) dequeue() int {
	id := <-u.C
	delete(u.ids, id)
	return id
}

func run(bot *tg.BotAPI) error {
	// fill update channel with constant rate
	go func() {
		for {
			time.Sleep(900 * time.Millisecond)
			pollID := pollsToUpdate.dequeue()
			pollsToUpdateConstRate <- pollID
		}
	}()
	go func() {
		for {
			time.Sleep(900 * time.Millisecond)
			pollID := pollsToDelete.dequeue()
			pollsToDeleteConstRate <- pollID
		}
	}()

	var st Store = newSQLStore(probedbFilename)
	defer st.Close()

	klog.Infof("Authorized on account %s", bot.Self.UserName)

	u := tg.NewUpdate(st.GetUpdateOffset())
	u.Timeout = 60

	// Start the poll service job
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			checkAndUpdatePolls(st)
		}
	}()

	// Reload periodically to get a fresh connection to the Telegram servers.
	reloadTimer := time.NewTimer(20 * time.Hour)
	// Start the update loop
	updates := bot.GetUpdatesChan(u)
	for {
		select {
		case <-reloadTimer.C:
			os.Exit(69)
		case updatePollID := <-pollsToUpdateConstRate:
			klog.Infof("Updating poll #%d\n", updatePollID)
			err := updatePollMessages(bot, updatePollID, st)
			if err != nil {
				klog.Infof("Could not update poll #%d: %v", updatePollID, err)
			}
		case deletePollID := <-pollsToDeleteConstRate:
			klog.Infof("Deleting poll #%d\n", deletePollID)
			err := deletePollMessages(bot, deletePollID, st)
			if err != nil {
				klog.Infof("Could not delete poll #%d messages: %v", deletePollID, err)
			}
		case update := <-updates:
			defer st.SaveUpdateOffset(update.UpdateID + 1)

			// INLINE QUERIES
			if update.InlineQuery != nil {
				klog.Infof("InlineQuery from [%s]: %s", update.InlineQuery.From.UserName, update.InlineQuery.Query)

				err := st.SaveUser(update.InlineQuery.From)
				if err != nil {
					klog.Infof("could not save user: %v", err)
					continue
				}

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
				klog.Infof("Added poll #%d to chat by %s", pollID, update.CallbackQuery.From.UserName)
				continue
			}

			// CALLBACK QUERIES
			if update.CallbackQuery != nil {
				err := st.SaveUser(update.CallbackQuery.From)
				if err != nil {
					klog.Infof("could not save user: %v", err)
					continue
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

			err := st.SaveUser(update.Message.From)
			if err != nil {
				klog.Infof("could not save user: %v", err)
				continue
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
