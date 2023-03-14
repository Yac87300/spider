package api

import (
	"example.com/mod/loadconf"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

func checkPushdata(name,data string)interface{}{

	if loadconf.Rule[name] == ""{
		return gin.H{
			"apiPush" : "Can't not find Rule: " + name,
		}
	}

	    v :=  loadconf.Rule[name]

		ydata := strings.Split(v,"√")
		cmdrun := ydata[0]

		if cmdrun != "apiPush"{
			return gin.H{
				"apipush" : "不合法 0",
			}
		}

		//fortime := ydata[1]
		alert := ydata[2]
		//datatype := ydata[3]
		alerdata := ydata[4]
		alertto := ydata[5]
		cmdrunReturndata := string(data)

	status :="ok"

		switch alert {
		case ">":
			cmdrunReturndata = strings.ReplaceAll(strings.ReplaceAll(string(data),"\n","")," ","")
			rundata := cmdrunReturndata
			comparedata,err :=  strconv.Atoi(alerdata)
			if err != nil {
				return gin.H{
					"apipush" : "不合法 1",
				}
			}
			rundataNum,err := strconv.Atoi(rundata)
			if err != nil {
				return gin.H{
					"apipush" : "不合法 2",
				}
			}
			if rundataNum > comparedata {
				cmdrunReturndata = cmdrunReturndata +  "～"  + " Cmddata:"  + rundata +  " > " + alerdata
				status ="bad"
			}
		case "<":
			cmdrunReturndata = strings.ReplaceAll(strings.ReplaceAll(string(data),"\n","")," ","")
			rundata := cmdrunReturndata
			comparedata,err :=  strconv.Atoi(alerdata)
			if err != nil {
				return gin.H{
					"apipush" : "不合法 3",
				}
			}
			rundataNum,err := strconv.Atoi(rundata)
			if err != nil {
				return gin.H{
					"apipush" : "不合法 4",
				}
			}
			if rundataNum < comparedata {
				status ="bad"
				cmdrunReturndata = cmdrunReturndata +  "～"  + " Cmddata:"  + rundata +  " < " + alerdata
			}
		case "=":
			rundata := cmdrunReturndata
			if rundata == alerdata {
				status ="bad"
				cmdrunReturndata = cmdrunReturndata +  "～"  + " Cmddata:"  + rundata +  " = " + alerdata
			}
		case "!=":
			rundata := cmdrunReturndata
			if rundata != alerdata {
				status ="bad"
				cmdrunReturndata = cmdrunReturndata +  "～"  + " Cmddata:"  + rundata +  " != " + alerdata
			}
		case "index":
			if strings.Index(alerdata,"|") != -1 {
				lip := strings.Split(alerdata,"|")
				msg := ""
				for i:=0;i<cap(lip);i++{
					if strings.Index(string(data),lip[i]) != -1 {
						msg =  msg + "～"  + " Cmddata:"  + cmdrunReturndata +  " index(检索到) " + lip[i]
						status = "bad"
					}
				}
				cmdrunReturndata = msg
			}else {
				if strings.Index(string(data), alerdata) != -1 {
					cmdrunReturndata = cmdrunReturndata + "～" + " Cmddata:" + cmdrunReturndata + " index(检索到) " + alerdata
					status = "bad"
				}
			}
		case "!index":
			if strings.Index(alerdata,"|") != -1 {
				lip := strings.Split(alerdata,"|")
				msg := ""
				for i:=0;i<cap(lip);i++{
					if strings.Index(string(data),lip[i]) == -1 {
						msg =  msg + "～"  + " Cmddata:"  + cmdrunReturndata +  " !index(没检索到) " + lip[i]
						status = "bad"
					}
				}
				cmdrunReturndata = msg
			}else {
				if strings.Index(string(data),alerdata) == -1 {
					cmdrunReturndata = cmdrunReturndata +  "～"  + " Cmddata:"  + cmdrunReturndata +  " !index(没检索到) " + alerdata
					status = "bad"
				}
			}
		}



	    loadconf.Addresultv2(name,cmdrunReturndata,status)
	    loadconf.Resultv2.Store(name+":Time",time.Now().Format("2006-01-02 15:04:05"))

	if status != "ok" {
		msgtemplate,_ := loadconf. Rulemsg.Load(name)
		loadconf.SendAlert(loadconf.Conf["Matser"],loadconf. ShareConfload.Tag,name,msgtemplate.(string),alertto)
	}

	return gin.H{
		   "apiPush" : "ok",
	}
}
