package app

import (
	"sort"
	"strconv"
	"time"

	"github.com/SongRunqi/go-todo/internal/i18n"
)

func TransToAlfredItem(todos *[]TodoItem) *[]AlfredItem {
	var items = make([]AlfredItem, 0)
	for i := 0; i < len(*todos); i++ {
		item := AlfredItem{}
		item.Title = "[" + strconv.Itoa((*todos)[i].TaskID) + "] 🎯" + (*todos)[i].TaskName + " " + (*todos)[i].Urgent
		completed := (*todos)[i].Status == "completed"
		var prefix string = ""
		if completed {
			prefix = "✅"
		} else {
			prefix = "⌛️"
		}
		item.Subtitle = prefix + (*todos)[i].TaskDesc
		item.Arg = strconv.Itoa((*todos)[i].TaskID)
		item.Autocomplete = (*todos)[i].TaskName
		items = append(items, item)
	}
	return &items
}

func sortedList(todos *[]TodoItem) []TodoItem {
	score := make(map[int64]int)
	now := time.Now().Unix()
	// assign score with task id, the less score, the higher priority
	for i, v := range *todos {
		s := v.EndTime.Unix() - now
		score[s] = i
	}

	times := make([]int64, 0)
	for k := range score {
		times = append(times, k)
	}
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})
	var newTodos []TodoItem = make([]TodoItem, 0)
	for _, v := range times {
		if item := &(*todos)[score[v]]; v < 0 {
			item.Urgent = i18n.T("time.expired")
		} else {
			days := v / 86400
			hours := (v % 86400) / 3600
			minutes := (v % 3600) / 60

			tip := ""
			if days > 0 {
				tip = tip + i18n.T("time.days", days) + " "
			} else if hours > 0 {
				tip = tip + i18n.T("time.hours", hours) + " "
			} else if minutes > 0 {
				tip = tip + i18n.T("time.minutes", minutes) + " "
			}

			if tip != "" {
				item.Urgent = i18n.T("time.remaining", tip)
			} else {
				item.Urgent = i18n.T("time.expired")
			}
		}
		newTodos = append(newTodos, (*todos)[score[v]])
	}
	return newTodos
}
