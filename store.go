package main

import tg "github.com/semog/telegram-bot-api"

// Store is an interface for the persistent storage
// should allow easier swapping of databases
type Store interface {
	Init(databaseFile string) error
	Close()
	AddMsgToPoll(pollID int, messageID int, chatID int64) error
	AddInlineMsgToPoll(pollID int, inlineMessageID string) error
	RemoveInlineMsg(inlineMessageID string) error
	GetPoll(pollID int) (*poll, error)
	GetUser(userID int) (*tg.User, error)
	GetPollsByUser(userID int) ([]*poll, error)
	GetPollID(messageID int) (int, error)
	GetPollNewer(pollID int, userID int) (*poll, error)
	GetPollOlder(pollID int, userID int) (*poll, error)
	GetAllPollMsg(pollID int) ([]pollident, error)
	GetAllPollInlineMsg(pollID int) ([]pollident, error)
	GetState(userID int) (state int, pollID int, err error)
	SaveState(userID int, pollID int, state int) error
	SaveUser(*tg.User) error
	SavePoll(*poll) (int, error)
	SaveOptions([]option) error
	SaveAnswer(*poll, answer) (unvoted bool, err error)
	DeletePoll(userID int, pollID int) error
}
