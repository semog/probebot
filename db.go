package main

import (
	"database/sql"
	"fmt"
	"time"

	tg "github.com/semog/go-bot-api/v4"
	"github.com/semog/go-sqldb"
	"k8s.io/klog"
)

type sqlStore struct {
	db *sqldb.SQLDb
}

func (st *sqlStore) Close() {
	err := st.db.Close()
	if err != nil {
		klog.Infof("could not close database properly: %v\n", err)
	}
}

type closable interface {
	Close() error
}

func close(c closable) {
	err := c.Close()
	if err != nil {
		klog.Infof("could not close stmt or rows properly: %v\n", err)
	}
}

func newSQLStore(databaseFile string) *sqlStore {
	st := &sqlStore{}
	err := st.Init(databaseFile)
	if err != nil {
		klog.Fatalf("could not open database %s: %v", databaseFile, err)
	}
	return st
}

func (st *sqlStore) GetUser(userID int) (*tg.User, error) {
	u := &tg.User{ID: userID}

	row := st.db.QueryRow("SELECT FirstName, LastName, UserName FROM user WHERE ID = ?", userID)
	if err := row.Scan(&u.FirstName, &u.LastName, &u.UserName); err != nil {
		return u, fmt.Errorf(`could not scan user "%d": %v`, u.ID, err)
	}

	return u, nil
}

func (st *sqlStore) GetPoll(pollID int) (*poll, error) {
	p := &poll{ID: pollID}
	var err error
	row := st.db.QueryRow("SELECT UserID, Question, Inactive, Type, DisplayPercent FROM poll WHERE ID = ?", pollID)
	if err := row.Scan(&p.UserID, &p.Question, &p.Inactive, &p.Type, &p.DisplayPercent); err != nil {
		return p, fmt.Errorf("could not scan poll #%d: %v", p.ID, err)
	}

	p.Options, err = st.GetOptions(p.ID)
	if err != nil {
		return p, fmt.Errorf("could not query options: %v", err)
	}

	p.Answers, err = st.GetAnswers(p.ID)
	if err != nil {
		return p, fmt.Errorf("could not query answers: %v", err)
	}

	return p, nil
}

func (st *sqlStore) GetPollNewer(pollID int, userID int) (*poll, error) {
	p := &poll{}
	var err error
	row := st.db.QueryRow("SELECT UserID, ID, Question, Inactive, Type, DisplayPercent FROM poll WHERE ID > ? AND UserID = ? ORDER BY ID ASC LIMIT 1", pollID, userID)
	if err := row.Scan(&p.UserID, &p.ID, &p.Question, &p.Inactive, &p.Type, &p.DisplayPercent); err != nil {
		return p, fmt.Errorf("could not scan poll #%d: %v", p.ID, err)
	}

	p.Options, err = st.GetOptions(p.ID)
	if err != nil {
		return p, fmt.Errorf("could not query options: %v", err)
	}

	p.Answers, err = st.GetAnswers(p.ID)
	if err != nil {
		return p, fmt.Errorf("could not query answers: %v", err)
	}

	return p, nil
}

func (st *sqlStore) GetPollOlder(pollID int, userID int) (*poll, error) {
	p := &poll{}
	var err error
	row := st.db.QueryRow("SELECT UserID, ID, Question, Inactive, Type, DisplayPercent FROM poll WHERE ID < ? AND UserID = ? ORDER BY ID DESC LIMIT 1", pollID, userID)
	if err := row.Scan(&p.UserID, &p.ID, &p.Question, &p.Inactive, &p.Type, &p.DisplayPercent); err != nil {
		return p, fmt.Errorf("could not scan poll #%d: %v", p.ID, err)
	}

	p.Options, err = st.GetOptions(p.ID)
	if err != nil {
		return p, fmt.Errorf("could not query options: %v", err)
	}

	p.Answers, err = st.GetAnswers(p.ID)
	if err != nil {
		return p, fmt.Errorf("could not query answers: %v", err)
	}

	return p, nil
}

