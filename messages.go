package main

import (
	"fmt"
	"html"

	"github.com/kyokomi/emoji"
	tg "github.com/semog/go-bot-api/v4"
	cmn "github.com/semog/go-common"
	"k8s.io/klog"
)

func getUpdateUserID(update tg.Update) (int, error) {
	if update.Message != nil {
		return update.Message.From.ID, nil
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID, nil
	}
	return 0, fmt.Errorf("invalid update info: no valid user ID found")
}

func sendMainMenuMessage(bot *tg.BotAPI, update tg.Update) (tg.Message, error) {
	userID, err := getUpdateUserID(update)
	if err != nil {
		return tg.Message{}, err
	}
	buttons := make([]tg.InlineKeyboardButton, 0)
	buttons = append(buttons, tg.NewInlineKeyboardButtonData("create poll", qryCreatePoll))
	markup := tg.NewInlineKeyboardMarkup(buttons)
	msg := tg.NewMessage(int64(userID), locMainMenu)
	msg.ReplyMarkup = markup
	return bot.Send(msg)
}

func sendInterMessage(bot *tg.BotAPI, update tg.Update, p *poll) (tg.Message, error) {
	userID, err := getUpdateUserID(update)
	if err != nil {
		return tg.Message{}, err
	}
	//shareButton := tg.InlineKeyboardButton{
	//Text:              locSharePoll,
	//SwitchInlineQuery: &p.Question,
	//}
	pollDoneButton := tg.NewInlineKeyboardButtonData(
		locPollDoneButton, fmt.Sprintf("%s:%d", qryPollDone, p.ID))

	buttons := make([]tg.InlineKeyboardButton, 0)
	buttons = append(buttons, pollDoneButton)
	//buttons = append(buttons, shareButton)

	markup := tg.NewInlineKeyboardMarkup(buttons)
	messageTxt := locAddedOption
	messageTxt += getFormattedPreviewPoll(p)
	msg := tg.NewMessage(int64(userID), messageTxt)
	msg.ParseMode = tg.ModeHTML
	msg.ReplyMarkup = markup
	return bot.Send(msg)
}

func sendNewQuestionMessage(bot *tg.BotAPI, update tg.Update, st Store) error {
	userID, err := getUpdateUserID(update)
	if err != nil {
		return err
	}
	msg := tg.NewMessage(int64(userID), locNewQuestion)
	_, err = bot.Send(&msg)
	if err != nil {
		return fmt.Errorf("could not send message: %v", err)
	}

	err = st.SaveState(userID, -1, waitingForQuestion)
	if err != nil {
		return fmt.Errorf("could not change state to waiting for questions: %v", err)
	}
	return nil
}

func sendEditMessage(bot *tg.BotAPI, chatID int64, p *poll) (tg.Message, error) {
	messageTxt := getSelectedPollHeader(p)
	messageTxt += getFormattedPreviewPoll(p)
	msg := tg.NewMessage(chatID, messageTxt)
	msg.ParseMode = tg.ModeHTML
	msg.ReplyMarkup = buildEditMarkup(p, false, false)
	return bot.Send(&msg)
}

func getSelectedPollHeader(p *poll) string {
	var pollType string

	if p.isRankedVoting() {
		pollType = "ranked voting"
	} else if p.isMultipleChoice() {
		pollType = "multiple choice"
	} else {
		pollType = "single choice"
	}
	return fmt.Sprintf(locCurrentlySelectedPoll, pollType)
}
func getFormattedPreviewPoll(p *poll) string {
	body := fmt.Sprintf("<pre>\n%s\n%s\n", p.Question, lineSep)
	for i, o := range p.Options {
		body += fmt.Sprintf("%d. %s", i+1, o.Text) + "\n"
	}
	body += "</pre>\n\n"
	return body
}

func buildPollMarkup(p *poll) *tg.InlineKeyboardMarkup {
	buttonrows := make([][]tg.InlineKeyboardButton, 0) //len(p.Options), len(p.Options))
	row := -1

	for _, o := range p.Options {
		textWidth := 0
		if row != -1 {
			for _, b := range buttonrows[row] {
				textWidth += len(b.Text)
			}
		}
		textWidth += len(o.Text)
		if row == -1 || textWidth > 30 {
			row++
			buttonrows = append(buttonrows, make([]tg.InlineKeyboardButton, 0))
		}
		label := o.Text
		callback := fmt.Sprintf("%d:%d", p.ID, o.ID)
		button := tg.NewInlineKeyboardButtonData(label, callback)
		buttonrows[row] = append(buttonrows[row], button)
	}
	markup := tg.NewInlineKeyboardMarkup(buttonrows...)
	return &markup
}

