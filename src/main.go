// HellCmd project main.go
package main

import (
	"fmt"
	_ "os"
	"reflect"
	_ "runtime"
	_ "sync"
	_ "syscall"
	"time"
)

var GChannelManager *ChannelManager = NewChannelManager()

type ChannelManager struct {
	dataToDbC chan DnfEventDataGroup
	dbToDataC chan DnfEventDataGroup
}

func NewChannelManager() *ChannelManager {
	return &ChannelManager{
		dataToDbC: make(chan DnfEventDataGroup),
		dbToDataC: make(chan DnfEventDataGroup),
	}
}

func (channelManager *ChannelManager) PutDbToDataC(eventDataGroup DnfEventDataGroup) {
	channelManager.dbToDataC <- eventDataGroup
}

func (channelManager *ChannelManager) PutDataToDbC(eventDataGroup DnfEventDataGroup) {
	channelManager.dataToDbC <- eventDataGroup
}

func (channelManager *ChannelManager) GetDataToDbC() chan DnfEventDataGroup {
	return channelManager.dataToDbC
}

func (channelManager *ChannelManager) GetDbToDataC() chan DnfEventDataGroup {
	return channelManager.dbToDataC
}

func main() {
	//전역변수 초기화
	gGlobalScope.Init()

	for {
		time.Sleep(1 * time.Millisecond)
	}
}

func test1(args ...interface{}) {
	test(args...)
}

func test(args ...interface{}) {
	for _, value := range args {
		fmt.Println(value)
		v := reflect.TypeOf(value)
		switch v.Kind() {
		case reflect.Int:
			fmt.Println("정수")
		case reflect.String:
			fmt.Println("문자")
		case reflect.Float64:
			fmt.Println("소수")
		}
	}

	fmt.Printf("%s %d %f", args...)
}
