/**
 * @Author: YMBoom
 * @Description:
 * @File:  main
 * @Version: 1.0.0
 * @Date: 2023/01/12 9:42
 */
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"jd-ymg-coupon/config"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func main() {
	config.ReadConfig()
	if config.Get() == nil {
		fmt.Println("获取配置失败")
		return
	}

	accounts := getAccounts()
	if len(accounts) == 0 {
		fmt.Println("获取账号失败")
		return
	}

	duration := diffDuration()
	if duration == 0 {
		fmt.Println("时间不正确 检查config.go中at")
		return
	}
	fmt.Println("在", duration, "后抢购")

	// 定时
	timer := time.NewTimer(duration)
	<-timer.C

	fmt.Println("======================= 开始 =======================")
	for _, a := range accounts {
		do(&a)
	}
}

type account struct {
	Cookie    string
	Random    string
	ExtraData string
}

type extraData struct {
	Random    string
	ExtraData string
}

func getAccounts() []account {
	var accounts []account
	// 每个extraData只能请求4次
	cfg := config.Get()
	totalCount := len(cfg.ExtraData) * 4

	var ext []extraData
	for _, data := range cfg.ExtraData {

		for i := 0; i < 4; i++ {
			ext = append(ext, extraData{
				Random:    data.Random,
				ExtraData: data.ExtraData,
			})
		}
	}

	// 每个账号请求的次数
	count := int(math.Floor(float64(totalCount / len(cfg.Cookies))))

	for _, cookie := range cfg.Cookies {
		for i := 0; i < count; i++ {
			data := ext[0]
			accounts = append(accounts, account{
				Cookie:    cookie,
				Random:    data.Random,
				ExtraData: data.ExtraData,
			})
			ext = ext[1:]
		}
	}
	return accounts
}

type Body struct {
	Extend       string `json:"extend"`
	RcType       string `json:"rcType"`
	Source       string `json:"source"`
	Random       string `json:"random"`
	CouponSource string `json:"couponSource"`
	ExtraData    string `json:"extraData"`
}

// 定时时间
func diffDuration() time.Duration {
	today := time.Now().Format("2006-01-02")

	t, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %s:00", today, config.Get().At), time.Local)
	if err != nil {
		return 0
	}

	jdTime := jdServerTime()
	diff := t.Sub(time.UnixMilli(jdTime))

	early := time.Duration(config.Get().Early) * time.Millisecond
	if diff < early {
		return 0
	}
	return diff - early
}

func headers(cookie string) http.Header {
	header := http.Header{}
	header.Add("Cookie", cookie)
	header.Add("Content-Type", config.Get().ContentType)
	header.Add("Referer", config.Get().Referer)
	header.Add("User-Agent", config.Get().UserAgent)
	return header
}

func reqUrl() string {
	return fmt.Sprintf("%s?functionId=%s", config.Get().Domain, config.Get().FunctionId)
}

func payload(a *account) io.Reader {
	body := Body{
		Extend:       config.Get().Extend,
		RcType:       "1",
		Source:       "couponCenter_app",
		Random:       a.Random,
		CouponSource: "1",
		ExtraData:    strings.Replace(a.ExtraData, "\\", "", -1),
	}

	b, _ := json.Marshal(body)

	val := url.Values{}
	val.Set("body", string(b))
	val.Set("appid", config.Get().Appid)
	val.Set("uuid", config.Get().Uuid)
	val.Set("client", config.Get().Client)
	val.Set("monitorSource", config.Get().MonitorSource)

	return strings.NewReader(val.Encode())
}

func do(a *account) {
	httpReq, _ := http.NewRequest(http.MethodPost, reqUrl(), payload(a))
	httpReq.Header = headers(a.Cookie)

	c := &http.Client{}
	response, err := c.Do(httpReq)
	if err != nil {
		fmt.Println("请求发生错误", err)
		return
	}

	defer response.Body.Close()
	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("解析返回值错误", err)
		return
	}
	fmt.Println("抢券返回:", string(all))
}

type jdServerTimeResponse struct {
	CurrentTime  string `json:"currentTime"`
	CurrentTime2 string `json:"currentTime2"`
	ReturnMsg    string `json:"returnMsg"`
	Code         string `json:"code"`
	SubCode      string `json:"subCode"`
}

func jdServerTime() (t int64) {
	t = time.Now().UnixMilli()
	httpReq, err := http.NewRequest(http.MethodGet, "https://api.m.jd.com/client.action?functionId=queryMaterialProducts&client=wh5", nil)
	if err != nil {
		return
	}
	httpReq.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36")

	c := &http.Client{}
	response, err := c.Do(httpReq)
	if err != nil {
		return 0
	}

	defer response.Body.Close()

	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	serverTime := new(jdServerTimeResponse)
	if err = json.Unmarshal(all, serverTime); err != nil {
		return
	}

	if serverTime.CurrentTime2 == "" {
		return
	}

	currentTime, err := strconv.Atoi(serverTime.CurrentTime2)
	if err != nil {
		return
	}

	return int64(currentTime)

}
