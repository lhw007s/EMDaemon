package main

import (
	"errors"
	"fmt"
)

func panicProcess() (err error) {
	recv := recover()
	if recv != nil {
		fmt.Println(recv.(string))
		err = errors.New(recv.(string))
	}

	return
}
