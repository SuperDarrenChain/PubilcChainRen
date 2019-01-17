package main

import (
	"PubChainRen/cli"
	"os"
	"fmt"
)

/**
 * 程序主入口
 */
func main() {

	node_id := os.Getenv("NODE_ID")
	if node_id == "" {
		fmt.Println("没有配置NODE_ID,请先配置NODE_ID")
		os.Exit(1)
	}
	fmt.Println(node_id)

	cli := &cli.CommandLine{node_id}
	cli.Run()

}
