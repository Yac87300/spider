package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

func SendAlert(m,ip,name,msg,methodName string){
	url := "http://"+ m +"/master/alert"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("ip", ip)
	_ = writer.WriteField("name", name)
	_ = writer.WriteField("msg", msg)
	_ = writer.WriteField("methodName", methodName)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return
	}


	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
