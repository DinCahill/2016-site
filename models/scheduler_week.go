package models

import (
	"github.com/UniversityRadioYork/myradio-go"
	"sort"
	"time"
)

// ShowModel is the model for the Show controller.
type ScheduleModel struct {
	Model
}

// NewShowModel returns a new ShowModel on the MyRadio session s.
func NewScheduleModel(s *myradio.Session) *ScheduleModel {
	// @TODO: Pass in the config options
	return &ScheduleModel{Model{session: s}}
}

func (m *ScheduleModel) GetWeek(year string, week string, padded bool) (schedule myradio.Schedule, err error) {
	schedule, err = m.session.GetWeekSchedule(week)
	if err != nil {
		return
	}
	if padded {
		err = m.session.PadWithJukebox(schedule)
		if err != nil {
			return
		}
	}
	return
}

// Return a sorted slice of Durations, since midnight, which represent
// the times to be printed on a schedule, and at on every hour in between.
// Only use when padded with jukebox.
func TableDurations(schedule myradio.Schedule) (durations DurationSlice, err error) {
	// Use a map for uniqueness
	set := make(map[time.Duration]struct{})
	for _, day := range schedule {
		dmid := day[0].StartTime
		midnight := time.Date(dmid.Year(), dmid.Month(), dmid.Day(), 0, 0, 0, 0, dmid.Location())
		for _, ts := range day {
			dstart := ts.StartTime.Sub(midnight)
			dend := ts.EndTime().Sub(midnight)
			ds := []time.Duration{dstart, dend}
			for _, d := range ds {
				set[d] = struct{}{}
			}
		}
	}
	// Pad with hours
	last := schedule[0][len(schedule[0])-1].EndTime()
	first := schedule[0][0].StartTime
	midnight := time.Date(first.Year(), first.Month(), first.Day(), 0, 0, 0, 0, first.Location())
	for d := first.Sub(midnight); midnight.Add(d).Before(last); d += time.Hour {
		set[d] = struct{}{}
	}
	// Convert to slice
	durations = make(DurationSlice, len(set))
	i := 0
	for k := range set {
		durations[i] = k
		i++
	}
	sort.Sort(durations)
	return
}

func TableTimes(durations DurationSlice) (times []string, err error) {
	// Convert to time text for schedule
	times = make([]string, len(durations))
	for k, v := range durations {
		t := time.Time{}.Add(v)
		times[k] = t.Format("15:04")
	}
	return
}

// Use padded schedule
func MakeTable(schedule myradio.Schedule) (Table, error) {
	durations, err := TableDurations(schedule)
	times, err := TableTimes(durations)
	if err != nil {
		return nil, err
	}
	// Make structure
	out := make(Table, len(durations)-1)
	for i := range out {
		out[i].TimeStr = times[i]
		out[i].Cells = make([]TableCell, len(schedule))
	}
	for durI, dur := range durations {
		for dayI, day := range schedule {
			for _, ts := range day {
				mbase := day[0].StartTime
				start := ts.StartTime
				midnight := time.Date(mbase.Year(), mbase.Month(), mbase.Day(), 0, 0, 0, 0, mbase.Location())
				t := midnight.Add(dur)
				if t.Equal(start) {
					// Set the cell if the show starts at this time
					out[durI].Cells[dayI].Timeslot = ts
					out[durI].Cells[dayI].RowSpan = 1
					continue
				}
			}
		}
	}
	// Set RowSpans
	for col := 0; col < len(out[0].Cells); col++ {
		rowspan := 0
		for rowI, row := range out {
			if row.Cells[col].RowSpan == 0 {
				rowspan++
			} else if rowI > 0 {
				prevrow := -10
				for prevrow = rowI - 1; out[prevrow].Cells[col].RowSpan == 0; prevrow-- {
				}
				out[prevrow].Cells[col].RowSpan += rowspan
				rowspan = 0
			}
		}
		// Fix for final day
		used := 0
		i := 0
		for i = len(out) - 1; out[i].Cells[col].RowSpan == 0; i-- {
		}
		for j := i; j >= 0; j-- {
			used += out[j].Cells[col].RowSpan
		}
		out[i].Cells[col].RowSpan += len(out) - used
	}
	return out, nil
}

type TableCell struct {
	Timeslot myradio.Timeslot
	RowSpan  int
}

type TableRow struct {
	TimeStr string
	Cells   []TableCell
}

type Table []TableRow

// Functions to facilitate sorting
type DurationSlice []time.Duration

func (p DurationSlice) Len() int {
	return len(p)
}

func (p DurationSlice) Less(i, j int) bool {
	return p[i] < p[j]
}

func (p DurationSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
