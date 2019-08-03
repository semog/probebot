package main

import tg "github.com/semog/telegram-bot-api"

// Store is an interface for the persistent storage
// should allow easier swapping of databases
type Store interface {
	Init(databaseFile string) error
	Close()
	AddMsgToPoll(pollid int, messageid int, chatid int64) error
	AddInlineMsgToPoll(pollid int, inlinemessageid string) error
	RemoveInlineMsg(inlinemessageid string) error
	GetPoll(pollid int) (*poll, error)
	GetUser(userid int) (*tg.User, error)
	GetPollsByUser(userid int) ([]*poll, error)
	GetPollID(messageid int) (int, error)
	GetPollNewer(pollid int, userid int) (*poll, error)
	GetPollOlder(pollid int, userid int) (*poll, error)
	GetAllPollMsg(pollid int) ([]pollident, error)
	GetAllPollInlineMsg(pollid int) ([]pollident, error)
	GetState(userid int) (state int, pollid int, err error)
	SaveState(userid int, pollid int, state int) error
	SaveUser(*tg.User) error
	SavePoll(*poll) (int, error)
	SaveOptions([]option) error
	SaveAnswer(*poll, answer) (unvoted bool, err error)
	DeletePoll(userid int, pollid int) error
}
