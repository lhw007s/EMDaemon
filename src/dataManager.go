package main

import (
	"container/list"
	"runtime"
	"sync"
	"time"
)

type DnfEventLog struct {
	logId           int64
	occTime         int
	eventType       int
	eventFlag       int8
	serverId        int8
	startTime       int64
	endTime         int64
	mId             int64
	expl            string
	etc             string
	coinParam       string
	percentageParam string
	indexParam      string
	amountParam     string
	typeParam       string
	valueParam      string
	timeParam       string
	npcParam        string
}

type DnfEventSchedule struct {
	logId        int64
	serverId     int8
	startDate    int64
	endDate      int64
	systemSDate  int64
	systemEDate  int64
	eventStatus  int8
	eventType    int
	eventWeekDay string
}

type DnfScheduleAddData struct {
	id            int64
	logId         int64
	eventStartDay int8
	eventEndDay   int8
	eventEndTime  string
}

type DnfEventDetailLog struct {
	logId           int64
	applyType       int16
	personalityType int16
	korName         string
	engName         string
	url             string
	startDate       int64
	endDate         int64
	popData         string
	eventExpl       string
	eventVisible    int32
}

type DnfEventDataManager struct {
	eventDataGroup map[int64]DnfEventDataGroup
	mutex          *sync.Mutex
}

type DnfEventDataGroup struct {
	eventLog         DnfEventLog
	eventSchedule    DnfEventSchedule
	eventScheduleAdd DnfScheduleAddData
	eventDetailLog   DnfEventDetailLog
	eventDBAction    int8
}

const (
	_ int8 = iota
	Update
	Load
	Complete
)

const (
	Wait = iota
	Start
	Stop
)

const DEFAULT_DATA_MANAGER_SIZE int = 1000
const DEFAULT_HEALTH_CHECK_SECOND int = 300

func GetDnfEventDataManager() *DnfEventDataManager {
	return &DnfEventDataManager{
		eventDataGroup: make(map[int64]DnfEventDataGroup),
		mutex:          new(sync.Mutex),
	}
}

/*
데이타 매니저 실행시 시작프로세스
*/
func (dataManager *DnfEventDataManager) Init() {
	GLogManager.Info("%s", "DataSaveGoRoutine Start...")
	dataManager.dataSaveGoRoutine()

	GLogManager.Info("%s", "Data Schedule Check Routine Start...")
	dataManager.dataScheduleCheckGoRoutine()

	GLogManager.Info("%s", "Data Receive  Routine Start...")
	dataManager.dataReceiveC(GChannelManager.GetDbToDataC())

	//DB 루틴
	GLogManager.Info("Event Data Load Start.....")
	dataManager.firstLoad()
	GLogManager.Info("Event Data Load End.....")
}

/*
메모리에 들고 있는 데이터를 주기적으로 DB업데이트 요청을 한다.
현재는 5분주기이다.
*/
func (dataManager *DnfEventDataManager) dataSaveGoRoutine() {
	ticker := time.NewTicker(DEFAULT_DB_GOROTUINE_MINUTE * time.Second)
	healthCheck := DEFAULT_HEALTH_CHECK_SECOND
	go func() {
		//이벤트 데이타를 처리하는 무한루프 구문
		for _ = range ticker.C {

			GLogManager.Debug("%s", "five minutes DataSaveRoutine")
			for logId, _ := range dataManager.eventDataGroup {
				dataManager.sendDataToDbC(Update, dataManager.eventDataGroup[logId])
			}

			time.Sleep(time.Millisecond * 1)

			//고루틴 헬스체크 5분마다 로그를 찍는다
			healthCheck = healthCheck - 1
			if 0 == healthCheck {
				healthCheck = DEFAULT_HEALTH_CHECK_SECOND
				GLogManager.Info("DataManagerHealthCheck Data Size (%d)",
					len(dataManager.eventDataGroup))
			}
			runtime.Gosched()
		}
	}()
}