func (st *sqlStore) GetState(userID int) (state int, pollID int, err error) {
	row := st.db.QueryRow("SELECT state, PollID FROM dialog WHERE UserID = ?", userID)
	if err := row.Scan(&state, &pollID); err != nil {
		return state, pollID, fmt.Errorf("could not scan state from row: %v", err)
	}
	return state, pollID, nil
}

func (st *sqlStore) SaveState(userID int, pollID int, state int) (err error) {
	res, err := st.db.ExecResults("UPDATE dialog SET state = ? WHERE UserID = ?", userID, state)
	if err != nil {
		return fmt.Errorf("could not save state: could not update state in database: %v", err)
	}

	if aff, err := res.RowsAffected(); aff == 0 || err != nil {
		err = st.db.Exec("INSERT OR REPLACE INTO dialog(UserID, PollID, state) values(?, ?, ?)", userID, pollID, state)
		if err != nil {
			return fmt.Errorf("could not save state: could not insert or replace state database entry: %v", err)
		}
	}

	return nil
}

func (st *sqlStore) GetPollsByUser(userID int) ([]*poll, error) {
	polls := make([]*poll, 0)
	var err error

	row, err := st.db.Query("SELECT ID, UserID, Question, Inactive, Type, DisplayPercent FROM poll WHERE UserID = ? ORDER BY ID DESC LIMIT 3", userID)
	if err != nil {
		return polls, fmt.Errorf("could not query polls for userID #%d: %v", userID, err)
	}

	for row.Next() {
		p := &poll{UserID: userID}
		if err := row.Scan(&p.ID, &p.UserID, &p.Question, &p.Inactive, &p.Type, &p.DisplayPercent); err != nil {
			return polls, fmt.Errorf("could not scan poll for userID #%d: %v", userID, err)
		}
		p.Options, err = st.GetOptions(p.ID)
		if err != nil {
			return polls, fmt.Errorf("could not query options: %v", err)
		}

		p.Answers, err = st.GetAnswers(p.ID)
		if err != nil {
			return polls, fmt.Errorf("could not query answers: %v", err)
		}

		polls = append(polls, p)
	}
	return polls, nil
}

func (st *sqlStore) GetPollID(messageID int) (int, error) {
	var pollID int

	rows, err := st.db.Query("SELECT PollID FROM pollmsg WHERE MessageID = ?", messageID)
	if err != nil {
		return pollID, fmt.Errorf("could not query pollID: %v", err)
	}
	defer close(rows)
	for rows.Next() {
		err = rows.Scan(&pollID)
		if err != nil {
			return pollID, fmt.Errorf("could not scan pollID: %v", err)
		}
	}
	return pollID, nil
}

type pollident struct {
	MessageID       int
	InlineMessageID string
	ChatID          int64
}

