package main

import (
	"fmt"
	"html"
	"log"
	"strconv"

	"github.com/kyokomi/emoji"
	tg "github.com/semog/telegram-bot-api"
)

func postPoll(bot *tg.BotAPI, p *poll, chatid int64) (tg.Message, error) {
	share := tg.InlineKeyboardButton{
		Text:              locSharePoll,
		SwitchInlineQuery: &p.Question,
	}
	new := tg.NewInlineKeyboardButtonData(locCreateNewPoll, qryCreatePoll)

	buttons := tg.NewInlineKeyboardRow(share, new)
	markup := tg.NewInlineKeyboardMarkup(buttons)
	messageTxt := locFinishedCreatingPoll
	messageTxt += p.Question + "\n" + lineSep + "\n"

	for i, o := range p.Options {
		messageTxt += strconv.Itoa(i+1) + ". " + o.Text + "\n"
	}
	msg := tg.NewMessage(chatid, messageTxt)
	msg.ReplyMarkup = markup

	return bot.Send(msg)
}

func sendMainMenuMessage(bot *tg.BotAPI, update tg.Update) (tg.Message, error) {
	buttons := make([]tg.InlineKeyboardButton, 0)
	buttons = append(buttons, tg.NewInlineKeyboardButtonData("create poll", qryCreatePoll))
	markup := tg.NewInlineKeyboardMarkup(buttons)
	messageTxt := locMainMenu
	msg := tg.NewMessage(int64(update.Message.From.ID), messageTxt)
	msg.ReplyMarkup = markup

	return bot.Send(msg)
}

func sendInterMessage(bot *tg.BotAPI, update tg.Update, p *poll) (tg.Message, error) {
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
	messageTxt += p.Question + "\n" + lineSep + "\n"

	for i, o := range p.Options {
		messageTxt += strconv.Itoa(i+1) + ". " + o.Text + "\n"
	}
	msg := tg.NewMessage(int64(update.Message.From.ID), messageTxt)
	msg.ReplyMarkup = markup

	return bot.Send(msg)
}

func sendNewQuestionMessage(bot *tg.BotAPI, update tg.Update, st Store) error {
	msg := tg.NewMessage(int64(update.CallbackQuery.From.ID), locNewQuestion)
	_, err := bot.Send(&msg)
	if err != nil {
		return fmt.Errorf("could not send message: %v", err)
	}

	err = st.SaveState(update.CallbackQuery.From.ID, -1, waitingForQuestion)
	if err != nil {
		return fmt.Errorf("could not change state to waiting for questions: %v", err)
	}
	return nil
}

func sendEditMessage(bot *tg.BotAPI, update tg.Update, p *poll) (tg.Message, error) {
	body := "Currently selected Poll:\n<pre>\n"
	body += p.Question + "\n" + lineSep + "\n"
	for i, o := range p.Options {
		body += fmt.Sprintf("%d. %s", i+1, o.Text) + "\n"
	}
	body += "</pre>\n\n"
	msg := tg.NewMessage(int64(update.Message.From.ID), body)
	msg.ParseMode = tg.ModeHTML

	msg.ReplyMarkup = buildEditMarkup(p, false, false)

	return bot.Send(&msg)
}

