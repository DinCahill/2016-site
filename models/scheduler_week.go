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

func tin(arr []time.Duration, elem time.Duration) bool {
	for _, v := range arr {
		if v == elem {
			return true
		}
	}
	return false
}

// Return a sorted slice of Durations, since midnight, which represent
// a minimum set of times to be printed on a schedule.
// Only use when padded with jukebox.
// TODO: Omit the last value, because it's an end time?
func TableDurations(schedule myradio.Schedule) (durations DurationSlice, err error) {
	for _, day := range schedule {
		dmid := day[0].StartTime
		midnight := time.Date(dmid.Year(), dmid.Month(), dmid.Day(), 0, 0, 0, 0, dmid.Location())
		for _, ts := range day {
			dstart := ts.StartTime.Sub(midnight)
			dend := ts.EndTime().Sub(midnight)
			if !tin(durations, dstart) {
				durations = append(durations, dstart)
			}
			if !tin(durations, dend) {
				durations = append(durations, dend)
			}
		}
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
	for i, _ := range out {
		out[i].TimeStr = times[i]
		out[i].Cells = make([]TableCell, len(schedule))
	}
	for durI, dur := range durations {
		for dayI, day := range schedule {
			for _, ts := range day {
				start := ts.StartTime
				midnight := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
				t := midnight.Add(dur)
				if t.Equal(start) {
					out[durI].Cells[dayI].Timeslot = ts
					out[durI].Cells[dayI].RowSpan = 1
				}
			}
		}
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
