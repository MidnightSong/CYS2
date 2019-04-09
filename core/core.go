package core

import (
	"CYS2/stack"
	"errors"
	"fmt"
	"github.com/go-resty/resty"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var JpgUrls = stack.New() //用来保存获取到的作者套图页面的URL
var ThreadSync sync.WaitGroup
var count int
var client = resty.New()
var Group string

func init() {
	client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36")
	client.SetHeader("Connection", "keep-alive")
	client.SetHeader("Accept", "*/*")
	client.SetHeader("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,zh-TW;q=0.6,ja;q=0.5")

	jar, _ := cookiejar.New(nil)
	client.SetCookieJar(jar)

	client.SetRedirectPolicy(resty.DomainCheckRedirectPolicy("img.ciyo.cn"))
}

/***
传入网址
@return 	result 获取到的页面数据
				nextUrl 该网页下一页的网址
				err 	错误
*/
func GetMainPage(url string) (result, nextUrl string, err error) {
	client.SetRetryCount(2)
	fmt.Printf("正在打开页面：%s\n", url)
	resp, err := client.R().Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode() != 200 {
		err = errors.New("GetMainPage server error:" + string(resp.StatusCode()))
		log.Println("GetMainPage server error:\n", string(resp.Body()))
		return
	}

	result = string(resp.Body())

	next := resp.Header().Get("CY-NextUrl")
	nextUrl = "http://www.ciyo.cn" + next
	err = nil
	return
}

/**
把主页面数据中的子页面网址提取到JpgUrls中保存
@pram 主页面数据
*/
func GetChildUrl(mainPage string) (err error) {
	compile := regexp.MustCompile(`<a href="(?s:(.*?))">`)
	if compile == nil {
		return errors.New("函数GetChildPage() regexp compile error")
	}
	childUrls := compile.FindAllStringSubmatch(mainPage, -1) //过滤网页内容，提取子页面网址，-1代表过滤全部
	for _, data := range childUrls {
		childUrl := data[1]
		if childUrl == "" {
			return errors.New("未获取到mainPage中URL地址")
		}
		JpgUrls.Push("http://www.ciyo.cn" + childUrl)
	}
	return nil
}

/**
获取作者套图页面数据
*/
func GetJpgPage() (err error) { //多线程

	defer ThreadSync.Done()
	var author string //保存作者名字
	s := JpgUrls.Pop()
	client.R().SetHeader("Upgrade-Insecure-Requests", "1")
	client.R().SetHeader("Host", "www.ciyo.cn")
	client.R().SetHeader("Referer", "http://www.ciyo.cn/")
	resp, err := client.R().Get(s)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if resp.StatusCode() != 200 {
		log.Println(string(resp.Body()))
		log.Println("GetJpgPage server error")
		return
	}

	body := resp.Body()
	res := string(body)

	//提取作者的名字
	name, e := regexp.Compile(`-name">(?s:(.*?))</span>`)
	if e != nil {
		e = errors.New("func GetJpgPage0 regexp compile error")
		return
	}
	realName := name.FindAllStringSubmatch(res, 1)

	for _, data := range realName {
		author = data[1]
		break
	}
	fmt.Println("作者名称：" + author)

	//提取图片地址
	compile, e := regexp.Compile(`class="images"(?s:(.*?))</ul>`)
	if e != nil {
		e = errors.New("func GetJpgPage1 regexp compile error")
		return
	}
	jpgs := compile.FindAllStringSubmatch(res, 1)
	var remix string
	for _, data := range jpgs {
		remix = data[1]
		break
	}
	compile2, e := regexp.Compile(`src="(?s:(.*?))" />`)
	if e != nil {
		e = errors.New("func GetJpgPage2 regexp compile error")
		return
	}
	var realimg []string
	jpgs2 := compile2.FindAllStringSubmatch(remix, -1)
	var mix string
	for _, data := range jpgs2 {
		mix = data[1]
		if mix == "" {
			strings.Trim(mix, " ")
			log.Println("未提取取到realJpg地址：\n" + res)
			return errors.New("未提取取到realJpg地址：\n" + res)
		}
		realimg = append(realimg, mix)
	}

	//调用获取真实的一张图片的接口

	for i := 0; i < len(realimg); i++ {
		ThreadSync.Add(1)
		go GetRealJpg(author, realimg[i])
	}
	return
}

func GetRealJpg(author string, imgUrl string) { //多线程
	time.Sleep(time.Second)
	defer ThreadSync.Done()
	client.R().SetHeader("Upgrade-Insecure-Requests", "1")
	client.R().SetHeader("Host", "qn.ciyocon.com")
	client.R().SetHeader("Cache-Control", "max-age=0")
	resp, err := client.R().Get(imgUrl)

	if err != nil {
		log.Println(imgUrl, "错误问题", err.Error())
		return
	}

	if resp.StatusCode() != 200 {
		log.Println("GetRealJpg server error code:", resp.StatusCode(), imgUrl, string(resp.Body()))
		return
	}
	WriteFile(author, resp.Body())
}

func WriteFile(author string, stream []byte) (err error) {
	_, ee := os.Stat("./Download")
	if ee != nil {
		os.Mkdir("./Download", os.ModePerm) //创建Download文件夹
	}

	if _, ee := os.Stat("./Download/" + Group); ee != nil {
		os.Mkdir("./Download/"+Group, os.ModePerm) //创建分类文件夹
	}
	_, ee = os.Stat("./Download/" + author)
	if ee != nil {
		os.Mkdir("./Download/"+Group+"/"+author, os.ModePerm) //以作者名字创建目录
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	filename := strconv.Itoa(r.Intn(999999))
	e := ioutil.WriteFile("./Download/"+Group+"/"+author+"/"+filename+".jpg", stream, 0666)
	if e != nil {
		fmt.Println(e.Error())
	}
	return
}