func buildPollMarkup(p *poll) *tg.InlineKeyboardMarkup {
	buttonrows := make([][]tg.InlineKeyboardButton, 0) //len(p.Options), len(p.Options))
	row := -1
	polledUsers := make(map[int]struct{})

	votesForOption := make(map[int]int)
	for _, a := range p.Answers {
		if !p.isMultipleChoice() {
			if _, ok := polledUsers[a.UserID]; ok {
				continue
			}
		}

		for _, o := range p.Options {

			if a.OptionID == o.ID {
				votesForOption[o.ID]++
				polledUsers[a.UserID] = struct{}{}
			}
		}
	}

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
		label := fmt.Sprintf("%s (%d)", o.Text, votesForOption[o.ID])
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
			if !p.isMultipleChoice() {
				if _, ok := polledUsers[a.UserID]; ok {
					continue
				}
			}
			if a.OptionID == o.ID {
				votesForOption[o.ID]++
				u, err := st.GetUser(a.UserID)
				if err != nil {
					log.Printf("could not get user: %v", err)
					listOfUsers[i] = append(listOfUsers[i], &tg.User{ID: a.UserID})
					continue
				}
				polledUsers[u.ID] = struct{}{}
				listOfUsers[i] = append(listOfUsers[i], u)
			}
		}
	}

	listing += emoji.Sprintf("<b>%s</b>\n%s", html.EscapeString(p.Question), lineSep)
	//log.Printf("Create listing for question: %s\n", p.Question)

	for i, o := range p.Options {
		// Only display the option if there is at least one choice, or else the poll is closed.
		usersOnAnswer := len(listOfUsers[i])
		if usersOnAnswer > 0 || p.isInactive() {
			var part string
			if p.isDisplayVotePercent() && len(p.Answers) > 0 {
				part = fmt.Sprintf(" (%.0f%%)", 100.*float64(votesForOption[o.ID])/float64(len(polledUsers)))
				if votesForOption[o.ID] != o.Ctr {
					log.Printf("Counter for option #%d is off: %d stored vs. %d counted", o.ID, o.Ctr, votesForOption[o.ID])
				}
			} else {
				part = ""
			}
			listing += emoji.Sprint(fmt.Sprintf("\n:ballot_box: <b>%s</b>%s", html.EscapeString(o.Text), part))

			if len(p.Answers) < maxNumberOfUsersListed && usersOnAnswer > 0 {
				for j := 0; j+1 < usersOnAnswer; j++ {
					listing += "\n\u251C " + getFormattedUserLink(listOfUsers[i][j])
				}
				listing += "\n\u2514 " + getFormattedUserLink(listOfUsers[i][usersOnAnswer-1])
			}
			listing += "\n"
		}
	}
	listing += emoji.Sprint(fmt.Sprintf("\n%d :busts_in_silhouette:\n", len(polledUsers)))
	return listing
}

func buildEditMarkup(p *poll, noOlder, noNewer bool) *tg.InlineKeyboardMarkup {
	buttonrows := make([][]tg.InlineKeyboardButton, 0)
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
	}
	buttonMultipleChoice := tg.NewInlineKeyboardButtonData(buttonMultipleChoiceText, p.fmtQuery(qryToggleMultipleChoice))
	buttonrows[1] = append(buttonrows[1], buttonShowVotePct, buttonMultipleChoice)

	buttonEditQuestion := tg.NewInlineKeyboardButtonData(locEditQuestionButton, p.fmtQuery(qryEditQuestion))
	buttonAddOptions := tg.NewInlineKeyboardButtonData(locAddOptionButton, p.fmtQuery(qryAddOptions))
	buttonrows[2] = append(buttonrows[2], buttonEditQuestion, buttonAddOptions)

	buttonInactiveText := locToggleOpen
	if p.isInactive() {
		buttonInactiveText = locToggleInactive
	}
	buttonInactive := tg.NewInlineKeyboardButtonData(buttonInactiveText, p.fmtQuery(qryToggleActive))
	buttonDelete := tg.NewInlineKeyboardButtonData(locDeletePollButton, p.fmtQuery(qryDeletePoll))
	buttonrows[3] = append(buttonrows[3], buttonInactive, buttonDelete)

	buttonShare := tg.InlineKeyboardButton{
		Text:              locSharePoll,
		SwitchInlineQuery: &p.Question,
	}
	buttonNew := tg.NewInlineKeyboardButtonData(locCreateNewPoll, qryCreatePoll)
	buttonrows[4] = append(buttonrows[4], buttonShare, buttonNew)

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
