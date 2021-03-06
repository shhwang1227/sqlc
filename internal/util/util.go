package util

import (
	"encoding/json"
	"fmt"
)

func Xiazeminlog(v interface{}) {
	//return
	//debug.PrintStack()
	d, _ := json.Marshal(v)
	fmt.Println(string(d))
}
