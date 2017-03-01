package main

import (
	"fmt"
)

type IQueue interface {
}

type QManager struct {
	//DbMsg []
	//PacketMsg
}

func SimpleChannel() {
	c := func() <-chan int {
		c := make(chan int)
		go func() {
			defer close(c)
			c <- 1
			c <- 2
			c <- 3
		}()
		return c
	}()

	for num := range c {
		fmt.Println(num)
	}

}
