package main

import (
	"fmt"
	"strconv"
	"strings"

	tg "github.com/semog/go-bot-api/v5"
	"k8s.io/klog"
)

func handleCallbackQuery(bot *tg.BotAPI, update tg.Update, st Store) error {
	if update.CallbackQuery.Data == qryDummy {
		return sendToastMessage(bot, update, "")
	}

	if update.CallbackQuery.Data[0] == qryEditPayload {
		return handlePollEditQuery(bot, update, st)
	}

	if update.CallbackQuery.Data == qryCreatePoll {
		return sendNewQuestionMessage(bot, update, st)
	}

	if strings.Contains(update.CallbackQuery.Data, qryPollDone) {
		return handlePollDoneQuery(bot, update, st)
	}

	pollID, optionid, err := parseQueryPayload(update)
	if err != nil {
		return fmt.Errorf("could not parse query payload: %v", err)
	}

	if update.CallbackQuery.InlineMessageID != "" {
		if err := st.AddInlineMsgToPoll(pollID, update.CallbackQuery.InlineMessageID); err != nil {
			return fmt.Errorf("could not add inline message to poll: %v", err)
		}
	}

	userID, err := getUpdateUserID(update)
	if err != nil {
		sendToastMessage(bot, update, locInvalidUserMessage)
		return err
	}
	p, err := st.GetPoll(pollID)
	if err != nil {
		sendToastMessage(bot, update, locErrUpdatingPollMessage)
		return fmt.Errorf("could not get poll: %v", err)
	}
	if p.isInactive() {
		sendToastMessage(bot, update, locPollIsInactive)
		return fmt.Errorf("poll %d is inactive", pollID)
	}

	newAnswer := answer{
		UserID:   userID,
		PollID:   pollID,
		OptionID: optionid,
	}
	unvoted, err := st.SaveAnswer(p, newAnswer)
	if err != nil {
		return fmt.Errorf("could not save answers: %v", err)
	}
	// polls were changed
	p, err = st.GetPoll(pollID)
	if err != nil {
		sendToastMessage(bot, update, locErrUpdatingPollMessage)
		return fmt.Errorf("could not get poll: %v", err)
	}

	choice, err := findChoice(p, newAnswer.OptionID)
	if err != nil {
		sendToastMessage(bot, update, locErrUpdatingPollMessage)
		return err
	}

	pollsToUpdate.enqueue(p.ID)

	var popupText string
	if unvoted {
		popupText = locSelectionRemoved
	} else {
		popupText = fmt.Sprintf(locYouSelected, choice.Text)
	}

	return sendToastMessage(bot, update, popupText)
}

func findChoice(p *poll, optionID int) (option, error) {
	for _, o := range p.Options {
		if o.ID == optionID {
			return o, nil
		}
	}
	return option{}, fmt.Errorf("could not find option")
}

func updatePollMessages(bot *tg.BotAPI, pollID int, st Store) error {
	p, err := st.GetPoll(pollID)
	if err != nil {
		return fmt.Errorf("could not find poll #%d: %v", pollID, err)
	}

	listing := buildPollListing(p, st)

	var ed tg.EditMessageTextConfig
	ed.Text = listing
	ed.ParseMode = tg.ModeHTML

	if !p.isInactive() {
		ed.ReplyMarkup = buildPollMarkup(p)
	}

	// reset
	ed.ChatID = 0
	ed.MessageID = 0

	msgs, err := st.GetAllPollInlineMsg(p.ID)
	if err != nil {
		return fmt.Errorf("could not get all poll inline messages: %v", err)
	}

	for _, msg := range msgs {
		ed.Text = listing
		ed.InlineMessageID = msg.InlineMessageID
		bot.Send(ed)
	}

	return nil
}

func deletePollMessages(_ *tg.BotAPI /*bot*/, _ int /*pollID*/, _ Store /*st*/) error {
	klog.Infof("TODO: delete existing shared poll messages. Scan pollinlinemsg table and send delete messages for the IDs")
	return nil
}

