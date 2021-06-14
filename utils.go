package main

import "time"

func getTimeStamp() int64 {
	return time.Now().UTC().UnixNano()
}

func intrg_contains(rg []int, item int) bool {
	for _, i := range rg {
		if i == item {
			return true
		}
	}
	return false
}
