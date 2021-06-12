package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tg "github.com/semog/go-bot-api/v4"
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

	p, err := st.GetPoll(pollID)
	if err != nil {
		return fmt.Errorf("could not get poll: %v", err)
	}
	if p.isInactive() {
		sendToastMessage(bot, update, locPollIsInactive)
		return fmt.Errorf("poll %d is inactive", pollID)
	}

	newAnswer := answer{
		UserID:   update.CallbackQuery.From.ID,
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
		return fmt.Errorf("could not get poll: %v", err)
	}

	var choice option
	for _, o := range p.Options {
		if o.ID == newAnswer.OptionID {
			choice = o
		}
	}

	pollsToUpdate.enqueue(p.ID)

	var popupText string
	if unvoted {
		popupText = "Selection removed"
	} else {
		popupText = fmt.Sprintf(`You selected "%s"`, choice.Text)
	}

	return sendToastMessage(bot, update, popupText)
}

func updatePollMessages(bot *tg.BotAPI, pollID int, st Store) error {
	p, err := st.GetPoll(pollID)
	if err != nil {
		return fmt.Errorf("could not get poll: %v", err)
	}

	//msgs, err := st.GetAllPollMsg(pollID)
	//if err != nil {
	//return fmt.Errorf("could not get all pollmsgs: %v", err)
	//}

	listing := buildPollListing(p, st)

	var ed tg.EditMessageTextConfig
	ed.Text = listing
	ed.ParseMode = tg.ModeHTML

	if !p.isInactive() {
		ed.ReplyMarkup = buildPollMarkup(p)
	}

	//for _, msg := range msgs {
	//ed.ChatID = msg.ChatID
	//ed.MessageID = msg.MessageID

	//_, err = bot.Send(ed)
	//if err != nil {
	//klog.Infof("Could not edit message: %v \nThe message was: %s\n", err, ed.Text)
	//klog.Infof("Could not edit message: %v\n", err)
	//splits := strings.Split(ed.Text, "\n")

	//ed.Text = ""
	//for _, l := range splits {
	//if !strings.HasPrefix(l, "\u251C") && !strings.HasPrefix(l, "\u2514") {
	//ed.Text += l + "\n"
	//}
	//}
	//klog.Infof("try again:\n %s", ed.Text)
	//_, err = bot.Send(ed)
	//if err != nil {
	//klog.Infof("could not update message: %v\n", err)
	//continue
	//}
	//}
	//}

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
		_, err := bot.Send(ed)
		if err != nil {
			if strings.Contains(err.Error(), "MESSAGE_ID_INVALID") {
				klog.Infof("Remove inline message %s\n", msg.InlineMessageID)
				st.RemoveInlineMsg(msg.InlineMessageID)
				continue
			}
			// if strings.Contains(err.Error(), "chat not found") {
			// 	klog.Infof("Remove inline message %s\n", msg.InlineMessageID)
			// 	st.RemoveInlineMsg(msg.InlineMessageID)
			// }
			if strings.Contains(err.Error(), "message is not modified") {
				continue
			}
			time.Sleep(20 * time.Millisecond)
			klog.Infof("\n\n\nCould not edit inline message: %v \nThe message was: %s\n", err, ed.Text)
			klog.Infof("Could not edit inline message: %v\n", err)
			splits := strings.Split(ed.Text, "\n")

			ed.Text = ""
			for _, l := range splits {
				if !strings.HasPrefix(l, "\u251C") && !strings.HasPrefix(l, "\u2514") {
					ed.Text += l + "\n"
				}
			}
			klog.Infof("try again:\n %s", ed.Text)
			_, err = bot.Send(ed)
			if err != nil {
				klog.Infof("could not update inline message: %v\n", err)
				continue
			}
		}
		time.Sleep(20 * time.Millisecond)
	}

	return nil
}

func deletePollMessages(bot *tg.BotAPI, pollID int, st Store) error {
	klog.Infof("TODO: delete existing shared poll messages.")
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

	p, err := st.GetPoll(pollID)
	if err != nil {
		return fmt.Errorf("could not get poll: %v", err)
	}
	_, err = sendEditMessage(bot, int64(update.CallbackQuery.From.ID), p)
	if err != nil {
		return fmt.Errorf("could not edit finished poll: %v", err)
	}
	err = st.SaveState(update.CallbackQuery.From.ID, p.ID, pollDone)
	if err != nil {
		return fmt.Errorf("could not change state to poll done: %v", err)
	}
	return nil
}