func (st *sqlStore) GetAllPollMsg(pollID int) ([]pollident, error) {
	msgs := make([]pollident, 0)

	rows, err := st.db.Query("SELECT MessageID, ChatID FROM pollmsg WHERE PollID = ?", pollID)
	if err != nil {
		return msgs, fmt.Errorf("could not query pollmsgs: %v", err)
	}
	defer close(rows)
	var msg pollident
	for rows.Next() {
		err = rows.Scan(&msg.MessageID, &msg.ChatID)
		if err != nil {
			return msgs, fmt.Errorf("could not scan pollmsgs: %v", err)
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (st *sqlStore) GetAllPollInlineMsg(pollID int) ([]pollident, error) {
	msgs := make([]pollident, 0)

	rows, err := st.db.Query("SELECT InlineMessageID FROM pollinlinemsg WHERE PollID = ?", pollID)
	if err != nil {
		return msgs, fmt.Errorf("could not query pollinlinemsg: %v", err)
	}
	defer close(rows)
	var msg pollident
	for rows.Next() {
		err = rows.Scan(&msg.InlineMessageID)
		if err != nil {
			return msgs, fmt.Errorf("could not scan pollID: %v", err)
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (st *sqlStore) GetOptions(pollID int) ([]option, error) {

	options := make([]option, 0)

	rows, err := st.db.Query("SELECT Ctr, PollID, ID, Text FROM option WHERE PollID = ? ORDER BY ID", pollID)
	if err != nil {
		return options, fmt.Errorf("could not query options: %v", err)
	}
	defer close(rows)
	var o option
	for rows.Next() {
		err = rows.Scan(&o.Ctr, &o.PollID, &o.ID, &o.Text)
		if err != nil {
			return options, fmt.Errorf("could not scan option: %v", err)
		}
		options = append(options, o)
	}
	return options, nil
}

func (st *sqlStore) GetAnswers(pollID int) ([]answer, error) {
	answers := make([]answer, 0)

	rows, err := st.db.Query("SELECT ID, PollID, OptionID, UserID FROM answer WHERE PollID = ?", pollID)
	if err != nil {
		return answers, fmt.Errorf("could not query answers: %v", err)
	}
	defer close(rows)
	var a answer
	for rows.Next() {
		err = rows.Scan(&a.ID, &a.PollID, &a.OptionID, &a.UserID)
		if err != nil {
			return answers, fmt.Errorf("could not scan answer: %v", err)
		}
		answers = append(answers, a)
	}
	return answers, nil
}

func (st *sqlStore) SaveAnswer(p *poll, a answer) (unvoted bool, err error) {
	tx, err := st.db.Begin()
	if err != nil {
		return false, fmt.Errorf("could not begin database transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				klog.Infof("could not rollback database change: %v", err)
			}
			return
		}
		err = tx.Commit()
	}()

	// find previous votes in this poll
	stmt, err := tx.Prepare("SELECT OptionID FROM answer WHERE PollID = ? AND UserID = ?")
	if err != nil {
		return false, fmt.Errorf("could not prepare sql statement: %v", err)
	}
	defer close(stmt)

	rows, err := stmt.Query(a.PollID, a.UserID)
	if err != nil {
		return false, fmt.Errorf("could not query option rows: %v", err)
	}
	defer close(rows)
	var optionid int
	prevOpts := make([]int, 0) // len should be 0 or 1
	for rows.Next() {
		err = rows.Scan(&optionid)
		if err != nil {
			return false, fmt.Errorf("could not scan optionid: %v", err)
		}
		prevOpts = append(prevOpts, optionid)
	}

	if len(prevOpts) > 0 { // user voted before

		// user clicked the same answer again
		if contains(prevOpts, a.OptionID) {
			stmt, err = tx.Prepare("DELETE FROM answer where PollID = ? AND UserID = ? AND OptionID = ?")
			if err != nil {
				return false, fmt.Errorf("could not prepare sql statement: %v", err)
			}
			_, err = stmt.Exec(a.PollID, a.UserID, a.OptionID)
			if err != nil {
				return false, fmt.Errorf("could not delete previous answer: %v", err)
			}

			// decrement previously selected option(s)
			stmt, err = tx.Prepare("UPDATE option SET Ctr = Ctr - 1 WHERE Ctr > 0 AND ID = ?")
			if err != nil {
				return false, fmt.Errorf("could not prepare sql statement: %v", err)
			}
			if _, err = stmt.Exec(a.OptionID); err != nil {
				return false, fmt.Errorf("could not decrement option: %v", err)
			}
			return true, nil
		}

		if !p.isMultipleChoice() {

			// decrement previously selected option(s)
			stmt, err = tx.Prepare("UPDATE option SET Ctr = Ctr - 1 WHERE ID = ? AND Ctr > 0")
			if err != nil {
				return false, fmt.Errorf("could not prepare sql statement: %v", err)
			}
			for _, o := range prevOpts {
				if _, err = stmt.Exec(o); err != nil {
					return false, fmt.Errorf("could not decrement option: %v", err)
				}
			}

			// update answer
			stmt, err = tx.Prepare("UPDATE answer SET OptionID = ?, LastSaved = ? WHERE UserID = ? AND PollID = ?")
			if err != nil {
				return false, fmt.Errorf("could not prepare sql statement: %v", err)
			}
			_, err = stmt.Exec(a.OptionID, time.Now().UTC().Unix(), a.UserID, a.PollID)
			if err != nil {
				return false, fmt.Errorf("could not update vote: %v", err)
			}
		} else {
			// new vote
			stmt, err = tx.Prepare("INSERT INTO answer(PollID, OptionID, UserID, LastSaved, CreatedAt) values(?, ?, ?, ?, ?)")
			if err != nil {
				return false, fmt.Errorf("could not prepare sql statement: %v", err)
			}
			_, err = stmt.Exec(a.PollID, a.OptionID, a.UserID, time.Now().UTC().Unix(), time.Now().UTC().Unix())
			if err != nil {
				return false, fmt.Errorf("could not insert answer: %v", err)
			}
		}
	} else {
		// new vote
		stmt, err = tx.Prepare("INSERT INTO answer(PollID, OptionID, UserID, LastSaved, CreatedAt) values(?, ?, ?, ?, ?)")
		if err != nil {
			return false, fmt.Errorf("could not prepare sql statement: %v", err)
		}
		_, err = stmt.Exec(a.PollID, a.OptionID, a.UserID, time.Now().UTC().Unix(), time.Now().UTC().Unix())
		if err != nil {
			return false, fmt.Errorf("could not insert answer: %v", err)
		}
	}

	// increment counter for option
	stmt, err = tx.Prepare("UPDATE option set Ctr = Ctr + 1 WHERE ID = ?")
	if err != nil {
		return false, fmt.Errorf("could not prepare sql statement: %v", err)
	}

	_, err = stmt.Exec(a.OptionID)
	if err != nil {
		return false, fmt.Errorf("could not increment option counter: %v", err)
	}

	return false, nil
}

func (st *sqlStore) AddMsgToPoll(pollID int, messageID int, chatID int64) error {
	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin database transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				klog.Infof("could not rollback database change: %v", err)
			}
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO pollmsg(PollID, MessageID, ChatID) values(?, ?, ?)")
	if err != nil {
		return fmt.Errorf("could not build sql insert statement: %v", err)
	}
	defer close(stmt)

	_, err = stmt.Exec(pollID, messageID, chatID)
	if err != nil {
		return fmt.Errorf("could not add message to poll: %v", err)
	}

	return nil
}

func (st *sqlStore) AddInlineMsgToPoll(pollID int, inlinemessageid string) error {
	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin database transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				klog.Infof("could not rollback database change: %v", err)
			}
			return
		}
		err = tx.Commit()
	}()

	// InlineMessageId is the primary key
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO pollinlinemsg(PollID, InlineMessageID) values(?, ?)")
	if err != nil {
		return fmt.Errorf("could not build sql insert statement: %v", err)
	}
	defer close(stmt)

	_, err = stmt.Exec(pollID, inlinemessageid)
	if err != nil {
		return fmt.Errorf("could not add message to poll: %v", err)
	}

	return nil
}

func (st *sqlStore) RemoveInlineMsg(inlinemessageid string) error {
	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin database transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				klog.Infof("could not rollback database change: %v", err)
			}
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare("DELETE FROM pollinlinemsg WHERE InlineMessageID = ?")
	if err != nil {
		return fmt.Errorf("could not build sql insert statement: %v", err)
	}
	defer close(stmt)

	_, err = stmt.Exec(inlinemessageid)
	if err != nil {
		return fmt.Errorf("could not remove inline message: %v", err)
	}

	return nil

}