func buildPollListing(p *poll, st Store) (listing string) {
	polledUsers := make(map[int]struct{})
	listOfUsers := make([][]*tg.User, len(p.Options))
	votesForOption := make(map[int]int)
	for i, o := range p.Options {
		for _, a := range p.Answers {
			if p.isSingleChoice() {
				if _, ok := polledUsers[a.UserID]; ok {
					continue
				}
			}
			if a.OptionID == o.ID {
				votesForOption[o.ID]++
				u, err := st.GetUser(a.UserID)
				if err != nil {
					klog.Infof("could not get user: %v", err)
					listOfUsers[i] = append(listOfUsers[i], &tg.User{ID: a.UserID})
					continue
				}
				polledUsers[u.ID] = struct{}{}
				listOfUsers[i] = append(listOfUsers[i], u)
			}
		}
	}

	listing += emoji.Sprintf("<b>%s</b>\n%s", html.EscapeString(p.Question), lineSep)
	numPolledUsers := len(polledUsers)
	var irvOptionIDs []int
	if p.isRankedVoting() && numPolledUsers > 0 {
		irvOptionIDs = getIRVOptionIDs(p)
	}

	for i, o := range p.Options {
		// Only display the option if there is at least one choice
		usersOnAnswer := len(listOfUsers[i])
		if usersOnAnswer < 1 {
			continue
		}

		part := ""
		if len(p.Answers) > 0 {
			// Show the number of people that voted for the answer
			part += emoji.Sprintf(" (%d :busts_in_silhouette:", usersOnAnswer)
			if p.isDisplayVotePercent() {
				part += fmt.Sprintf(" %.0f%%", 100.*float64(votesForOption[o.ID])/float64(numPolledUsers))
			}
			part += ")"
		}

		listing += emoji.Sprint(fmt.Sprintf("\n<b>%s</b>%s", html.EscapeString(o.Text), part))
		if intrg_contains(irvOptionIDs, o.ID) {
			listing += emoji.Sprint("  :1st_place_medal:")
		}

		if usersOnAnswer > 0 {
			maxNumberDisplayUsers := cmn.Mini(usersOnAnswer, maxNumberOfUsersListed)
			for j := 0; j+1 < maxNumberDisplayUsers; j++ {
				listing += "\n\u251C " + getFormattedUserLink(listOfUsers[i][j])
			}
			listing += "\n\u2514 " + getFormattedUserLink(listOfUsers[i][usersOnAnswer-1])
		}
		listing += "\n"
	}
	listing += emoji.Sprint(fmt.Sprintf("\n%d :busts_in_silhouette:\n", numPolledUsers))
	return listing
}

func buildEditMarkup(p *poll, noOlder, noNewer bool) *tg.InlineKeyboardMarkup {
	buttonrows := make([][]tg.InlineKeyboardButton, 0)
	buttonrows = append(buttonrows, make([]tg.InlineKeyboardButton, 0))
	buttonrows = append(buttonrows, make([]tg.InlineKeyboardButton, 0))
	buttonrows = append(buttonrows, make([]tg.InlineKeyboardButton, 0))
	buttonrows = append(buttonrows, make([]tg.InlineKeyboardButton, 0))
	buttonrows = append(buttonrows, make([]tg.InlineKeyboardButton, 0))
	buttonrows = append(buttonrows, make([]tg.InlineKeyboardButton, 0))

	buttonLast := tg.NewInlineKeyboardButtonData("\u2B05", p.fmtQuery(qryPrevPoll))
	buttonNext := tg.NewInlineKeyboardButtonData("\u27A1", p.fmtQuery(qryNextPoll))
	if noOlder {
		buttonLast = tg.NewInlineKeyboardButtonData(emoji.Sprint(":checkered_flag: (EOL)"), qryDummy)
	}
	if noNewer {
		buttonNext = tg.NewInlineKeyboardButtonData(emoji.Sprint(":checkered_flag: (EOL)"), qryDummy)
	}
	buttonrows[0] = append(buttonrows[0], buttonLast, buttonNext)

	buttonShowVotePctText := locToggleShowVotePct
	if p.isDisplayVotePercent() {
		buttonShowVotePctText = locToggleHideVotePct
	}
	buttonShowVotePct := tg.NewInlineKeyboardButtonData(buttonShowVotePctText, p.fmtQuery(qryToggleShowVotePct))
	buttonMultipleChoiceText := locToggleSingleChoice
	if p.isMultipleChoice() {
		buttonMultipleChoiceText = locToggleMultipleChoice
	} else if p.isRankedVoting() {
		buttonMultipleChoiceText = locToggleRankedVoting
	}
	buttonMultipleChoice := tg.NewInlineKeyboardButtonData(buttonMultipleChoiceText, p.fmtQuery(qryToggleMultipleChoice))
	buttonrows[1] = append(buttonrows[1], buttonShowVotePct, buttonMultipleChoice)

	buttonEditQuestion := tg.NewInlineKeyboardButtonData(locEditQuestionButton, p.fmtQuery(qryEditQuestion))
	buttonEditOptions := tg.NewInlineKeyboardButtonData(locEditOptionsButton, p.fmtQuery(qryEditOptions))
	buttonrows[2] = append(buttonrows[2], buttonEditQuestion, buttonEditOptions)

	buttonInactiveText := locToggleOpen
	if p.isInactive() {
		buttonInactiveText = locToggleInactive
	}
	buttonInactive := tg.NewInlineKeyboardButtonData(buttonInactiveText, p.fmtQuery(qryToggleActive))
	buttonDelete := tg.NewInlineKeyboardButtonData(locDeletePollButton, p.fmtQuery(qryDeletePoll))
	buttonrows[3] = append(buttonrows[3], buttonInactive, buttonDelete)

	buttonResetPoll := tg.NewInlineKeyboardButtonData(locResetPollButton, p.fmtQuery(qryResetPoll))
	buttonrows[4] = append(buttonrows[4], buttonResetPoll)

	buttonShare := tg.InlineKeyboardButton{
		Text:              locSharePoll,
		SwitchInlineQuery: &p.Question,
	}
	buttonNew := tg.NewInlineKeyboardButtonData(locCreateNewPoll, qryCreatePoll)
	buttonrows[5] = append(buttonrows[5], buttonShare, buttonNew)

	markup := tg.NewInlineKeyboardMarkup(buttonrows...)
	return &markup
}

func getFormattedUserLink(u *tg.User) string {
	return fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", u.ID, html.EscapeString(getDisplayUserName(u)))
}

func getDisplayUserName(u *tg.User) string {
	name := u.FirstName
	if len(u.LastName) > 0 {
		name += " " + u.LastName
	}
	if len(name) == 0 {
		name = u.UserName
	}
	return name
}
