package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	mail "github.com/sunreaver/goTools/mail"
	sys "github.com/sunreaver/goTools/system"
	"github.com/sunreaver/mahonia"
)

var (
	// StockDataRegexp sina接口返回数据提取
	StockDataRegexp = regexp.MustCompile(`="(.*)";`)
)

type Config struct {
	Mail   []string `json:"emails"`
	Stocks []string `json:"stocks"`
}

func main() {
	configs, e := readFile(sys.CurPath() + sys.SystemSep() + "stock.json")
	if e != nil {
		fmt.Println(e.Error())
		return
	}

	format := "%s:\t>>>>>昨收:%s\t今收:%s\t今开:%s\t幅度:%s<<<<<\r\n"

	for _, cfg := range configs {

		if len(cfg.Mail) == 0 {
			continue
		}
		outStr := ""
		for _, v := range cfg.Stocks {
			resp, err := http.Get("http://hq.sinajs.cn/list=" + v)
			if err != nil {
				continue
			}
			defer resp.Body.Close()
			body, e := ioutil.ReadAll(resp.Body)
			if e != nil {
				continue
			}

			str := string(body)
			enc := mahonia.NewDecoder("gbk")
			matchs := StockDataRegexp.FindAllStringSubmatch(enc.ConvertString(str), -1)
			if len(matchs) == 0 || len(matchs[0]) < 2 {
				continue
			}

			numerical := strings.Split(matchs[0][1], ",")
			if len(numerical) < 5 {
				continue
			}

			upDown := "未知"
			yestoday, e1 := strconv.ParseFloat(numerical[2], 64)
			today, e2 := strconv.ParseFloat(numerical[3], 64)
			if e1 == nil && e2 == nil {
				upDown = strconv.FormatFloat(today-yestoday, 'f', -1, 64)[0:6]
			}
			outStr = outStr + fmt.Sprintf(format, numerical[0], numerical[2], numerical[3], numerical[1], upDown)
		}

		outStr = outStr + "\r\nHappy day!\r\n"

		e = mail.SendMail(sys.CurPath()+sys.SystemSep()+"auth.json", outStr, cfg.Mail)
		if e != nil {
			fmt.Println(e.Error())
		}
	}
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
}

func readFile(fileName string) ([]Config, error) {
	b, e := ioutil.ReadFile(fileName)
	if e != nil {
		return nil, e
	}

	var cfg []Config
	e = json.Unmarshal(b, &cfg)
	if e != nil {
		return nil, e
	}
	return cfg, e
}
