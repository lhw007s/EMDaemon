// DataManager
package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

/*문자형 날짜 변환*/
const YEAR_LEN = 4
const MONTH_LEN = 2
const DAY_LEN = 2
const HOUR_LEN = 2
const MINUTES_LEN = 2
const SECOND_LEN = 2

type I_Time interface {
	StrToTime(rule string, date string) (r time.Time, err error)
	TimeToStr(rule string, date *time.Time) (r string, err error)
	getYear(date *time.Time) (r string, err error)
	getMonth(date *time.Time) (r string, err error)
	getDay(date *time.Time) (r string, err error)
	getHour(date *time.Time) (r string, err error)
	getMinute(date *time.Time) (r string, err error)
	getSecond(date *time.Time) (r string, err error)
	PanicProcess()
	TimeDebug()
}

type TimeManager struct {
	year   int
	month  int
	day    int
	hour   int
	minute int
	second int
}

func NewTimeManager() *TimeManager {
	return &TimeManager{}
}

/*
Y  - 1999
M  - 01
D - 12
h  - 01
m  - 60
s  - 60
rule : Y(년도 4자리), M(월 2자리) D(일 2자리) h(시간 2자리) m(분 2자리) s(초 2자리)
date : 문자열 시간
return
	정상 : 변환시간, nil
	오류 : 현재시간, 에러
*/
func (t *TimeManager) StrToTime(rule string, date string) (r time.Time, err error) {
	var startIdx int = 0
	defer func() {
		rec := recover()
		if rec != nil {
			r = time.Now()
			err = errors.New(rec.(string))
		}
	}()

	if len(rule) == 0 {
		panic("파싱룰이 필요합니다.")
	}

	if len(date) == 0 {
		panic("변환할 시간이 필요합니다.")
	}

	err = nil

	for i := 0; i < len(rule); i++ {
		switch string(rule[i]) {
		case "Y":
			t.year, err = strconv.Atoi(date[startIdx:YEAR_LEN])
			startIdx += YEAR_LEN
		case "M":
			t.month, err = strconv.Atoi(date[startIdx:(startIdx + MONTH_LEN)])
			startIdx += MONTH_LEN
		case "D":
			t.day, err = strconv.Atoi(date[startIdx:(startIdx + DAY_LEN)])
			startIdx += DAY_LEN
		case "h":
			t.hour, err = strconv.Atoi(date[startIdx:(startIdx + HOUR_LEN)])
			startIdx += HOUR_LEN
		case "m":
			t.minute, err = strconv.Atoi(date[startIdx:(startIdx + MINUTES_LEN)])
			startIdx += MINUTES_LEN
		case "s":
			t.second, err = strconv.Atoi(date[startIdx:(startIdx + SECOND_LEN)])
			startIdx += SECOND_LEN
		default:
			startIdx += 1
		}

		if err != nil {
			panic(err)
		}
	}

	var arrMonth [13]time.Month = [13]time.Month{
		0,
		time.January,
		time.February,
		time.March,
		time.April,
		time.May,
		time.June,
		time.July,
		time.August,
		time.September,
		time.October,
		time.November,
		time.December,
	}

	r = time.Date(t.year, arrMonth[t.month], t.day, t.hour, t.minute, t.second, 0, time.Local)
	return
}

func (t *TimeManager) TimeToStr(rule string, date *time.Time) (r string, err error) {

	if len(rule) == 0 {
		panic("파싱룰이 필요합니다.")
	}

	if date == nil {
		panic("변환할 시간이 필요합니다.")
	}

	defer t.panicProcess(&err)

	for i := 0; i < len(rule); i++ {
		var ret string
		switch string(rule[i]) {
		case "Y":
			ret, err = t.getYear(date)
		case "M":
			ret, err = t.getMonth(date)
		case "D":
			ret, err = t.getDay(date)
		case "h":
			ret, err = t.getHour(date)
		case "m":
			ret, err = t.getMinute(date)
		case "s":
			ret, err = t.getSecond(date)
		default:
			ret = string(rule[i])
		}

		if err != nil {
			panic(err)
		}

		r += ret
	}

	return
}

func (t *TimeManager) getYear(date *time.Time) (r string, err error) {
	if date == nil {
		panic("설정된 시간이 없습니다")
	}

	defer t.panicProcess(&err)

	r = fmt.Sprintf("%02d", date.Year())
	return
}

func (t *TimeManager) getMonth(date *time.Time) (r string, err error) {
	if date == nil {
		panic("설정된 시간이 없습니다")
	}

	defer t.panicProcess(&err)

	r = fmt.Sprintf("%02d", date.Month())
	return
}

func (t *TimeManager) getDay(date *time.Time) (r string, err error) {
	if date == nil {
		panic("설정된 시간이 없습니다")
	}

	defer t.panicProcess(&err)

	r = fmt.Sprintf("%02d", date.Day())
	return
}

func (t *TimeManager) getHour(date *time.Time) (r string, err error) {
	if date == nil {
		panic("설정된 시간이 없습니다")
	}

	defer t.panicProcess(&err)

	r = fmt.Sprintf("%02d", date.Hour())
	return
}

func (t *TimeManager) getMinute(date *time.Time) (r string, err error) {
	if date == nil {
		panic("설정된 시간이 없습니다")
	}

	defer t.panicProcess(&err)

	r = fmt.Sprintf("%02d", date.Minute())
	return
}

func (t *TimeManager) getSecond(date *time.Time) (r string, err error) {
	if date == nil {
		panic("설정된 시간이 없습니다")
	}

	defer t.panicProcess(&err)

	r = fmt.Sprintf("%02d", date.Second())
	return
}

func (t *TimeManager) panicProcess(err *error) {
	recv := recover()
	if recv != nil {
		*err = errors.New(recv.(string))
		return
	}
}
