package main

import (
	"time"

	"k8s.io/klog"

	_ "github.com/semog/go-sqldb"
)

func checkAndUpdatePolls(st Store) {
	autoClosePolls(st)
	autoResetPolls(st)
	autoOpenPolls(st)
}

func autoClosePolls(st Store) {
	now := time.Now().Unix()
	polls, err := st.GetPollsWithCloseAtBefore(now)
	if err != nil {
		klog.Errorf("could not get polls with close at: %v", err)
		return
	}

	for _, p := range polls {
		klog.Infof("Closing poll %d automatically", p.ID)
		p.Inactive = inactive
		if p.CloseEvery != "" {
			// Calculate the new CloseAt and update the poll
			p.CloseAt = calcNextServiceTime(now, p.CloseEvery)
		}
		_, err = st.SavePoll(p)
		if err != nil {
			klog.Errorf("could not auto close poll: %v", err)
			continue
		}
		pollsToUpdate.enqueue(p.ID)
	}
}

func autoResetPolls(st Store) {
	now := time.Now().Unix()
	polls, err := st.GetPollsWithResetAtBefore(now)
	if err != nil {
		klog.Errorf("could not get polls with reset at: %v", err)
		return
	}

	for _, p := range polls {
		klog.Infof("Resetting poll %d automatically", p.ID)
		err = st.ResetPoll(p.UserID, p.ID)
		if err != nil {
			klog.Errorf("could not reset poll: %v", err)
			continue
		}
		if p.ResetEvery != "" {
			// Calculate the new ResetAt and update the poll
			p.ResetAt = calcNextServiceTime(now, p.ResetEvery)
		} else {
			p.ResetAt = 0
		}
		_, err = st.SavePoll(p)
		if err != nil {
			klog.Errorf("could not save poll: %v", err)
		}
		pollsToUpdate.enqueue(p.ID)
	}
}

func autoOpenPolls(st Store) {
	now := time.Now().Unix()
	polls, err := st.GetPollsWithOpenAtBefore(now)
	if err != nil {
		klog.Errorf("could not get polls with open at: %v", err)
		return
	}

	for _, p := range polls {
		klog.Infof("Opening poll %d automatically", p.ID)
		p.Inactive = open
		if p.OpenEvery != "" {
			// Calculate the new OpenAt and update the poll
			p.OpenAt = calcNextServiceTime(now, p.OpenEvery)
		}
		_, err = st.SavePoll(p)
		if err != nil {
			klog.Errorf("could not auto open poll: %v", err)
		}
		pollsToUpdate.enqueue(p.ID)
	}
}
