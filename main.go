package main

import (
	"fmt"
	"gopher/SyncMap"
)

func main() {
	rwMap := SyncMap.NewRWMutexMap[int, string]()
	rwMap.Set(2, "你好")
	value, err := rwMap.Get(2)
	if err != nil {
		fmt.Println("error")
		return
	}
	fmt.Println(value)
	rwMap.Set(1, "a")
	rwMap.Set(2, "b")
	rwMap.Set(3, "c")
	a := rwMap.GetKeys()
	fmt.Println(a)
	fmt.Println("chanMap------------------------------------------")
	chanMap := SyncMap.NewChannelMap[int, string]()
	chanMap.Set(1, "a")
	value, err = chanMap.Get(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(value)
	chanMap.Set(2, "b")
	chanMap.Delete(1)
	chanMap.Set(1, "z")
	b := chanMap.GetKeys()
	fmt.Println(b)
	for _, key := range b {
		fmt.Println(key)
		value, err = chanMap.Get(key)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(value)
	}
}
