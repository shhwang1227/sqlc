package util

import (
	"encoding/json"
	"fmt"
)

func Xiazeminlog(index string, v interface{}) {
	return
	//debug.PrintStack()
	d, _ := json.Marshal(v)
	fmt.Println(index, ":", string(d))
}