func handlePollDoneQuery(bot *tg.BotAPI, update tg.Update, st Store) error {
	splits := strings.Split(update.CallbackQuery.Data, ":")
	if len(splits) < 2 {
		return fmt.Errorf("query did not contain the pollID")
	}
	pollID, err := strconv.Atoi(splits[1])
	if err != nil {
		return fmt.Errorf("could not convert string payload to int: %v", err)
	}

	userID, err := getUpdateUserID(update)
	if err != nil {
		return err
	}
	p, err := st.GetUserPoll(pollID, userID)
	if err != nil {
		return fmt.Errorf("could not get poll: %v", err)
	}
	_, err = sendEditMessage(bot, int64(userID), p)
	if err != nil {
		return fmt.Errorf("could not edit finished poll: %v", err)
	}
	err = st.SaveState(userID, p.ID, pollDone)
	if err != nil {
		return fmt.Errorf("could not change state to poll done: %v", err)
	}
	return nil
}

func handlePollEditQuery(bot *tg.BotAPI, update tg.Update, st Store) error {
	splits := strings.Split(update.CallbackQuery.Data, ":")
	if len(splits) < 3 {
		klog.Infoln(splits)
		sendToastMessage(bot, update, locErrUpdatingPollMessage)
		return fmt.Errorf("query wrongly formatted")
	}
	pollID, err := strconv.Atoi(splits[1])
	if err != nil {
		sendToastMessage(bot, update, locErrUpdatingPollMessage)
		return fmt.Errorf("could not convert string payload to int: %v", err)
	}

	var p *poll
	noNewer := false
	noOlder := false
	toggleInactive := false
	toggleMultipleChoice := false
	toggleShowVotePct := false
	userID, err := getUpdateUserID(update)
	if err != nil {
		return err
	}
	switch splits[2] {
	case qryNextPoll:
		p, err = st.GetPollNewer(pollID, userID)
		if err != nil {
			klog.Infof("could not get older poll: %v\n", err)
			noNewer = true
			p, err = st.GetUserPoll(pollID, userID)
			if err != nil {
				sendToastMessage(bot, update, locErrUpdatingPollMessage)
				return fmt.Errorf("could not get poll: %v", err)
			}
		}
	case qryPrevPoll:
		p, err = st.GetPollOlder(pollID, userID)
		if err != nil {
			klog.Infof("could not get older poll: %v\n", err)
			noOlder = true
			p, err = st.GetUserPoll(pollID, userID)
			if err != nil {
				sendToastMessage(bot, update, locErrUpdatingPollMessage)
				return fmt.Errorf("could not get poll: %v", err)
			}
		}
	case qryToggleActive:
		p, err = st.GetUserPoll(pollID, userID)
		if err != nil {
			sendToastMessage(bot, update, locErrUpdatingPollMessage)
			return fmt.Errorf("could not get poll: %v", err)
		}
		toggleInactive = true
	case qryToggleMultipleChoice:
		p, err = st.GetUserPoll(pollID, userID)
		if err != nil {
			sendToastMessage(bot, update, locErrUpdatingPollMessage)
			return fmt.Errorf("could not get poll: %v", err)
		}
		toggleMultipleChoice = true
	case qryToggleShowVotePct:
		p, err = st.GetUserPoll(pollID, userID)
		if err != nil {
			sendToastMessage(bot, update, locErrUpdatingPollMessage)
			return fmt.Errorf("could not get poll: %v", err)
		}
		toggleShowVotePct = true
	case qryEditOptions:
		p, err = st.GetUserPoll(pollID, userID)
		if err != nil {
			sendToastMessage(bot, update, locErrUpdatingPollMessage)
			return fmt.Errorf("could not get poll: %v", err)
		}
		state := waitingForOption
		err = st.SaveState(userID, pollID, state)
		if err != nil {
			return err
		}

		_, err = sendInterMessage(bot, update, p)
		if err != nil {
			return fmt.Errorf("could not send inter message: %v", err)
		}
		return nil
	case qryEditQuestion:
		state := editQuestion
		err = st.SaveState(userID, pollID, state)
		if err != nil {
			sendToastMessage(bot, update, locErrUpdatingPollMessage)
			return err
		}

		msg := tg.NewMessage(update.CallbackQuery.Message.Chat.ID, locEditQuestion)
		_, err = bot.Send(&msg)
		if err != nil {
			return fmt.Errorf("could not send message: %v", err)
		}
		return nil
	case qryResetPoll:
		err = st.ResetPoll(userID, pollID)
		if err != nil {
			sendToastMessage(bot, update, locErrUpdatingPollMessage)
			return fmt.Errorf("could not reset poll: %v", err)
		}
		pollsToUpdate.enqueue(pollID)
		sendToastMessage(bot, update, locResetPollMessage)
		return nil
	case qryDeletePoll:
		p, err = st.GetUserPoll(pollID, userID)
		if err != nil {
			sendToastMessage(bot, update, locErrUpdatingPollMessage)
			return fmt.Errorf("could not get poll: %v", err)
		}

		// The poll must be closed first
		if inactive != p.Inactive {
			sendToastMessage(bot, update, locClosePollBeforeDeleteMessage)
			return nil
		}

		err = st.DeletePoll(userID, pollID)
		if err != nil {
			sendToastMessage(bot, update, locErrDeletingPollMessage)
			return fmt.Errorf("could not delete poll: %v", err)
		}
		sendToastMessage(bot, update, fmt.Sprintf(locPollDeletedMessage, p.Question))
		pollsToDelete.enqueue(pollID)
		// Move to the next poll.
		p, err = st.GetPollOlder(pollID, userID)
		if err != nil {
			klog.Infof("could not get older poll: %v\n", err)
			noOlder = true
			// Move to the previous poll.
			p, err = st.GetPollNewer(pollID, userID)
			if err != nil {
				klog.Infof("could not get older poll: %v\n", err)
				err = st.SaveState(userID, -1, ohHi)
				if err != nil {
					return err
				}
				_, err = sendMainMenuMessage(bot, update)
				if err != nil {
					return fmt.Errorf("could not send main menu message: %v", err)
				}
				return nil
			}
		}
	default:
		return fmt.Errorf("query wrongly formatted")
	}

	if err != nil {
		return err
	}

	if toggleInactive {
		if p.Inactive == open {
			p.Inactive = inactive
		} else {
			p.Inactive = open
		}
		_, err = st.SavePoll(p)
		if err != nil {
			klog.Infoln("could not save toggled inactive state.")
			sendToastMessage(bot, update, locErrUpdatingPollMessage)
		}
	}

	if toggleMultipleChoice {
		if p.Type == standard {
			p.Type = multipleChoice
		} else if p.Type == multipleChoice {
			p.Type = rankedVoting
		} else {
			p.Type = standard
		}
		_, err = st.SavePoll(p)
		if err != nil {
			klog.Infoln("could not save toggled multiple choice state.")
			sendToastMessage(bot, update, locErrUpdatingPollMessage)
		}
	}

	if toggleShowVotePct {
		if p.DisplayPercent == displayVotePercent {
			p.DisplayPercent = hideDisplayVotePercent
		} else {
			p.DisplayPercent = displayVotePercent
		}
		_, err = st.SavePoll(p)
		if err != nil {
			klog.Infoln("could not save toggled show vote percent state.")
			sendToastMessage(bot, update, locErrUpdatingPollMessage)
		}
	}

	messageTxt := getSelectedPollHeader(p)
	messageTxt += getFormattedPreviewPoll(p)

	var ed tg.EditMessageTextConfig
	ed.Text = messageTxt
	ed.ParseMode = tg.ModeHTML
	ed.ReplyMarkup = buildEditMarkup(p, noOlder, noNewer)

	ed.ChatID = update.CallbackQuery.Message.Chat.ID
	ed.MessageID = update.CallbackQuery.Message.MessageID

	_, err = bot.Send(ed)
	if err != nil {
		klog.Infof("could not update message: %v\n", err)
	}
	pollsToUpdate.enqueue(p.ID)
	return nil
}

func sendToastMessage(bot *tg.BotAPI, update tg.Update, msg string) error {
	callbackConfig := tg.NewCallback(update.CallbackQuery.ID, msg)
	_, err := bot.Request(callbackConfig)
	if err != nil {
		klog.Infof("could not send toast message: %v\n", err)
	}
	return nil
}

func parseQueryPayload(update tg.Update) (pollID int, optionid int, err error) {
	dataSplit := strings.Split(update.CallbackQuery.Data, ":")
	if len(dataSplit) != 2 {
		return pollID, optionid, fmt.Errorf("could not parse response")
	}
	pollID, err = strconv.Atoi(dataSplit[0])
	if err != nil {
		return pollID, optionid, fmt.Errorf("could not convert CallbackQuery data pollID to int: %v", err)
	}

	optionid, err = strconv.Atoi(dataSplit[1])
	if err != nil {
		return pollID, optionid, fmt.Errorf("could not convert CallbackQuery data OptionID to int: %v", err)
	}
	return pollID, optionid, nil
}