func handlePollEditQuery(bot *tg.BotAPI, update tg.Update, st Store) error {
	splits := strings.Split(update.CallbackQuery.Data, ":")
	if len(splits) < 3 {
		klog.Infoln(splits)
		return fmt.Errorf("query wrongly formatted")
	}
	pollID, err := strconv.Atoi(splits[1])
	if err != nil {
		return fmt.Errorf("could not convert string payload to int: %v", err)
	}

	var p *poll
	noNewer := false
	noOlder := false
	toggleInactive := false
	toggleMultipleChoice := false
	toggleShowVotePct := false
	switch splits[2] {
	case qryNextPoll:
		p, err = st.GetPollNewer(pollID, update.CallbackQuery.From.ID)
		if err != nil {
			klog.Infof("could not get older poll: %v\n", err)
			noNewer = true
		}
	case qryPrevPoll:
		p, err = st.GetPollOlder(pollID, update.CallbackQuery.From.ID)
		if err != nil {
			klog.Infof("could not get older poll: %v\n", err)
			noOlder = true
		}
	case qryToggleActive:
		p, err = st.GetPoll(pollID)
		if err != nil {
			klog.Infof("could not get poll: %v\n", err)
		}
		toggleInactive = true
	case qryToggleMultipleChoice:
		p, err = st.GetPoll(pollID)
		if err != nil {
			klog.Infof("could not get poll: %v\n", err)
		}
		toggleMultipleChoice = true
	case qryToggleShowVotePct:
		p, err = st.GetPoll(pollID)
		if err != nil {
			klog.Infof("could not get poll: %v\n", err)
		}
		toggleShowVotePct = true
	case qryAddOptions:
		state := waitingForOption
		err = st.SaveState(update.CallbackQuery.From.ID, pollID, state)
		if err != nil {
			return err
		}

		msg := tg.NewMessage(update.CallbackQuery.Message.Chat.ID, locAddOption)
		_, err = bot.Send(&msg)
		if err != nil {
			return fmt.Errorf("could not send message: %v", err)
		}
		return nil
	case qryEditQuestion:
		state := editQuestion
		err = st.SaveState(update.CallbackQuery.From.ID, pollID, state)
		if err != nil {
			return err
		}

		msg := tg.NewMessage(update.CallbackQuery.Message.Chat.ID, locEditQuestion)
		_, err = bot.Send(&msg)
		if err != nil {
			return fmt.Errorf("could not send message: %v", err)
		}
		return nil
	case qryResetPoll:
		err = st.ResetPoll(update.CallbackQuery.From.ID, pollID)
		if err != nil {
			return fmt.Errorf("could not reset poll: %v", err)
		}
		pollsToUpdate.enqueue(pollID)
		sendToastMessage(bot, update, locResetPollMessage)
		return nil
	case qryDeletePoll:
		// The poll must be closed first
		p, err = st.GetPoll(pollID)
		if err != nil {
			sendToastMessage(bot, update, locErrDeletingPollMessage)
			return fmt.Errorf("could not get poll: %v", err)
		}

		if inactive != p.Inactive {
			sendToastMessage(bot, update, locClosePollBeforeDeleteMessage)
			return nil
		}

		err = st.DeletePoll(update.CallbackQuery.From.ID, pollID)
		if err != nil {
			sendToastMessage(bot, update, locErrDeletingPollMessage)
			return fmt.Errorf("could not delete poll: %v", err)
		}
		sendToastMessage(bot, update, fmt.Sprintf(locPollDeletedMessage, p.Question))
		pollsToDelete.enqueue(pollID)
		// Move to the next poll.
		p, err = st.GetPollOlder(pollID, update.CallbackQuery.From.ID)
		if err != nil {
			klog.Infof("could not get older poll: %v\n", err)
			noOlder = true
			// Move to the previous poll.
			p, err = st.GetPollNewer(pollID, update.CallbackQuery.From.ID)
			if err != nil {
				klog.Infof("could not get older poll: %v\n", err)
				err = st.SaveState(update.CallbackQuery.From.ID, -1, ohHi)
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
		p, err = st.GetPoll(pollID)
		if err != nil {
			return fmt.Errorf("could not get poll by poll id: %v", err)
		}
	}

	// danger! malicious client could send pollID from another user in query
	if p.UserID != update.CallbackQuery.From.ID {
		return fmt.Errorf("user does not own poll: %v", err)
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
		}
	}

	if toggleMultipleChoice {
		if p.Type == standard {
			p.Type = multipleChoice
		} else {
			p.Type = standard
		}
		_, err = st.SavePoll(p)
		if err != nil {
			klog.Infoln("could not save toggled multiple choice state.")
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
		}
	}

	messageTxt := locCurrentlySelectedPoll
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
	_, err := bot.AnswerCallbackQuery(callbackConfig)
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
