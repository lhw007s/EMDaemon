package main

import (
	"container/list"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type ScheduleList map[int64]DnfEventSchedule
type ScheduleAddData map[int64]DnfScheduleAddData
type EventDetail map[int64]DnfEventDetailLog
type EventLog map[int64]DnfEventLog

type DBJob struct {
	dbManager  *DBManager
	load_logId *list.List
	mutex      *sync.Mutex
}

func GetDBJob() *DBJob {
	return &DBJob{
		dbManager:  nil,
		load_logId: list.New(),
		mutex:      new(sync.Mutex),
	}
}

func (dbJob *DBJob) Init() {

	//DB접속
	GLogManager.Info("DB Init..")
	dbJob.dbManager = GetDBManager().Init()

	//	dbJob.startGoRoutine()

	//업데이트 전용 데이터 전용 채널
	dbJob.receiveDataC(GChannelManager.GetDataToDbC())
}

/*
func (dbJob *DBJob) startGoRoutine() {
	GLogManager.WriteLog(INFO, "DB 고루틴 실행합니다.")
	ticker := time.NewTicker(DEFAULT_DB_GOROTUINE_MINUTE * time.Second)
	go func() {
		for _ = range ticker.C {
			//일정시간마다 현재 메모리에 존재하는 이벤트 리스트와 DB리스트를 비교한다.
			//이벤트 리스트 누락 방지용
			//GLogManager.Debug("DB Read Event Data logId (%d)", logId.(int64))

			GLogManager.Debug("DB Save TickCount !")
			dbJob.loadEventData(0)
			time.Sleep(time.Millisecond * 1)
		}
	}()
}
*/

/*
전체 이벤트 스케쥴 정보 Select
대기중, 시작중 이벤트리스트를 확보한다.
*/
func (dbJob *DBJob) getEventScheduleData(logId int64) (map[int64]DnfEventSchedule, error) {

	if dbJob.dbManager == nil {
		panic("DB is nil")
	}

	dbJob.dbManager.SetField("count(*) as cnt")

	if 0 != logId {
		dbJob.dbManager.Where_in("log_id", logId)
	}
	dbJob.dbManager.Where_in("event_status", 0, 1)
	rows, err := dbJob.dbManager.Get("event_schedule")
	if nil != err {
		GLogManager.Error("%s", err)
		panic(err.Error)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		fmt.Println(count)
		if nil != err {
			GLogManager.Error("%s", err)
		}
	}

	dbJob.dbManager.SendPing()
	dbJob.dbManager.SetField("log_id, server_id, unix_timestamp(start_date)" +
		", unix_timestamp(end_date), unix_timestamp(system_sdate) " +
		", unix_timestamp(system_edate), event_status, event_type" +
		", event_weekday")

	if 0 != logId {
		dbJob.dbManager.Where_in("log_id", logId)
	}
	dbJob.dbManager.Where_in("event_status", 0, 1)
	rows, err = dbJob.dbManager.Get("event_schedule")
	if nil != err {

		GLogManager.Error("%s", err)
		return nil, err
	}

	eventScheduleList := make(map[int64]DnfEventSchedule)
	for rows.Next() {
		var dnfEventSchedule DnfEventSchedule
		if err := rows.Scan(&dnfEventSchedule.logId, &dnfEventSchedule.serverId,
			&dnfEventSchedule.startDate, &dnfEventSchedule.endDate,
			&dnfEventSchedule.systemSDate, &dnfEventSchedule.systemEDate,
			&dnfEventSchedule.eventStatus, &dnfEventSchedule.eventType,
			&dnfEventSchedule.eventWeekDay); nil != err {
			GLogManager.Error("%s", err)
			continue
		}
		GLogManager.Debug("%s", dnfEventSchedule)
		eventScheduleList[dnfEventSchedule.logId] = dnfEventSchedule
	}

	if err = rows.Err(); nil != err {
		GLogManager.Error("%s", err)
	}

	return eventScheduleList, err
}

func (dbJob *DBJob) updateEventSchedule(eventSchedule DnfEventSchedule) {
	if nil == dbJob.dbManager {
		panic("DB is nil")
	}

	dbJob.dbManager.SingleSet("system_sdate", "from_unixtimestamp("+strconv.FormatInt(eventSchedule.systemSDate, 32)+")")
	dbJob.dbManager.SingleSet("system_edate", "from_unixtimestamp("+strconv.FormatInt(eventSchedule.systemEDate, 32)+")")
	dbJob.dbManager.SingleSet("event_status", eventSchedule.eventStatus)
	dbJob.dbManager.Where("log_id", eventSchedule.logId)
	results, err := dbJob.dbManager.Update("event_schedule")
	if nil != err {
		GLogManager.Error("Update Error (EventSchedule) logId(%d) system_sdate(%d) "+
			"system_edate(%d), event_status(%d)", eventSchedule.logId, eventSchedule.systemSDate,
			eventSchedule.systemEDate, eventSchedule.eventStatus)
	}

	if affectCnt, _ := results.RowsAffected(); affectCnt > 0 {
		GLogManager.Info("Update Success (EventSchedule) logId(%d) system_sdate(%d) "+
			"system_edate(%d), event_status(%d)", eventSchedule.logId, eventSchedule.systemSDate,
			eventSchedule.systemEDate, eventSchedule.eventStatus)
	}
}

func (dbJob *DBJob) getEventLogData(logId ...interface{}) (map[int64]DnfEventLog, error) {

	if nil == dbJob.dbManager {
		panic("DB is nil")
	}

	dbJob.dbManager.SetField("log_id, occ_time, event_type, " +
		"event_flag, server_id, from_unixtime(start_time)," +
		"from_unixtime(end_time), m_id, expl, etc ," +
		"coin_param, percentage_param, index_param , " +
		"amount_param, type_param, value_param, " +
		"time_param, npc_param")

	if len(logId) > 1 {
		dbJob.dbManager.Where_in("log_id", logId...)
	} else {
		dbJob.dbManager.Where("log_id", logId[0])
	}

	rows, err := dbJob.dbManager.Get("dnf_event_log")
	if nil != err {
		GLogManager.Error("%s", err)
		return nil, err
	}
	defer rows.Close()

	eventLogList := make(map[int64]DnfEventLog)
	for rows.Next() {
		var eventLog DnfEventLog
		if err := rows.Scan(
			&eventLog.logId, &eventLog.occTime, &eventLog.eventType,
			&eventLog.eventFlag, &eventLog.serverId, &eventLog.startTime,
			&eventLog.endTime, &eventLog.mId, &eventLog.expl,
			&eventLog.etc, &eventLog.coinParam, &eventLog.percentageParam,
			&eventLog.indexParam, &eventLog.amountParam, &eventLog.typeParam,
			&eventLog.valueParam, &eventLog.timeParam, &eventLog.npcParam); nil != err {
			GLogManager.Error("%s", err)
			continue
		}
		eventLogList[eventLog.logId] = eventLog
	}

	if err = rows.Err(); nil != err {
		GLogManager.Error("%s", err)
	}
	return eventLogList, err
}

func (dbJob *DBJob) updateEvetLogData(eventLogData DnfEventLog) {
	if nil != dbJob.dbManager {
		panic("DB is nil")
	}

	dbJob.dbManager.SingleSet("end_time", eventLogData.endTime)
	dbJob.dbManager.Where("log_id", eventLogData.logId)
	results, err := dbJob.dbManager.Update("dnf_event_log")
	if nil != err {
		GLogManager.Error("EventLog Update Error , LogId (%d) endTime(%d) , err(%s)",
			eventLogData.logId, eventLogData.endTime, err)
	}

	if rowAffectCnt, _ := results.RowsAffected(); rowAffectCnt > 0 {
		GLogManager.Info("EventLog Update Success , LogId (%d) endTime(%d) , err(%s)",
			eventLogData.logId, eventLogData.endTime, err)
	}
}

func (dbJob *DBJob) getEventDetailData(logId ...interface{}) (map[int64]DnfEventDetailLog, error) {

	if nil == dbJob.dbManager {
		panic("DB is nil")
	}

	if len(logId) == 0 {
		GLogManager.Error("%s", "logId is 0")
		return nil, errors.New("logId is 0")
	}

	dbJob.dbManager.SetField("log_id, apply_type, personality_type" +
		", kor_name, eng_name, url" +
		", start_date, end_date, pop_data" +
		", event_expl, event_visible",
	)

	if len(logId) > 1 {
		dbJob.dbManager.Where_in("log_id", logId...)
	} else {
		dbJob.dbManager.Where("log_id", logId[0])
	}

	rows, err := dbJob.dbManager.Get("dnf_event_detail_log")
	defer rows.Close()
	if nil != err {
		GLogManager.Error("%s", err)
		return nil, err
	}

	eventDetailLogList := make(map[int64]DnfEventDetailLog)
	for rows.Next() {
		eventDetailLog := DnfEventDetailLog{}
		if err := rows.Scan(&eventDetailLog.logId, &eventDetailLog.applyType,
			&eventDetailLog.personalityType, &eventDetailLog.korName,
			&eventDetailLog.engName, &eventDetailLog.url,
			&eventDetailLog.startDate, &eventDetailLog.endDate,
			&eventDetailLog.popData, &eventDetailLog.eventExpl,
			&eventDetailLog.eventVisible); nil != err {
			GLogManager.Cri("%s", err)
			continue
		}
		eventDetailLogList[eventDetailLog.logId] = eventDetailLog
	}

	return eventDetailLogList, nil
}

func (dbJob *DBJob) getEventScheduleAddData(logId ...interface{}) (map[int64]DnfScheduleAddData, error) {

	if nil == dbJob.dbManager {
		panic("DB is nil")
	}

	if len(logId) == 0 {
		GLogManager.Error("%s", "logId is 0")
		return nil, errors.New("logId is 0")
	}

	dbJob.dbManager.SetField("id, log_id, eventStartDay, eventEndDay, eventEndTime")
	if len(logId) > 0 {
		dbJob.dbManager.Where_in("log_id", logId...)
	} else {
		dbJob.dbManager.Where("log_id", logId)
	}
	rows, err := dbJob.dbManager.Get("event_schedule_add_info")
	if nil != err {
		GLogManager.Error("%s", err)
		return nil, err
	}

	eventScheduleAddDataList := make(map[int64]DnfScheduleAddData)
	for rows.Next() {
		eventScheduleAddData := DnfScheduleAddData{}
		err := rows.Scan(&eventScheduleAddData.id, &eventScheduleAddData.logId,
			&eventScheduleAddData.eventStartDay, &eventScheduleAddData.eventEndDay,
			&eventScheduleAddData.eventEndTime)
		if nil != err {
			continue
		}
		eventScheduleAddDataList[eventScheduleAddData.logId] = eventScheduleAddData
	}
	return eventScheduleAddDataList, nil
}

func (dbJob *DBJob) loadEventData(logId int64) {

	//최초 기동될때 모든 데이터를 읽어들인다.
	eventScheduleList, err := dbJob.getEventScheduleData(logId)
	if nil != err {
		GLogManager.Error("%s", err)
	}

	logIdList, err := dbJob.doExtractLogId(eventScheduleList)
	if nil != err {
		panic(err)
	}

	scheduleAddDataList, _ := dbJob.getEventScheduleAddData(logIdList...)
	eventDetailDataList, _ := dbJob.getEventDetailData(logIdList...)
	eventLogDataList, _ := dbJob.getEventLogData(logIdList...)

	dbJob.doSendEventDataGroup(eventScheduleList, scheduleAddDataList, eventDetailDataList, eventLogDataList)
}

func (dbJob *DBJob) doExtractLogId(eventScheduleList map[int64]DnfEventSchedule) ([]interface{}, error) {
	if 0 == len(eventScheduleList) {
		return nil, errors.New("scheduleList length  0")
	}

	fmt.Println(eventScheduleList)
	fmt.Println(len(eventScheduleList))

	var count int = 0
	var logIdList []interface{} = make([]interface{}, len(eventScheduleList))
	for logId := range eventScheduleList {
		if true == dbJob.isLoadLogId(logId) {
			GLogManager.Debug("error duplicated log_id (%d)", logId)
			continue
		}

		if 0 < logId {
			logIdList[count] = logId
			count++
		}
	}

	return logIdList, nil
}

func (dbJob *DBJob) dataSendChannel(c chan<- DnfEventDataGroup) {
	go func() {
		for {

			time.Sleep(time.Millisecond * 1)
			runtime.Gosched()
		}
	}()
}

func (dbJob *DBJob) doSendEventDataGroup(eventScheduleList ScheduleList,
	scheduleAddData ScheduleAddData, eventDetailLog EventDetail, eventLog EventLog) {

	for logId, eventSchedule := range eventScheduleList {
		//패킷 루틴으로 최종적으로 보내진 아이디만 저장한다.
		GLogManager.Debug("%d", logId)
		dbJob.saveLogId(logId)
		if 0 >= eventSchedule.logId {
			continue
		}
		eventDataGroup := DnfEventDataGroup{
			eventLog[logId],
			eventSchedule,
			scheduleAddData[logId],
			eventDetailLog[logId],
			Complete,
		}
		GLogManager.Debug("%s", eventDataGroup)
		GChannelManager.PutDbToDataC(eventDataGroup)
	}
}

func (dbJob *DBJob) receiveDataC(c <-chan DnfEventDataGroup) {
	go func() {
		for {
			eventDataGroup := <-c
			switch eventDataGroup.eventDBAction {
			case Update:
				dbJob.updateEventSchedule(eventDataGroup.eventSchedule)
				dbJob.updateEvetLogData(eventDataGroup.eventLog)
			case Load:
				dbJob.loadEventData(eventDataGroup.eventSchedule.logId)
			}
			runtime.Gosched()
			time.Sleep(time.Millisecond * 1)
		}
	}()
}

func (dbJob *DBJob) saveLogId(logId interface{}) {
	dbJob.mutex.Lock()
	dbJob.load_logId.PushBack(logId)
	dbJob.mutex.Unlock()
}

func (dbJob *DBJob) isLoadLogId(logId interface{}) bool {
	for e := dbJob.load_logId.Front(); e != nil; e = e.Next() {
		if logId.(int64) == e.Value.(int64) {
			return true
		}
	}
	return false
}
