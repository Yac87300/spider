package script

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func Sendmsg(key,msg string){
	Hookaddress := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=" + key
	content := `{"msgtype": "text",
      "text": {"content": "` + msg + ` "}
               }`

	content2 := strings.NewReader(content)
	req, err := http.NewRequest("POST", Hookaddress, content2)
	if err != nil {
		fmt.Sprint(1,err)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	req.Header.Set("Content-Type", "application/json; charset=uft-8")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Sprint(2,err)
	}
	defer resp.Body.Close()
}
