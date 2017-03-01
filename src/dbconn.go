// dbmanager
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	_ "regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const SELECT_STRING string = "select "
const UPDATE_STRING string = "update "
const INSERT_STRING string = "insert "
const DELETE_STRING string = "delete "

const DEFAULT_QUEUE_SIZE int = 1000
const DB_NAME string = "mysql"

type IDBConn interface {
	Init() *DBManager
	SetDsn(host string, port int, account, password, db, option string) (string, error)
	Connect() (*sql.DB, error)
	SetHost(value string)
	SetPort(value int)
	SetAccount(value string)
	SetPassword(value string)
	SetDbName(value string)
	SetOption(value string)
	GetDsnStr() string
	DbClose()
	panicProcess()

	SetField(field string)
	Where(field string, value interface{})
	Where_in(field string, value ...interface{})
	Or_Where(field string, value interface{})
	Or_Where_in(field string, value ...interface{})
	Get(tableName string) (sql.Rows, error)
	SingleSet(string, interface{})
	MultiSet(map[string]interface{})
	Update(tableName string) (sql.Result, error)
	Delete(tableName string) (sql.Result, error)
	Insert(tableName string) (sql.Result, error)

	ExecSelect(query string) (sql.Rows, error)
	ExecUpdate(query string, args ...interface{}) (sql.Result, error)
	ExecDelete(query string, args ...interface{}) (sql.Result, error)
	ExecInsert(query string, args ...interface{}) (sql.Result, error)
	SetOrderBy(args ...interface{})
	Escape(value string) string
	AddQuartQueryValue(value interface{}) string
	GetQuery() string
}

type DBDsn struct {
	host       string
	port       int
	account    string
	password   string
	dbName     string
	option     string
	isDsnError bool
	dsnStr     string
}

type DBEnv struct {
	connectCount      int
	idleConnnectCount int
}

type DBManager struct {

	//queue
	Dsn   DBDsn
	DbEnv DBEnv
	//masterDb *sql.DB
	dbPool         sync.Pool
	Db             *sql.DB
	rows           *sql.Rows
	field          string
	whereValue     []interface{}
	whereQuery     string
	whereFieldCnt  int8
	orderBy        string
	queryKeyList   []string
	queryValueList []interface{}
	lastFullQuery  string
}

func GetDBManager() *DBManager {
	return &DBManager{
		dbPool:         sync.Pool{},
		field:          "",
		whereQuery:     "",
		whereFieldCnt:  0,
		orderBy:        "",
		whereValue:     make([]interface{}, 100),
		queryKeyList:   make([]string, 100),
		queryValueList: make([]interface{}, 100),
	}
}

type EventScheduleSchema struct {
	logId        int
	serverId     int
	startDate    int64
	endDate      int64
	systemSDate  int64
	systemEDate  int64
	eventType    int8
	eventStatus  int8
	eventWeekDay string
}

func (m *DBManager) Init() *DBManager {
	masterDb := GEnvManager.MasterDB["master"]
	port, _ := strconv.Atoi(masterDb["port"])
	m.SetDsn(masterDb["host"], port, masterDb["account"], masterDb["password"], masterDb["db"], masterDb["option"])

	m.DbEnv.connectCount, _ = strconv.Atoi(GEnvManager.MasterDB["master"]["connection_count"])
	m.DbEnv.idleConnnectCount, _ = strconv.Atoi(GEnvManager.MasterDB["master"]["db_idle_min_connection"])

	_, _ = m.Connect()

	return m
}

func (m *DBManager) SetDsn(host string, port int, account, password, db, option string) (string, error) {
	m.Dsn.host = host
	m.Dsn.port = port
	m.Dsn.account = account
	m.Dsn.password = password
	m.Dsn.dbName = db
	m.Dsn.option = option

	m.Dsn.isDsnError = false
	if m.Dsn.host == "" || m.Dsn.port == 0 || m.Dsn.account == "" || m.Dsn.password == "" || m.Dsn.dbName == "" {
		m.Dsn.isDsnError = true
		return m.Dsn.dsnStr, errors.New("DB DSN정보가 설정이 필요합니다.")
	}

	m.Dsn.dsnStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", m.Dsn.account, m.Dsn.password, m.Dsn.host, m.Dsn.port, m.Dsn.dbName, m.Dsn.option)

	return m.Dsn.dsnStr, nil
}

func (m *DBManager) Connect() (*sql.DB, error) {
	if len(m.Dsn.dsnStr) == 0 {
		GLogManager.WriteLog(ERR, "설정된 DSN정보가 없습니다.")
		return nil, errors.New("설정된 DSN정보가 없습니다.")
	}

	var err error

	m.Db, err = sql.Open(DB_NAME, m.Dsn.dsnStr)
	if err != nil {
		panic(err.Error())
	} else {
		m.Db.SetMaxIdleConns(m.DbEnv.idleConnnectCount)
		m.Db.SetMaxOpenConns(m.DbEnv.connectCount)
		m.Db.SetConnMaxLifetime(time.Minute * 60)
	}

	return m.Db, err
}

