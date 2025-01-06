package main

import (
	"fmt"
	"strconv"

	tg "github.com/semog/go-bot-api/v5"
	"k8s.io/klog"
)

func handleInlineQuery(bot *tg.BotAPI, update tg.Update, st Store) error {
	polls, err := st.GetPollsByUser(update.InlineQuery.From.ID)
	if err != nil {
		return fmt.Errorf("could not get polls for user: %v", err)
	}

	if len(polls) > maxPollsInlineQuery {
		polls = polls[0 : maxPollsInlineQuery-1]
	}
	results := make([]interface{}, len(polls))
	for i, p := range polls {
		klog.Infoln(p)
		article := tg.NewInlineQueryResultArticleHTML(strconv.Itoa(p.ID), p.Question, buildPollListing(p, st))
		if len(p.Options) > 0 {
			article.ReplyMarkup = buildPollMarkup(p)
		}
		article.Description = locInlineInsertPoll

		results[i] = article

	}
	inlineConfig := tg.InlineConfig{
		InlineQueryID:     update.InlineQuery.ID,
		Results:           results,
		IsPersonal:        true,
		CacheTime:         0,
		SwitchPMText:      locCreateNewPoll,
		SwitchPMParameter: qryCreateNewPoll,
	}

	_, err = bot.Request(inlineConfig)
	if err != nil {
		return fmt.Errorf("could not answer inline query: %v", err)
	}

	return nil
}
