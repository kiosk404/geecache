package test

import (
	"testing"
	"time"
	"fmt"
	"sync"
)

var set = make(map[int]bool, 0)
var m sync.Mutex

func printOnce(num int) {
	if _, exist := set[num]; !exist {
		fmt.Println(num)
	}
	set[num] = true
}

func printOnceSafe(num int) {
	m.Lock()
	if _, exist := set[num]; !exist {
		fmt.Println(num)
	}
	set[num] = true
	m.Unlock()
}

func Test_Concurrent(t *testing.T) {
	for i := 0; i < 10; i++ {
		go printOnce(100)
	}
	time.Sleep(time.Second * 30)
}


func Test_Core(t *testing.T) {
	fmt.Println(2<<10)
}