func (m *DBManager) DbClose() {
	if err := m.Db.Close(); nil != err {
		GLogManager.Error("DB Close Error")
	}
}

func (m *DBManager) ExecSelect(query string) (*sql.Rows, error) {
	if nil == m.Db {
		GLogManager.Error("%s", "DB Connection Error!")
		panic("DB Connection Error!")
	}
	rows, err := m.Db.Query(query)
	if err != nil {
		GLogManager.Error("%s", err)
		return nil, err
	}
	return rows, nil
}

func (m *DBManager) QueryModify(query string) (sql.Result, error) {
	defer panicProcess()
	res, err := m.Db.Exec(query)
	if err != nil {
		GLogManager.WriteLog(CRI, err.Error())
		return nil, err
	}

	return res, nil
}

func (m *DBManager) SetField(field string) error {
	if field == "" {
		return errors.New("질의어 필드를 설정하세요.")
	}

	if m.field != "" {
		m.field += "," + field
	} else {
		m.field = " " + field + " "
	}
	return nil
}

func (m *DBManager) Where(field string, value interface{}) {
	defer panicProcess()
	if field == "" {
		panic("Where문 설정되지 않았습니다.")
		return

	}
	m.setWhereAddTag(true)
	if strings.Contains(field, "<") ||
		strings.Contains(field, ">") ||
		strings.Contains(field, "<=") ||
		strings.Contains(field, ">=") {
		m.whereQuery += (field + m.AddQuartQueryValue(value))
	} else {
		m.whereQuery += (field + "=" + m.AddQuartQueryValue(value))
	}

	//m.whereQuery += (field + "=?")
	//m.whereValue = append(m.whereValue, value)
}

func (m *DBManager) Where_in(field string, value ...interface{}) {
	defer panicProcess()
	if field == "" {
		panic("Where문 설정되지 않았습니다.")
	}

	if value == nil {
		panic("Where문 설정되지 않았습니다.")
	}

	m.setWhereAddTag(true)

	m.whereQuery += (field + " in (")
	for idx := range value {
		m.whereQuery += (m.AddQuartQueryValue(value[idx]) + ",")
	}
	m.whereQuery = m.whereQuery[0 : len(m.whereQuery)-1]
	m.whereQuery += ") "
}

func (m *DBManager) Or_Where_in(field string, value ...interface{}) {
	defer panicProcess()
	if field == "" {
		panic("Where문 설정되지 않았습니다.")
	}

	if value == nil {
		panic("Where문 설정되지 않았습니다.")
	}

	m.setWhereAddTag(false)
	m.whereQuery += (field + " in (")
	for idx := range value {
		m.whereQuery += (m.AddQuartQueryValue(value[idx]) + ",")
	}
	m.whereQuery = m.whereQuery[0 : len(m.whereQuery)-1]
	m.whereQuery += ") "
}

func (m *DBManager) Or_Where(field string, value interface{}) {
	defer panicProcess()
	if field == "" {
		panic("Where문 설정되지 않았습니다.")
		return
	}

	m.setWhereAddTag(false)

	if strings.Contains(field, "<") ||
		strings.Contains(field, ">") ||
		strings.Contains(field, "<=") ||
		strings.Contains(field, ">=") {
		m.whereQuery += (field + m.AddQuartQueryValue(value))
	} else {
		m.whereQuery += (field + "=" + m.AddQuartQueryValue(value))
	}
}

/*
MysqlEscape 구현
*/
func (m *DBManager) Escape(value string) string {

	//DB escape 구현
	return ""
}

func (m *DBManager) AddQuartQueryValue(value interface{}) string {

	switch reflect.ValueOf(value).Kind() {
	case reflect.String:
		if true == strings.Contains(value.(string), "(") &&
			true == strings.Contains(value.(string), ")") {
			return value.(string)
		}

		return strconv.Quote(value.(string)) //"'" + value.(string) + "'"
	case reflect.Int:
		return strconv.Itoa(value.(int))
	case reflect.Int64:
		return strconv.FormatInt(value.(int64), 36)
	case reflect.Float64:
		return strconv.FormatFloat(value.(float64), 'f', 2, 32)
	}

	return ""
}

func (m *DBManager) fieldValueConverter(value interface{}) string {

	var fieldValue string = "="
	switch reflect.TypeOf(value).Kind() {
	case reflect.Int:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fieldValue += strconv.Itoa(value.(int))
	case reflect.String:
		v := value.(string)

		//re := regexp.MustCompile("""")
		//v = re.ReplaceAllLiteralString(v, "\"""")
		fieldValue += v
	}
	return fieldValue //field=value
}

func (m *DBManager) setWhereAddTag(isAndTag bool) {
	if m.whereQuery == "" {
		m.whereQuery += " where "
	}

	if m.whereFieldCnt > 0 {
		var addQueryTag string
		switch isAndTag {
		case true:
			addQueryTag = " and "
		case false:
			addQueryTag = " or "
		}
		m.whereQuery += addQueryTag
	}

	m.whereFieldCnt += 1
}

