package util

import (
	"fmt"
	"encoding/json"
)

func Xiazeminlog(v interface{}){
	d,_:=json.Marshal(v)
	fmt.Println(string(d))
}