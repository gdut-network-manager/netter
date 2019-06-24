// +build windows

package main

import (
	"fmt"
)

func main() {
	if CheckVersion() { //检查版本更新
		fmt.Println("欢迎使用GDUT网络测试工具！")
		fmt.Println("1.单次测试（约需一分钟）")
		fmt.Println("2.持续测试（需要手动关闭窗口来终止）")
		fmt.Println("提示：可以按Ctrl+Shift+H隐藏窗口，按Ctlr+Shift+S显示窗口（可能需要等半分钟才生效）")
		go HotkeyHandel()
		fmt.Printf("请输入数字并按回车：")
		var choice int
		fmt.Scanf("%d", &choice)
		fmt.Println()
		switch choice {
		case 1:
			StartTest()
		case 2:
			StartMultiTest()
		case 3:
			fmt.Println()
		}
	}
	fmt.Println("\n按回车退出。")
	fmt.Scanln()
	fmt.Scanln() //windows要读多一行，不然会直接退出
}