func (st *sqlStore) SaveOptions(options []option) error {
	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin database transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				klog.Infof("could not rollback database change: %v", err)
			}
			return
		}
		err = tx.Commit()
	}()
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO option(ID, PollID, Ctr, Text) values(?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("could not prepare insert sql statement for options: %v", err)
	}
	defer close(stmt)

	for i := 0; i < len(options); i++ {
		if options[i].ID == 0 {
			// Adding a new option
			id64, err := st.db.GetGkey()
			if err != nil {
				return fmt.Errorf("could not get gkey for option: %v", err)
			}
			options[i].ID = int(id64)
		}
		_, err = stmt.Exec(options[i].ID, options[i].PollID, options[i].Ctr, options[i].Text)
		if err != nil {
			return fmt.Errorf("could not insert or update option into sql database: %v", err)
		}
	}
	return nil
}

func (st *sqlStore) DeleteOptions(options []option) error {
	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin database transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				klog.Infof("could not rollback database change: %v", err)
			}
			return
		}
		err = tx.Commit()
	}()
	stmt, err := tx.Prepare("DELETE FROM option WHERE ID = ?")
	if err != nil {
		return fmt.Errorf("could not prepare insert sql statement for options: %v", err)
	}
	defer close(stmt)

	for i := 0; i < len(options); i++ {
		_, err = stmt.Exec(options[i].ID)
		if err != nil {
			klog.Errorf("could not delete option from database: %v\n", err)
		}
	}
	return nil
}

