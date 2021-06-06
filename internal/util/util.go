package util

import (
	"encoding/json"
	"fmt"
)

func Xiazeminlog(index string, v interface{}, effect bool) {
	if !effect {
		return
	}
	//debug.PrintStack()
	d, err := json.Marshal(v)
	fmt.Println(index, ":", string(d), err)
}
