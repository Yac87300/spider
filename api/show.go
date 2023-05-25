package api

import (
	"bytes"
	"example.com/mod/loadconf"
	"example.com/mod/master"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)


func Signrouter(server *gin.Engine){
	ruleGroup := server.Group("/rule",Check)
	{
		ruleGroup.GET("/get",GetAllrule)
		ruleGroup.GET("/getrs",GetAllruleAndresults)
		ruleGroup.POST("/push",PushFromapi)
		ruleGroup.POST("/GetOpenfuctionFile",GetOpenfuctionFile)
	}
	server.POST("/rule/legproxy",legproxy)
	server.GET("/ping",func(c *gin.Context){
		c.JSON(200,gin.H{
			   "msg" : "pong",
		})
	})

}


//func FunctionDo(c *gin.Context){
//	 Runmode := c.PostForm("Howtodo")
//
//}

func GetOpenfuctionFile(c *gin.Context) {

	if strings.Index(loadconf.Conf["Matser"],c.ClientIP()) == -1   {
		c.JSON(200,gin.H{
			"msg" : "Agent: 非法调用",
		})
		return
	}
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer dst.Close()

	// 复制上传的文件到目标文件
	if _, err = io.Copy(dst, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "File uploaded successfully!",
	})
}


func legproxy(c *gin.Context){
	if strings.Index(loadconf.Conf["Matser"],c.ClientIP()) == -1   {
		 c.JSON(200,gin.H{
			 "msg" : "Agent: 非法调用",
		 })
		 return
	}
     runmode := c.Query("runmode")
     if runmode == "" {
     	c.JSON(200,gin.H{
     		"msg" : "Agent: failed",
		})
		 return
	 }
	 end :=  master.Cmdrun(runmode)
	 c.JSON(200,gin.H{
		"msg" : end,
	 })
}

func Check(c *gin.Context)  {
	key := c.PostForm("key")
	if key != loadconf.Conf["Passwd"] {
		fmt.Println(key,loadconf.Conf["Passwd"])
		c.JSON(503,gin.H{
			"SpiderAgentCheck" : "Passwd not right",
		})
		c.Abort()
		return
	}
	c.Next()
}

func GetAllrule(c *gin.Context){

	list := loadconf.ShareConfload
	cjs := loadconf.Rulejson{}
	list.AlertMethod = cjs.AlertMethod
	c.JSON(200,loadconf.ShareConfload)
}


func Synctasksstatus(taskid,taskname,taskresult,Taskstatus,Datafrom,Tasktime,Alertmethod,AlertData,Alerttype string){
	url := "http://" + loadconf.Conf["Matser"] + "/master/proxy/task"
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("taskid", taskid)
	_ = writer.WriteField("taskname", taskname)
	_ = writer.WriteField("taskresult",taskresult)
	_ = writer.WriteField("Taskstatus", Taskstatus)
	_ = writer.WriteField("Datafrom", Datafrom)
	_ = writer.WriteField("Tasktime", Tasktime)
	_ = writer.WriteField("Alertmethod", Alertmethod)
	_ = writer.WriteField("AlertData", AlertData)
	_ = writer.WriteField("Alerttype", Alerttype)
	_ = writer.WriteField("Aletstatus", "Running")
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

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func GetAllruleAndresults(c *gin.Context){

	rjs := loadconf.ShareConfload
	cjs := loadconf.Rulejson{}
	rjs.AlertMethod = cjs.AlertMethod
	for _,v :=  range rjs.Rulesource{
		result,_ := loadconf.Resultv2.Load(v.Name)
		split := []string{}
		if result == nil || result == ""{
			result = ""
			split = strings.Split("null√null√unkonw","√")
		}else {
			split = strings.Split(result.(string),"√")
		}
		l,_ := loadconf.Md5com.Load(v.Name+loadconf.Conf["Tag"])
		cjs.Rulesource = append(cjs.Rulesource,loadconf.Data{
			Name: v.Name,
			Rs: split[0],
			Status: split[2],
			Time: split[1],
			DataType: v.DataType,
			From: v.From,
			Alert:v.Alert,
			Alertdata: v.Alertdata,
			AlertTo: v.AlertTo,
			ForTime: v.ForTime,
			Md5c: l.(string),
		})
	}
	rjs.Rulesource = cjs.Rulesource
	c.JSON(200,rjs)
}

func PushFromapi(c *gin.Context){
	name := c.PostForm("name")
	data := c.PostForm("data")

	c.JSON(200,
	checkPushdata(name,data))
}