/*
스케쥴을 주기적으로 체크해서 네트워크 루틴, DB루틴으로 데이타를 처리요청한다.
*/
func (dataManager *DnfEventDataManager) dataScheduleCheckGoRoutine() {
	ticker := time.NewTicker(500 * time.Millisecond)
	healthCheck := DEFAULT_HEALTH_CHECK_SECOND
	go func() {
		for t := range ticker.C {
			//패킷 채널로 서버로 전송할 패킷 정보를 넘긴다.
			dataManager.scheduleChecker(t)

			//고루틴 헬스체크 5분마다 로그를 찍는다
			healthCheck = healthCheck - 1
			if 0 == healthCheck {
				healthCheck = DEFAULT_HEALTH_CHECK_SECOND
				GLogManager.Info("DataManagerHealthCheck [dataScheduleCheckGoRoutine] Data Size (%d)",
					len(dataManager.eventDataGroup))
			}

			runtime.Gosched()
		}
	}()
}

func (dataManager *DnfEventDataManager) scheduleChecker(t time.Time) {
	for logId := range dataManager.eventDataGroup {
		scheduleData := dataManager.eventDataGroup[logId].eventSchedule
		isSendPacket := false

		switch scheduleData.eventStatus {
		case Wait:
			if scheduleData.startDate <= t.Unix() {
				scheduleData.eventStatus = Start
				scheduleData.systemSDate = t.Unix()
				isSendPacket = true
			}
		case Start:
			if scheduleData.endDate <= t.Unix() {
				scheduleData.eventStatus = Stop
				scheduleData.systemEDate = t.Unix()
				isSendPacket = true
			}
		}

		if true == isSendPacket {
			GLogManager.Debug("Send Event Packet logId(%d)", logId)
			//DB고루틴으로 DB정보를 변경 요청한다.
			dataManager.sendDataToDbC(Update, dataManager.eventDataGroup[logId])
			//패킷루틴으로 패킷을 보내라고 처리요청을 보낸다.

		}
	}
}

func (dataManager *DnfEventDataManager) isExistEventLog(logId int64) bool {
	for key := range dataManager.eventDataGroup {
		if logId == key {
			return true
		}
	}
	return false
}

func (dataManager *DnfEventDataManager) putData(eventDataGroup DnfEventDataGroup) {
	dataManager.mutex.Lock()
	dataManager.eventDataGroup[eventDataGroup.eventLog.logId] = eventDataGroup
	dataManager.mutex.Unlock()
}

//데이타를 넣어주는 채널
func (dataManager *DnfEventDataManager) dataReceiveC(c <-chan DnfEventDataGroup) {
	go func() {
		for {
			eventDataGroup := <-c
			if eventDataGroup.eventLog.logId <= 0 {
				GLogManager.Debug("It is unregistered logId(%d) by eventTable", eventDataGroup.eventSchedule.logId)
				continue
			}

			if false == dataManager.isExistEventLog(eventDataGroup.eventLog.logId) {
				GLogManager.Debug("Data Register : logId(%d)", eventDataGroup.eventLog.logId)
				dataManager.putData(eventDataGroup)
			}

			time.Sleep(time.Millisecond * 1)
			runtime.Gosched()
		}
	}()
}

func (dataManager *DnfEventDataManager) sendDataToDbC(dbAction int8, eventDataGroup DnfEventDataGroup) {
	eventDataGroup.eventDBAction = dbAction
	GChannelManager.PutDataToDbC(eventDataGroup)
	GLogManager.Debug("SendDbChannel logId(%d)", eventDataGroup.eventLog.logId)
}

func (dataManager *DnfEventDataManager) GetEventLogIdList(eventDataGroup DnfEventDataGroup) *list.List {

	logIdList := list.New()
	for logId, _ := range dataManager.eventDataGroup {
		logIdList.PushBack(logId)
	}

	return logIdList
}

func (dataManager *DnfEventDataManager) firstLoad() {
	var eventDataGroup DnfEventDataGroup
	eventDataGroup.eventDBAction = Load
	eventDataGroup.eventLog.logId = 0
	dataManager.sendDataToDbC(Load, eventDataGroup)
}
