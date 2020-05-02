package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type VoiceMsg struct {
	Ret  int       `json:"ret"`
	Data VoiceNode `json:"data"`
}
type VoiceNode struct {
	TrackId         int    `json:"trackId"`
	CanPlay         bool   `json:"canPlay"`
	IsPaid          bool   `json:"isPaid"`
	HasBuy          bool   `json:"hasBuy"`
	Src             string `json:"src"`
	AlbumIsSample   bool   `json:"albumIsSample"`
	SampleDuration  int    `json:"sampleDuration"`
	IsBaiduMusic    bool   `json:"isBaiduMusic"`
	FirstPlayStatus bool   `json:"firstPlayStatus"`
}

func main() {
	// Instantiate default collector
	c := colly.NewCollector()

	// On every a element which has href attribute call callback
	//c.OnHTML("a[href]", func(e *colly.HTMLElement) {
	//	link := e.Attr("href")
	//	title := e.Attr("title")
	// Print link
	// fmt.Printf("Link found: %s -> %s\n", title, link)
	// Visit link found on page
	// Only those links are visited which are in AllowedDomains
	//c.Visit(e.Request.AbsoluteURL(link))
	//})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnResponse(func(resp *colly.Response) {
		fmt.Println("response received", resp.StatusCode)

		// goquery直接读取resp.Body的内容
		htmlDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))

		// 读取url再传给goquery，访问url读取内容，此处不建议使用
		// htmlDoc, err := goquery.NewDocument(resp.Request.URL.String())

		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("hello")
		// 找到抓取项 <div class="text _Vc"><a title="下册Lesson 1" href="/waiyu/6906831/31494061"><span class="title _Vc">下册Lesson 1</span></a></div> 的所有的a解析
		htmlDoc.Find("div.text._Vc a").Each(func(i int, s *goquery.Selection) {

			//time.Sleep(5000 * time.Millisecond)

			link, _ := s.Attr("href")
			title := s.Text()

			exp := regexp.MustCompile(`(\d+)`)
			params := exp.FindStringSubmatch(title)

			var mediaFileNamePrefix string
			var lessonNum int
			if len(params) < 2 {
				mediaFileNamePrefix = title
			} else {
				if lessonNum, err = strconv.Atoi(params[1]); err != nil {
					mediaFileNamePrefix = title
				} else {
					mediaFileNamePrefix = "Lesson" + fmt.Sprintf("%02d", lessonNum)
				}
			}
			fmt.Printf("link %d: %s - %s - %s\n", i, link, title, mediaFileNamePrefix)

			exp = regexp.MustCompile(`\/(\d+)$`)
			params = exp.FindStringSubmatch(link)
			//for _,param :=range params {
			//	fmt.Println(param)
			//}
			audio_id := params[1]
			//return

			// https://www.ximalaya.com/revision/play/v1/audio?id=31494061&ptype=1
			url := "https://www.ximalaya.com/revision/play/v1/audio?id=" + audio_id + "&ptype=1"
			//fmt.Println(url)
			//resp, err := http.Get(url)
			client := &http.Client{}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				// handle error'
				panic(err)
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
			//程序在使用完回复后必须关闭回复的主体。

			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			//fmt.Prin	tln(string(body))
			fmt.Printf("[Req][Url] %s\n", url)
			fmt.Printf("[Resp][ContentLength] %d\n", resp.ContentLength)
			fmt.Printf("[Resp][Header] %s\n", resp.Header)
			fmt.Printf("[Resp][StatusCode] %d\n", resp.StatusCode)
			defer resp.Body.Close()

			var vm VoiceMsg
			err = json.Unmarshal(body, &vm)
			if err != nil {
				fmt.Println("error:", err)
			}
			fmt.Println(vm.Data.Src)

			// download file
			url = vm.Data.Src
			if req, err = http.NewRequest("GET", url, nil); err != nil {
				// handle error'
				panic(err)
			}

			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")

			//resp, err = http.Get(url)
			//if err != nil {
			//	panic(err)
			//}

			resp, err = client.Do(req)
			if err != nil {
				panic(err)
			}
			//程序在使用完回复后必须关闭回复的主体。

			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("下载失败:%s\n", err)
				panic(err)
			}
			fmt.Printf("[Req][Url] %s\n", url)
			fmt.Printf("[Resp][ContentLength] %d\n", resp.ContentLength)
			fmt.Printf("[Resp][Header] %s\n", resp.Header)
			fmt.Printf("[Resp][StatusCode] %d", resp.StatusCode)

			defer resp.Body.Close()

			//filename := strconv.Itoa(vm.Data.TrackId)+".m4a"
			filename := mediaFileNamePrefix + ".m4a"
			err = ioutil.WriteFile(filename, body, 0666)
			if err != nil {
				fmt.Printf("下载失败:%s\n", err)
				panic(err)
			}
			fmt.Printf("!!!下载成功:%s 到：%s\n", url, filename)

			//f, err := os.Create(strconv.Itoa(vm.Data.TrackId)+".m4a")
			//if err != nil {
			//	panic(err)
			//}
			//io.Copy(f, resp.Body)
			//defer f.Close()
			//c.Visit(band)
		})

	})

	// Start scraping on https://hackerspaces.org
	c.Visit("https://www.ximalaya.com/waiyu/6906831/")
}