func (st *sqlStore) SaveUser(u *tg.User) error {
	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin database transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				klog.Infof("could not rollback database change: %v", err)
			}
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare("SELECT count(1) FROM user WHERE ID = ?")
	if err != nil {
		return fmt.Errorf("could not prepare sql statement: %v", err)
	}
	defer close(stmt)

	var cnt int
	err = stmt.QueryRow(u.ID).Scan(&cnt)
	if err != nil {
		return fmt.Errorf("could not check if user '%s' exists: %v", u.UserName, err)
	}
	if cnt != 0 {
		stmt, err = tx.Prepare("UPDATE user SET FirstName = ?, LastName = ?, UserName = ?, LastSaved = ? WHERE ID = ?")
		if err != nil {
			return fmt.Errorf("could not prepare sql statement: %v", err)
		}
		_, err = stmt.Exec(u.FirstName, u.LastName, u.UserName, time.Now().UTC().Unix(), u.ID)
		if err != nil {
			return fmt.Errorf("could not update user entry: %v", err)
		}
		return nil
	}

	stmt, err = tx.Prepare("INSERT INTO user(ID, FirstName, LastName, UserName, LastSaved, CreatedAt) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("could not prepare sql insert statement: %v", err)
	}
	defer close(stmt)

	_, err = stmt.Exec(u.ID, u.FirstName, u.LastName, u.UserName, time.Now().UTC().Unix(), time.Now().UTC().Unix())
	if err != nil {
		return fmt.Errorf("could not execute sql insert statement: %v", err)
	}

	return nil
}

func (st *sqlStore) SavePoll(p *poll) (id int, err error) {
	tx, err := st.db.Begin()
	if err != nil {
		return int(id), fmt.Errorf("could not begin database transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				klog.Infof("could not rollback database change: %v", err)
			}
			return
		}
		err = tx.Commit()
	}()

	if p.ID != 0 {
		var stmt *sql.Stmt
		stmt, err = tx.Prepare("UPDATE poll SET UserID = ?, Question = ?, Inactive = ?, Private = ?, Type = ?, DisplayPercent = ?, LastSaved = ?, CreatedAt = ? WHERE ID = ?")
		if err != nil {
			return id, fmt.Errorf("could not prepare sql statement: %v", err)
		}
		defer close(stmt)
		_, err = stmt.Exec(p.UserID, p.Question, p.Inactive, p.Private, p.Type, p.DisplayPercent, time.Now().UTC().Unix(), time.Now().UTC().Unix(), p.ID)
		if err != nil {
			return id, fmt.Errorf("could not update user entry: %v", err)
		}
		return id, nil
	}

	id64, err := st.db.GetGkey()
	if err != nil {
		return id, fmt.Errorf("could not get poll gkey id: %v", err)
	}
	id = int(id64)

	stmt, err := tx.Prepare("INSERT INTO poll(ID, UserID, Question, Inactive, Private, Type, DisplayPercent, LastSaved, CreatedAt) values(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return id, fmt.Errorf("could not prepare sql insert statement: %v", err)
	}
	defer close(stmt)

	_, err = stmt.Exec(id, p.UserID, p.Question, p.Inactive, p.Private, p.Type, p.DisplayPercent, time.Now().UTC().Unix(), time.Now().UTC().Unix())
	if err != nil {
		return id, fmt.Errorf("could not execute sql insert statement: %v", err)
	}

	return id, nil
}