func (m *DBManager) Get(tableName string) (*sql.Rows, error) {
	defer panicProcess()
	if tableName == "" {
		panic("테이블명이 없습니다.")
	}

	fullQuery := SELECT_STRING + m.field + " from " + tableName + m.whereQuery

	if m.orderBy != "" {
		fullQuery += (" order by " + m.orderBy)
	}

	m.lastFullQuery = fullQuery
	GLogManager.Debug("%s", fullQuery)
	rows, err := m.ExecSelect(fullQuery)
	if err != nil {
		GLogManager.Cri("%s", fullQuery)
	}

	m.clean()
	return rows, err
}

func (m *DBManager) SetOrderBy(args ...interface{}) {
	for idx := range args {
		if m.orderBy != "" {
			m.orderBy += ","
		}
		m.orderBy += args[idx].(string) //field desc, field asc
	}
}

/*
Data : Map[필드]값
Update Field=값
*/
func (m *DBManager) MultiSet(data map[string]interface{}) {
	for key := range data {
		m.queryKeyList = append(m.queryKeyList, key)
		m.queryValueList = append(m.queryValueList, data[key])
	}
}

func (m *DBManager) SingleSet(data string, value interface{}) {
	if data == "" {
		return
	}

	m.queryKeyList = append(m.queryKeyList, data)
	m.queryValueList = append(m.queryValueList, value)
}

func (m *DBManager) clean() {
	m.field = ""
	m.whereQuery = ""
	m.whereFieldCnt = 0
	m.orderBy = ""
	m.queryKeyList = m.queryKeyList[:cap(m.queryKeyList)]
	m.queryValueList = m.queryValueList[:cap(m.queryValueList)]
	m.whereValue = m.whereValue[:cap(m.whereValue)]
}

func (m *DBManager) Update(tableName string) (sql.Result, error) {
	defer panicProcess()

	if tableName == "" {
		GLogManager.WriteLog(CRI, "테이블명을 설정하세요")
		return nil, errors.New("테이블명을 설정하세요")
	}

	updateFullQuery := UPDATE_STRING + tableName + " set "

	for key := range m.queryKeyList {
		if key > 0 {
			updateFullQuery += ","
		}
		updateFullQuery += (m.queryKeyList[key] + "=" + m.AddQuartQueryValue(m.queryValueList[key]))
	}

	if m.whereFieldCnt > 0 {
		updateFullQuery += m.whereQuery
	}

	fmt.Println(m.queryValueList)
	GLogManager.WriteLog(DEBUG, updateFullQuery)

	rows, err := m.QueryModify(updateFullQuery)
	if err != nil {
		GLogManager.Cri("%s", updateFullQuery)
	}
	m.clean()
	m.lastFullQuery = updateFullQuery
	return rows, err
}

func (m *DBManager) Insert(tableName string) (sql.Result, error) {
	defer panicProcess()

	if tableName == "" {
		GLogManager.WriteLog(CRI, "테이블명을 설정하세요")
		return nil, errors.New("테이블명을 설정하세요")
	}
	fullQueryValue := " ("
	fullQuery := INSERT_STRING + "into " + tableName + "("
	for key := range m.queryKeyList {
		if key > 0 {
			fullQuery += ","
			fullQueryValue += ","
		}

		fullQuery += m.queryKeyList[key]
		fullQueryValue += m.AddQuartQueryValue(m.queryValueList[key])
	}
	fullQuery += ") "
	fullQueryValue += ")"

	fullQuery += (" values " + fullQueryValue)

	m.lastFullQuery = fullQuery
	rows, err := m.QueryModify(fullQuery)
	if err != nil {
		GLogManager.Cri("%s", fullQuery)
	}
	m.clean()
	return rows, err
}

func (m *DBManager) Delete(tableName string) (sql.Result, error) {
	defer panicProcess()
	if tableName == "" {
		GLogManager.WriteLog(CRI, "테이블명을 설정하세요")
		return nil, errors.New("테이블명을 설정하세요")
	}

	deleteFullQuery := DELETE_STRING + "from " + tableName
	if m.whereFieldCnt > 0 {
		deleteFullQuery += m.whereQuery
	}

	m.lastFullQuery = deleteFullQuery
	rows, err := m.QueryModify(deleteFullQuery)
	if err != nil {
		GLogManager.Cri("%s", deleteFullQuery)
	}
	return rows, err
}

func (m *DBManager) SendPing() error {
	if err := m.Db.Ping(); nil != err {
		GLogManager.Error("DB Ping Error")
		return err
	}
	return nil
}

func (m *DBManager) GetLastFullQuery() string {
	return m.lastFullQuery
}

func (m *DBManager) PanicProcess() {
	recv := recover()
	if recv != nil {
		m.Db.Close()
	}
}

func (m *DBDsn) SetHost(value string) {
	m.host = value
}

func (m *DBDsn) SetPort(value int) {
	m.port = value
}

func (m *DBDsn) SetAccount(value string) {
	m.account = value
}

func (m *DBDsn) SetPassword(value string) {
	m.password = value
}

func (m *DBDsn) SetDbName(value string) {
	m.dbName = value
}

func (m *DBDsn) SetOption(value string) {
	m.option = value
}
