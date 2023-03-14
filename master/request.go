package master

import (
	"bytes"
	"encoding/json"
	"example.com/mod/loadconf"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

func Getrule(ip,passwd,path string)(loadconf.Rulejson,bool){
	ar := loadconf.Rulejson{}
	url :=  "http://" +ip  + path
	method := "GET"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("key", passwd)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return ar,false
	}


	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return ar,false
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ar,false
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ar,false
	}
	 _ = json.Unmarshal(body,&ar)
	return ar,true
}

func VenusPost(c2 *gin.Context,Allvars ...interface{})interface{}{
	AllvarsLocalStorage := Allvars
	Vars := []interface{}{}
	for _,v := range AllvarsLocalStorage{
		varname := v.(string) + "VARS"
		varname = c2.PostForm(v.(string))
		if varname == "nilnulls"{
			varname = ""
		}else if varname == ""{
			c2.JSON(503, gin.H{
				"error": "定义了变量:(" + fmt.Sprint(v.(string)) + "),请求中未找到此变量",
			})
			break
		}
		Vars = append(Vars,varname)
	}
	return Vars
}

func Tranceforinterface(x interface{})(v []string){
	for _,va := range x.([]interface{}){
		v = append(v,va.(string))
	}
	return v
}

func Ginvars(x interface{})(v []string){
	for _,va := range x.([]interface{}){
		v = append(v,va.(string))
	}
	return v
}



