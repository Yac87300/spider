package api

import (
	"example.com/mod/loadconf"
	"example.com/mod/master"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)


func Signrouter(server *gin.Engine){
	ruleGroup := server.Group("/rule",Check)
	{
		ruleGroup.GET("/get",GetAllrule)
		ruleGroup.GET("/getrs",GetAllruleAndresults)
		ruleGroup.POST("/push",PushFromapi)
	}
	server.POST("/rule/legproxy",legproxy)
	server.GET("/ping",func(c *gin.Context){
		c.JSON(200,gin.H{
			   "msg" : "pong",
		})
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

