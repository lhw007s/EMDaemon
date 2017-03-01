// DBTest project DBTest.go
package main

import (
	_ "database/sql"
	_ "fmt"
	_ "strings"
	_ "time"

	_ "github.com/go-sql-driver/mysql"
)

type reserverList struct {
	idx          int
	no           int
	reserverTime int64
}

func debug() {
	//var reserverList []reserverList = make([]reserverList, MAX_RESERVER_LIST)

	//db := connect()
	//load(db, &reserverList)

	/*
			for {
				fmt.Printf("%2d, %2d\n", time.Now().Minute(), time.Now().Second())
				time.Sleep(time.Second * 1)
			}

				var rule string = "YMDhms"
		var date string =
	*/
	//var timeDebug TimeManager
	//t, err := timeDebug.StrToTime("Y-M-D h:m:s", "2018-02-21 01:20:40")
	//t, err := timeDebug.Parser("", "2018-02-21 01:20:40")

	//var currTime *time.Time = new(time.Time)
	//	currTime = &time.Now()

	/*
		var a time.Time
		a = time.Now()
		t, err := timeDebug.TimeToStr("Y-M-D h:m:s", &a)
		fmt.Println(r)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(t)
	*/
	//	src := "./cfg/config.properties"
	//data := make(map[string](map[string]string))

	/*
		err := stringTool.loadProperties(src, &data)

		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(data)
			fmt.Println(data["server"])
		}
	*/
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