func (st *sqlStore) ResetPoll(userID int, pollID int) error {
	// Ensure this user owns the poll
	p := &poll{}
	var err error
	row := st.db.QueryRow("SELECT UserID, ID FROM poll WHERE UserID = ? AND ID = ? ORDER BY ID ASC LIMIT 1", userID, pollID)
	if err := row.Scan(&p.UserID, &p.ID); err != nil {
		return fmt.Errorf("could not scan poll #%d: %v", p.ID, err)
	}

	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin database transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				klog.Infof("could not rollback database change: %v", err)
			}
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare("DELETE FROM answer where PollID = ?")
	if err != nil {
		return fmt.Errorf("could not prepare answer sql statement: %v", err)
	}
	defer close(stmt)
	_, err = stmt.Exec(pollID)
	if err != nil {
		return fmt.Errorf("could not delete poll answers: %v", err)
	}
	return nil
}

func (st *sqlStore) DeletePoll(userID int, pollID int) error {
	// Ensure this user owns the poll
	p := &poll{}
	var err error
	row := st.db.QueryRow("SELECT UserID, ID FROM poll WHERE UserID = ? AND ID = ? ORDER BY ID ASC LIMIT 1", userID, pollID)
	if err := row.Scan(&p.UserID, &p.ID); err != nil {
		return fmt.Errorf("could not scan poll #%d: %v", p.ID, err)
	}

	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin database transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				klog.Infof("could not rollback database change: %v", err)
			}
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare("DELETE FROM answer where PollID = ?")
	if err != nil {
		return fmt.Errorf("could not prepare answer sql statement: %v", err)
	}
	defer close(stmt)
	_, err = stmt.Exec(pollID)
	if err != nil {
		return fmt.Errorf("could not delete poll answers: %v", err)
	}

	stmt, err = tx.Prepare("DELETE FROM option where PollID = ?")
	if err != nil {
		return fmt.Errorf("could not prepare option sql statement: %v", err)
	}
	_, err = stmt.Exec(pollID)
	if err != nil {
		return fmt.Errorf("could not delete poll options: %v", err)
	}

	stmt, err = tx.Prepare("DELETE FROM dialog where PollID = ?")
	if err != nil {
		return fmt.Errorf("could not prepare dialog sql statement: %v", err)
	}
	_, err = stmt.Exec(pollID)
	if err != nil {
		return fmt.Errorf("could not delete poll dialog: %v", err)
	}

	stmt, err = tx.Prepare("DELETE FROM pollinlinemsg where PollID = ?")
	if err != nil {
		return fmt.Errorf("could not prepare pollinlinemsg sql statement: %v", err)
	}
	_, err = stmt.Exec(pollID)
	if err != nil {
		return fmt.Errorf("could not delete poll pollinlinemsg: %v", err)
	}

	stmt, err = tx.Prepare("DELETE FROM pollmsg where PollID = ?")
	if err != nil {
		return fmt.Errorf("could not prepare pollmsg sql statement: %v", err)
	}
	_, err = stmt.Exec(pollID)
	if err != nil {
		return fmt.Errorf("could not delete poll pollmsg: %v", err)
	}

	stmt, err = tx.Prepare("DELETE FROM poll where ID = ?")
	if err != nil {
		return fmt.Errorf("could not prepare poll sql statement: %v", err)
	}
	_, err = stmt.Exec(pollID)
	if err != nil {
		return fmt.Errorf("could not delete poll: %v", err)
	}

	return nil
}

func contains(slice []int, n int) bool {
	for _, i := range slice {
		if i == n {
			return true
		}
	}
	return false
}
