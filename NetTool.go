package main

//网络测试工具

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sparrc/go-ping"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const VERSION = "1.0"
const PING_TIMES = 100                       //发包次数
const PING_INTERVAL = time.Millisecond * 600 //发包间隔
var PRINT_STAT_INTERVAL = 5                  //每隔几个包打印一次统计数据
const GET_SERVERS_URL = "https://network.gdutnic.com/api/get_servers"
const UPLOAD_RESULT_URL = "https://network.gdutnic.com/api/upload_result"
const VERSION_URL = "https://network.gdutnic.com/api/version"
const GATEWAY_IP = "10.30.112.1"

var pingMissionWaitGroup sync.WaitGroup
var realTTL int
var testResult = make(map[string]interface{})

func Ping(name string, host string) {
	defer pingMissionWaitGroup.Add(-1)
	pinger, _ := ping.NewPinger(host)

	pinger.OnRecv = func(pkt *ping.Packet) {
		//隔几个包输出一下当前统计状态
		if pkt.Seq%PRINT_STAT_INTERVAL == 0 { //已知BUG：如果刚好这个包丢了，就会缺少这一行
			stats := pinger.Statistics()
			//fmt.Printf("%s\t%6.3fms    %2.1f%%     %4dms     %4dms     %6.3fms     %3d     %3d\n",//win7的延迟读不到小数，所以没有必要显示小数了
			fmt.Printf("%s\t%6.3fms   %5.1f%%   %6.3fms  %7.3fms    %7.3fms     %3d     %3d\n",
				name, float32(stats.AvgRtt)/1e6, stats.PacketLoss,
				float32(stats.MinRtt)/1e6, float32(stats.MaxRtt)/1e6, float32(stats.StdDevRtt)/1e6,
				stats.PacketsSent, stats.PacketsRecv)
		}
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		//输出测试结果
		fmt.Printf("%s\t的测试结果：平均延迟%6.3fms，丢包率%5.1f%%\n", name, float32(stats.AvgRtt)/1e6, stats.PacketLoss)
		//保存数据，待上传
		testResult[name] = []float32{float32(stats.AvgRtt) / 1e6, float32(stats.PacketLoss)}
	}

	pinger.Count = PING_TIMES
	pinger.Interval = PING_INTERVAL
	pinger.Timeout = pinger.Interval * PING_TIMES //这个是整个ping测试的超时时间
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		pinger.SetPrivileged(true) //On windows, must true, or error.
	}
	pinger.Run() //开始ping
}

func PrintStat() {
	for t := 0; t < PING_TIMES/PRINT_STAT_INTERVAL; t++ {
		fmt.Println("\n描述\t 平均延迟   丢包率   最低延迟   最高延迟   延迟标准差  发包数  收包数")
		if t+1 < PING_TIMES/PRINT_STAT_INTERVAL {
			time.Sleep(PING_INTERVAL * time.Duration(PRINT_STAT_INTERVAL))
		}
	}
	time.Sleep(time.Second)
	fmt.Println()
}

func CheckTTL() {
	if runtime.GOOS == "windows" {
		//由于golang有bug，Windows下看不到ttl，所以调用系统命令行的ping来得到ttl
		cmdResult, err := exec.Command("ping", "-n", "2", GATEWAY_IP).Output()
		if err != nil {
			fmt.Println(err)
			return
		}
		cmdResultStr := string(cmdResult)
		ttlStr := cmdResultStr[strings.LastIndex(cmdResultStr, "TTL")+4 : strings.LastIndex(cmdResultStr, "TTL")+7]
		realTTL, _ = strconv.Atoi(strings.Trim(ttlStr, "\r\n "))
	} else {
		//非win系统可以正常得到ttl
		pinger, err := ping.NewPinger(GATEWAY_IP)
		if err != nil {
			panic(err)
		}
		pinger.OnRecv = func(pkt *ping.Packet) {
			realTTL = pkt.Ttl
		}
		if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
			pinger.SetPrivileged(true) //On windows, must true, or error.
		}
		pinger.Count = 2
		pinger.Interval = time.Millisecond * 300
		pinger.Timeout = time.Second
		pinger.Run()
	}
}

func StartTest() {
	//从服务器获取测试数据（运营商IP列表和描述）
	fmt.Println("正在获取运营商IP……")
	resp, err := http.Get(GET_SERVERS_URL)
	if err != nil {
		fmt.Println("服务器连接失败，无法获取运营商IP。")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("服务器返回错误码%d。\n", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var serverDatas map[string]interface{}
	err = json.Unmarshal(body, &serverDatas) //json转map
	if err != nil {
		fmt.Println(err)
	}

	//检查网关TTL
	fmt.Println("现在开始进行测试，总共时长为一分钟。")
	CheckTTL()

	//运行测试协程
	go PrintStat()
	time.Sleep(time.Millisecond * 50)
	for name, ip := range serverDatas {
		pingMissionWaitGroup.Add(1)
		go Ping(name, ip.(string))
		time.Sleep(time.Millisecond * 50) //Sleep以防print输出错乱，同时可以防止多个协程同时访问testResult
	}
	pingMissionWaitGroup.Wait()

	//上传数据
	testResult["TTL"] = realTTL
	testResult["time"] = time.Now().UnixNano()
	json_datas, _ := json.Marshal(testResult) //map转json
	http.Post(UPLOAD_RESULT_URL, "application/json;charset=utf-8", bytes.NewReader(json_datas))
}

func StartMultiTest() {
	PRINT_STAT_INTERVAL = 20 //长期测试不输出那么快，免得刷屏
	times := 1
	for {
		fmt.Printf("持续测试：第%d次。\n", times)
		StartTest()
		times++
		fmt.Println()
		time.Sleep(time.Millisecond * 500)
	}
}

func CheckVersion() bool {
	fmt.Println("正在检查软件版本……")
	resp, err := http.Get(VERSION_URL)
	if err != nil {
		fmt.Println("服务器连接失败，无法获取软件版本。")
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("服务器返回错误码%d。\n", resp.StatusCode)
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if string(body) != VERSION {
		fmt.Printf("当前软件版本为%s，最新版本为%s\n", VERSION, string(body))
		fmt.Println("请更新软件。下载地址：https://network.gdutnic.com/")
		return false
	} else {
		fmt.Println("当前软件版本为最新版！")
		return true
	}

}
