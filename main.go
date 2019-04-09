package main

import (
	"CYS2/core"
	"fmt"
	"log"
	"time"
)

func main() {
	var group string
	var num int
	fmt.Printf("将从次元社爬去照片\n从哪个分类下爬去图片：1.精选  2.萌妹  3.交友  4.日常  5.绘画  6.COS  7.兴趣  8.壁纸头像\n")
	fmt.Scan(&num)
	switch num {
	case 1:
		group = "精选"
	case 2:
		group = "萌妹"
	case 3:
		group = "交友"
	case 4:
		group = "日常"
	case 5:
		group = "绘画"
	case 6:
		group = "COS"
	case 7:
		group = "兴趣"
	case 8:
		group = "壁纸头像"
	}
	fmt.Printf("输入爬取多少页（>=1）:")
	fmt.Scan(&num)

	go spinner(100 * time.Millisecond)

	result, nextUrl, err := core.GetMainPage("http://www.ciyo.cn/home_posts?group=" + group)
	if err != nil {
		log.Println(err.Error())
		return
	}
	err = core.GetChildUrl(result) //获取到数据保存在core.JpgUrls中
	if err != nil {
		log.Println(err.Error())
		return
	}
	//爬取大于1页时执行
	var res = &result
	var next = &nextUrl
	for i := 1; i < num; i++ {
		*res, *next, err = core.GetMainPage(nextUrl)
		if err != nil {
			log.Println(err.Error())
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
	fmt.Println("执行完毕，图片保存在当前文件夹下Download文件夹下")
	time.Sleep(time.Second * 5)
}

func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}
