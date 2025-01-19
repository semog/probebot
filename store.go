package main

import tg "github.com/semog/go-bot-api/v5"

// Store is an interface for the persistent storage
// should allow easier swapping of databases
type Store interface {
	Init(databaseFile string) error
	Close()
	GetUpdateOffset() int
	SaveUpdateOffset(offset int) error
	AddInlineMsgToPoll(pollID int, inlineMessageID string) error
	RemoveInlineMsg(inlineMessageID string) error
	GetUser(userID int64) (*tg.User, error)
	GetPoll(pollID int) (*poll, error)
	GetUserPoll(pollID int, userID int64) (*poll, error)
	GetPollsByUser(userID int64) ([]*poll, error)
	GetPollNewer(pollID int, userID int64) (*poll, error)
	GetPollOlder(pollID int, userID int64) (*poll, error)
	GetAllPollInlineMsg(pollID int) ([]pollident, error)
	GetState(userID int64) (state int, pollID int, err error)
	SaveState(userID int64, pollID int, state int) error
	SaveUser(*tg.User) error
	SavePoll(*poll) (int, error)
	SaveOptions([]option) error
	DeleteOptions([]option) error
	SaveAnswer(*poll, answer) (unvoted bool, err error)
	ResetPoll(userID int64, pollID int) error
	DeletePoll(userID int64, pollID int) error
}
