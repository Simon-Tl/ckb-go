package main

import (
	"ckb-go/transactions"
	"fmt"
)

func main() {
	err, res, _ := transactions.QuerySudtAmount("ckt1qzda0cr08m85hc8jlnfp3zer7xulejywt49kt2rr0vthywaa50xwsqgfll7ltphp84wlxulv3vczzt9z5x9dwnsqracel")
	fmt.Println(res)
	if err != nil {
		fmt.Println(err)
	}
}
