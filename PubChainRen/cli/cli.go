package cli

import (
	"PubChainRen/utils"
	"flag"
	"os"
	"fmt"

)

/**
 * 命令行交互结构体
 */
type CommandLine struct {
	NODE_ID string
}

/**
 * 命令行
 */
func (cli *CommandLine) Run() {

	//解析参数合法性
	IsArgsValid()

	/**
	 * 创建钱包
	 */
	createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)

	/**
	 * 打印钱包地址列表
	 */
	printAddressCmd := flag.NewFlagSet("printAddressList", flag.ExitOnError)

	/**
	 * 创建创世区块
	 */
	createBlockChainCmd := flag.NewFlagSet("createBlockChain", flag.ExitOnError)
	coinbaseAddress := createBlockChainCmd.String("address", "", "创世区块Coinbase交易地址")

	/**
	 * 转账功能
	 */
	sendCmd := flag.NewFlagSet("sendTransaction", flag.ExitOnError)
	fromData := sendCmd.String("from", "", "转账的发起者数据")
	toData := sendCmd.String("to", "", "转账的接收者数据")
	amoutData := sendCmd.String("amount", "", "转账的金额数据")

	/**
	* 打印所有的区块数据
	*/
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)

	/**
	 * 查询地址余额
	 */
	getBalaceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	address := getBalaceCmd.String("address", "", "要查询的地址")

	/**
	 * 启动节点程序
	 */
	startNodeCmd := flag.NewFlagSet("startNode", flag.ExitOnError)
	//rewardAddress := startNodeCmd.String("address", "", "指定奖励地址")

	switch os.Args[1] {
	case "createWallet": //创建区块链钱包
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err.Error())
		}
		if createWalletCmd.Parsed() {
			cli.createWallet()
		}
	case "printAddressList": //打印地址列表数据
		err := printAddressCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err.Error())
		}
		if printAddressCmd.Parsed() {
			cli.PrintAddressLists()
		}
	case "createBlockChain": //创建区块链，指定创世区块数据
		err := createBlockChainCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err.Error())
		}
		if createBlockChainCmd.Parsed() {
			cli.CreateBlockChain(*coinbaseAddress)
		}
	case "sendTransaction": //添加转账交易数据
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err.Error())
		}

		//判断转账的发起者 接收者 及转账数据不能为空
		if *fromData == "" || *toData == "" || *amoutData == "" {
			fmt.Println("请输入转账参数")
			CommandUsage()
			os.Exit(1)
		}

		if sendCmd.Parsed() {
			fromArray := utils.JSONToStringArray(*fromData)
			toArray := utils.JSONToStringArray(*toData)
			amountArray := utils.JSONToIntArray(*amoutData)
			cli.SendTransaction(fromArray, toArray, amountArray)
		}
	case "printChain": //打印所有区块
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err.Error())
		}
		if printChainCmd.Parsed() {
			cli.PrintChain()
		}
	case "getBalance":

		if err := getBalaceCmd.Parse(os.Args[2:]); err != nil {
			panic(err.Error())
		}
		cli.GetBalance(*address)
	case "startNode":
		if err := startNodeCmd.Parse(os.Args[2:]); err != nil {
			panic(err.Error())
		}
		cli.StartNode()
	default:
		//所有命令都没有解析到，打印命令交互清单程序就退出
		CommandUsage()
		os.Exit(1)
	}
}

/**
 * 判断参数是否合法
 */
func IsArgsValid() {
	if len(os.Args) < 2 {
		CommandUsage()
		os.Exit(1)
	}
}

/**
 * 功能命令清单交互说明
 */
func CommandUsage() {
	fmt.Println("The Support Command of the Chain and the Usage list：")
	fmt.Println()
	fmt.Println("\tcreateWallet\t创建钱包")
	fmt.Println()
	fmt.Println("\tprintAddressList\t输出钱包列表")
	fmt.Println()
	fmt.Println("\tcreateBlockChain\t创建区块链命令")
	fmt.Println("\t参数：-address\t创世区块Coinbase交易地址")
	fmt.Println()
	fmt.Println("\tsendTransaction\t添加转账交易")
	fmt.Println("\t参数：-from\t转账的发起者")
	fmt.Println("\t参数：-to\t转账的接收者")
	fmt.Println("\t参数：-amount\t转账的金额数据")
	fmt.Println()
	fmt.Println("\tprintChain\t打印区块全部区块信息")
	fmt.Println()
	fmt.Println("\tgetBalance\t查询某个地址的余额")
	fmt.Println("\t参数：-address\t要查询的地址")
	fmt.Println()
	fmt.Println("\tstartNode\t启动节点程序")
	fmt.Println("\t参数：-address\t指定奖励的地址")
}
