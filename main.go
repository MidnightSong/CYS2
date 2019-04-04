package main

import (
	"CYS2/core"
	"fmt"
	"time"
)

func main() {
	var num int
	fmt.Printf("输入起始页（>=1）:")
	fmt.Scan(&num)

	result, nextUrl, err := core.GetMainPage("http://www.ciyo.cn/home_posts?group=COS")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = core.GetChildUrl(result) //获取到数据保存在core.JpgUrls中
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//爬取大于1页时执行
	var res = &result
	var next = &nextUrl
	for i := 1; i < num; i++ {
		*res, *next, err = core.GetMainPage(nextUrl)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		core.GetChildUrl(result) //获取到数据保存在core.JpgUrls中
	}
	th := core.JpgUrls.Len()
	for i := 0; i < th; i++ {
		core.ThreadSync.Add(1)
		go core.GetJpgPage()
	}
	core.ThreadSync.Wait()
	time.Sleep(time.Second * 5)
	fmt.Println("执行完毕")
}
