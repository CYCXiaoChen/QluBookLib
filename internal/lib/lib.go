package lib

import (
	"cklib/internal/lib/model"
	"cklib/pkg/logger"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var Mlog = logger.New(nil, logger.LDEBUG, 0)

type Lib struct {
	Host       string
	C          *http.Client
	BookTimeId int
	token      string
}

func NewLib() *Lib {
	lib := &Lib{
		Host: "yuyue.lib.qlu.edu.cn",
		C:    &http.Client{
			//Timeout: time.Second * 5,
		},
	}
	lib.Updatetime()
	return lib
}

func (l *Lib) Updatetime() {
	//1499055 2022-10.26
	time1, _ := time.ParseInLocation("2006-01-02", "2022-10-26", time.Local)
	sub := time.Now().Local().Sub(time1)

	//fmt.Println(time.Now().Local(), sub)
	l.BookTimeId = 1551663 + int(sub.Hours()/24)
	//fmt.Println(time.Now().Local(), sub, l.BookTimeId)
}

func (l *Lib) Login(username string, passwd string) (bool, error) {

	url := "http://yuyue.lib.qlu.edu.cn/api.php/login"

	hash := md5.New()
	hash.Write([]byte(passwd))
	sum := hash.Sum(nil)
	strmd5 := hex.EncodeToString(sum)
	body := fmt.Sprintf("username=%s&password=%s&from=mobile", username, strmd5)
	bodyreader := strings.NewReader(body)

	request, _ := http.NewRequest("POST", url, bodyreader)
	setHeader(request)
	resp, err := l.C.Do(request)
	//resp, err := l.C.Post(url, "application/x-www-form-urlencoded; charset=UTF-8", bodyreader)
	if err != nil {
		return false, err
	}
	var respList model.Loginresp
	readAll, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(readAll, &respList)
	if err != nil {
		return false, err
	}
	//fmt.Printf("login:%s,%s\n", username, passwd)

	//fmt.Println(respList)
	if respList.Status != 1 {
		Mlog.PF(logger.LDEBUG, "%v", respList)
		return false, errors.New(respList.Msg)
	}
	l.token = respList.Data.Hash.AccessToken
	return true, nil
}

func (l *Lib) UpdatebookTimeId() (int, error) {
	//url_segment  = 'http://yuyue.lib.qlu.edu.cn/api.php/space_time_buckets?area=19&day=2022-'+ month+'-'+ day
	url := fmt.Sprintf("http://yuyue.lib.qlu.edu.cn/api.php/space_time_buckets?area=19&day=%s", time.Now().Format("2006-01-02"))
	res, _ := http.NewRequest("GET", url, nil)
	setHeader(res)
	response, err := l.C.Do(res)
	if err != nil {
		return 0, err
	}
	readAll, _ := ioutil.ReadAll(response.Body)
	var resp model.BookTimeIdresp
	err = json.Unmarshal(readAll, &resp)
	if err != nil {
		return 0, err
	}
	return resp.Data.List[0].BookTimeID, nil

}

func (l *Lib) Book(userid, id, advanceTime int) (model.Bookresp, error) {
	//url_book = ''
	url := fmt.Sprintf("http://yuyue.lib.qlu.edu.cn/api.php/spaces/%d/book", id)
	body := fmt.Sprintf("access_token=%s&userid=%d&type=1&id=%d&segment=%d", l.token, userid, id, l.BookTimeId+advanceTime)
	Mlog.PF(logger.LINFO, "开始预约：%d", id)
	bodyreader := strings.NewReader(body)
	request, _ := http.NewRequest("POST", url, bodyreader)
	setHeader(request)
	resp, err := l.C.Do(request)
	var Bookresp model.Bookresp
	if err != nil {
		return Bookresp, err
	}

	readAll, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(readAll, &Bookresp)
	if err != nil {
		return Bookresp, err
	}
	//fmt.Printf("login:%s,%s\n", username, passwd)
	Mlog.PF(logger.LINFO, "预约：%d->%s", id, Bookresp.Msg)
	return Bookresp, nil
}
func setHeader(request *http.Request) {
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	request.Header.Add("Accept", "application/json, text/plain, */*")
	request.Header.Add("Accept-Language", "zh-CN,zh-Hans;q=0.9")
	request.Header.Add("Host", "yuyue.lib.qlu.edu.cn")
	request.Header.Add("Origin", "http://www.skalibrary.com")
	request.Header.Add("Referer", "http://www.skalibrary.com")
	request.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.27(0x18001b33) NetType/WIFI Language/zh_CN")
}
