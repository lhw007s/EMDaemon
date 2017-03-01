package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type IEnvManager interface {
	Init()
	LoadProperties(src string, data *map[string]map[string]string) (e error)
}

type EnvManager struct {
	MasterDB map[string](map[string]string)
	Log      map[string](map[string]string)
}

func NewEnvManager() *EnvManager {
	return &EnvManager{
		make(map[string](map[string]string)),
		make(map[string](map[string]string)),
	}
}

func (envManager *EnvManager) Init() {
	src := "./cfg/db.properties"
	if err := envManager.LoadProperties(src, &envManager.MasterDB); nil != err {
		panic(err)
	}

	src = "./cfg/log.properties"
	if err := envManager.LoadProperties(src, &envManager.Log); nil != err {
		panic(err)
	}
}

func (envManager *EnvManager) LoadProperties(src string, data *map[string]map[string]string) (e error) {

	if data == nil {
		panic("LoadProperties data Object is Nil")
	}

	defer panicProcess()

	f, err := os.Open(src)
	if err != nil {
		panic("파일이 존재하지 않습니다.")
		return
	}

	scanner := bufio.NewScanner(f)
	if scanner == nil {
		return
	}

	var tagPattern string = "\\[(.*?)\\]"
	var valuePattern string = "([\\w\\W]*?)=([\\w\\W]*)"

	var isKeySet bool = false
	for scanner.Scan() {
		if len(strings.Trim(scanner.Text(), " ")) == 0 {
			continue
		}
		fmt.Println(scanner.Text())
		var keyStr string = ""
		matched, _ := regexp.MatchString(tagPattern, scanner.Text())
		if matched == true {
			pattern, _ := regexp.Compile(tagPattern)
			result := pattern.FindAllStringSubmatch(scanner.Text(), -1)

			if result != nil {
				fmt.Println(result)
				keyStr = strings.Trim(result[0][1], " ")
				fmt.Println(keyStr)
				fmt.Println(result[0][1])
			}
			isKeySet = true
		}

		if isKeySet == true {
			(*data)[keyStr] = make(map[string]string)
			for scanner.Scan() {
				var key string
				var value string
				pattern := regexp.MustCompile(valuePattern)
				result := pattern.FindStringSubmatch(scanner.Text())
				if result != nil {
					key = strings.Trim(result[1], " ")
					value = strings.Trim(result[2], " ")
					(*data)[keyStr][key] = value
				} else {
					break
				}
			}
			isKeySet = false
		}
	}

	e = nil
	return
}
