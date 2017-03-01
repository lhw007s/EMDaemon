package main

type EventSchedule struct {
	LogId        int
	ServerId     int8
	StartDate    int
	EndDate      int
	SystemSDate  int
	SystemEDate  int
	EventStatus  int
	EventType    int8
	EventWeekDay string
}

type EventScheduleAddInfo struct {
	Id            int
	LogId         int
	EventStartDay int8
	EventEndDay   int8
	EventEndTime  int
}
