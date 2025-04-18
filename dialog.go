package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	tg "github.com/semog/go-bot-api/v5"
	"k8s.io/klog"
)

func handleDialog(bot *tg.BotAPI, update tg.Update, st Store) error {
	state := ohHi
	pollID := -1
	userID, err := getUpdateUserID(update)
	if err != nil {
		return err
	}

	if strings.Contains(update.Message.Text, locAboutCommand) {
		msg := tg.NewMessage(userID, locAboutMessage)
		_, err = bot.Send(&msg)
		if err != nil {
			return fmt.Errorf("could not send message: %v", err)
		}
		return err
	}

	state, pollID, err = st.GetState(userID)
	if err != nil {
		// could not retrieve state -> state is zero
		state = ohHi
		klog.Infof("Starting new dialog with user %s\n", update.Message.From.UserName)
	}

	if strings.Contains(update.Message.Text, locEditCommand) {
		polls, err := st.GetPollsByUser(userID)
		if err != nil || len(polls) == 0 {
			klog.Infof("could not get polls of user with userID %d: %v", userID, err)
			state = ohHi
			err = st.SaveState(userID, pollID, state)
			if err != nil {
				return err
			}
			msg := tg.NewMessage(int64(update.Message.From.ID), locNoMessageToEdit)
			_, err = bot.Send(&msg)
			if err != nil {
				return fmt.Errorf("could not send message: %v", err)
			}
		} else {

			var p *poll
			for _, p = range polls {
				if p.ID == pollID {
					break
				}
			}

			_, err = sendEditMessage(bot, int64(userID), p)
			if err != nil {
				return fmt.Errorf("could not send edit message: %v", err)
			}
			return nil
		}
	}

	if strings.Contains(update.Message.Text, "/start") || pollID < 0 && state != waitingForQuestion {
		state = ohHi
		err = st.SaveState(userID, pollID, state)
		if err != nil {
			return err
		}
	}

	if state == ohHi {
		_, err = sendMainMenuMessage(bot, update)
		if err != nil {
			return fmt.Errorf("could not send main menu message: %v", err)
		}
		return nil
	}

	if state == waitingForQuestion {
		p := &poll{
			Question: update.Message.Text,
			UserID:   userID,
		}

		pollID, err = st.SavePoll(p)
		if err != nil {
			return fmt.Errorf("could not save poll: %v", err)
		}

		msg := tg.NewMessage(int64(userID), locGotQuestion)
		_, err = bot.Send(&msg)
		if err != nil {
			return fmt.Errorf("could not send message: %v", err)
		}

		state = waitingForOption
		err = st.SaveState(userID, pollID, state)
		if err != nil {
			return err
		}

		return nil
	}

	if state == waitingForCloseAt {
		p, err := st.GetUserPoll(pollID, userID)
		if err != nil {
			return fmt.Errorf("could not get poll: %v", err)
		}

		closeAt := calcNextServiceTime(time.Now().Unix(), update.Message.Text)
		if closeAt == 0 {
			msg := tg.NewMessage(int64(userID), locInvalidDateTime)
			_, err = bot.Send(&msg)
			if err != nil {
				return fmt.Errorf("could not send message: %v", err)
			}
			return nil
		}

		p.CloseEvery = update.Message.Text
		p.CloseAt = closeAt
		_, err = st.SavePoll(p)
		if err != nil {
			return fmt.Errorf("could not save poll: %v", err)
		}

		msg := tg.NewMessage(int64(userID), fmt.Sprintf(locCloseAtSet, time.Unix(closeAt, 0).Format(dateTimeFormat)))
		_, err = bot.Send(&msg)
		if err != nil {
			return fmt.Errorf("could not send message: %v", err)
		}

		state = editPoll
		err = st.SaveState(userID, pollID, state)
		if err != nil {
			return err
		}

		_, err = sendAdvancedEditMessage(bot, int64(userID), p)
		if err != nil {
			return fmt.Errorf("could not send advanced edit message: %v", err)
		}
		pollsToUpdate.enqueue(pollID)
		return nil
	}

	if state == waitingForOpenAt {
		p, err := st.GetUserPoll(pollID, userID)
		if err != nil {
			return fmt.Errorf("could not get poll: %v", err)
		}

		openAt := calcNextServiceTime(time.Now().Unix(), update.Message.Text)
		if openAt == 0 {
			msg := tg.NewMessage(int64(userID), locInvalidDateTime)
			_, err = bot.Send(&msg)
			if err != nil {
				return fmt.Errorf("could not send message: %v", err)
			}
			return nil
		}

		p.OpenEvery = update.Message.Text
		p.OpenAt = openAt
		_, err = st.SavePoll(p)
		if err != nil {
			return fmt.Errorf("could not save poll: %v", err)
		}

		msg := tg.NewMessage(int64(userID), fmt.Sprintf(locOpenAtSet, time.Unix(openAt, 0).Format(dateTimeFormat)))
		_, err = bot.Send(&msg)
		if err != nil {
			return fmt.Errorf("could not send message: %v", err)
		}

		state = editPoll
		err = st.SaveState(userID, pollID, state)
		if err != nil {
			return err
		}

		_, err = sendAdvancedEditMessage(bot, int64(userID), p)
		if err != nil {
			return fmt.Errorf("could not send advanced edit message: %v", err)
		}
		pollsToUpdate.enqueue(pollID)
		return nil
	}

	if state == waitingForResetAt {
		p, err := st.GetUserPoll(pollID, userID)
		if err != nil {
			return fmt.Errorf("could not get poll: %v", err)
		}

		resetAt := calcNextServiceTime(time.Now().Unix(), update.Message.Text)
		if resetAt == 0 {
			msg := tg.NewMessage(int64(userID), locInvalidDateTime)
			_, err = bot.Send(&msg)
			if err != nil {
				return fmt.Errorf("could not send message: %v", err)
			}
			return nil
		}

		p.ResetEvery = update.Message.Text
		p.ResetAt = resetAt
		_, err = st.SavePoll(p)
		if err != nil {
			return fmt.Errorf("could not save poll: %v", err)
		}

		msg := tg.NewMessage(int64(userID), fmt.Sprintf(locResetAtSet, time.Unix(resetAt, 0).Format(dateTimeFormat)))
		_, err = bot.Send(&msg)
		if err != nil {
			return fmt.Errorf("could not send message: %v", err)
		}

		state = editPoll
		err = st.SaveState(userID, pollID, state)
		if err != nil {
			return err
		}

		_, err = sendAdvancedEditMessage(bot, int64(userID), p)
		if err != nil {
			return fmt.Errorf("could not send advanced edit message: %v", err)
		}
		pollsToUpdate.enqueue(pollID)
		return nil
	}

	if state == editQuestion {
		p, err := st.GetUserPoll(pollID, userID)
		if err != nil {
			return fmt.Errorf("could not get poll: %v", err)
		}

		p.Question = update.Message.Text

		_, err = st.SavePoll(p)
		if err != nil {
			return fmt.Errorf("could not save poll: %v", err)
		}

		msg := tg.NewMessage(int64(userID), fmt.Sprintf(locGotEditQuestion, p.Question))
		_, err = bot.Send(&msg)
		if err != nil {
			return fmt.Errorf("could not send message: %v", err)
		}

		state = editPoll
		err = st.SaveState(userID, pollID, state)
		if err != nil {
			return err
		}
		pollsToUpdate.enqueue(pollID)
	}

	if state == pollDone {
		state = editPoll
		err = st.SaveState(userID, pollID, state)
		if err != nil {
			return err
		}
	}

	if state == editPoll {
		p, err := st.GetUserPoll(pollID, userID)
		if err != nil {
			return fmt.Errorf("could not get poll: %v", err)
		}

		_, err = sendEditMessage(bot, int64(update.Message.Chat.ID), p)
		if err != nil {
			return fmt.Errorf("could not send message: %v", err)
		}
	}

	if state == waitingForOption {
		p, err := st.GetUserPoll(pollID, userID)
		if err != nil {
			return fmt.Errorf("could not get poll: %v", err)
		}
		isEdit, err := regexp.MatchString(`^[0-9]+\.[A-Za-z0-9 ]+`, update.Message.Text)
		if err != nil {
			return fmt.Errorf("could not match regexp: %v", err)
		}
		isDelete, err := regexp.MatchString(`^[0-9]+\.$`, update.Message.Text)
		if err != nil {
			return fmt.Errorf("could not match regexp: %v", err)
		}
		optID := 0
		optText := update.Message.Text
		if isEdit || isDelete {
			r, err := regexp.Compile(`^([0-9]+)\.[ \t]*(.*)`)
			if err != nil {
				return fmt.Errorf("could not compile regexp: %v", err)
			}
			optionMatches := r.FindStringSubmatch(update.Message.Text)
			if optionMatches == nil {
				return fmt.Errorf("could not match regexp: %v", err)
			}
			optNum, err := strconv.Atoi(optionMatches[1])
			if err != nil {
				return fmt.Errorf("could not convert string to number: %v", err)
			}
			if optNum < 1 || optNum > len(p.Options) {
				return fmt.Errorf("option out of range for edit or delete: %v", optNum)
			}
			optID = p.Options[optNum-1].ID
			if len(optionMatches) > 2 {
				optText = optionMatches[2]
			}
		}

		opts := []option{
			{
				ID:     optID,
				PollID: pollID,
				Text:   optText,
			}}
		if isDelete {
			err = st.DeleteOptions(opts)
		} else {
			err = st.SaveOptions(opts)
		}

		if err != nil {
			return fmt.Errorf("could not save option: %v", err)
		}

		pollsToUpdate.enqueue(pollID)
		// Refresh the poll
		p, err = st.GetUserPoll(pollID, userID)
		if err != nil {
			return fmt.Errorf("could not get poll: %v", err)
		}

		_, err = sendInterMessage(bot, update, p)
		if err != nil {
			return fmt.Errorf("could not send inter message: %v", err)
		}
		return nil
	}

	return nil
}
