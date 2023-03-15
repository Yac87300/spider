package loadconf

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)


var ShareConfload Rulejson
var Legconf Legst
var Groups  Groupjson

type Rulejson struct {
	Tag                     string `json:"Tag"`
	AllowsPushdataFromlocal string `json:"AllowsPushdataFromlocal"`
	Matser                  string `json:"Matser"`
	Port                    string `json:"Port"`
	VGRpc                   string `json:"V-gRpc"`
	Passwd                  string `json:"Passwd"`
	Bind                    string `json:"Bind"`
	CommandLimit            string `json:"CommandLimit"`
	AgentPort               string `json:"AgentPort"`
	AgentCheckAlert         string `json:"agentcheckalert"`

	AlertMethod             []struct {
		                    Path    string `json:"Path"`
		                    RunMode string `json:"RunMode"`
		                    AlertForward string `json:"alertforward"`
	                        } `json:"AlertMethod"`

	Rulesource              []Data `json:"rulesource"`
	Role                    string `json:"role"`

	GatewayProxy            []struct{
		                    OutPath string `json:"outpath"`
		                    InsidePath string `json:"insidePath"`
		                    InsidePathPort string `json:"insidePathPort"`
		                    Servers string `json:"servers"`
	                      	Lvs string `json:"lvs"`
                     		Timecheck string `json:"timecheck"`
		                    Tokencheck string `json:"tokencheck"`
		                    MaxFailedFromOneIP int `json:"max_failed_from_one_ip"`
		                    H  string `json:"h"`
 	}
}

type Groupjson struct {
	Groups []GroupAdd `json:"groups"`
	Distribution []GroupipAdd `json:"distribution"`
}

type GroupAdd struct {
	Name string `json:"name"`
}
type GroupipAdd struct {
	Dt string `json:"dt"`
}

type Data struct {
	Name      string `json:"Name"`
	DataType  string `json:"DataType"`
	From      string `json:"From"`
	Alert     string `json:"Alert"`
	Alertdata string `json:"Alertdata"`
	AlertTo   string `json:"AlertTo"`
	ForTime   int    `json:"ForTime"`
	Status string `json:"status"`
	Time string `json:"time"`
	Rs string `json:"rs"`
	Msg                     string `json:"Msg"`
	Enable string `json:"enable"`
	Md5c   string `json:"md5c"`
}

type Legst struct {
	Legs []struct {
		Legname      string `json:"Legname"`
		Method       string `json:"Method"`
		VarsFromApi  string `json:"VarsFromApi"`
		HandleMode   string `json:"HandleMode"`
		ReturnResult string `json:"ReturnResult"`
		Token        string `json:"Token"`
		Proxyto      string `json:"proxyto"`
	} `json:"legs"`
}

//低频map
var Rule map[string]string
var Conf map[string]string

//测试版本 废弃⚠️
var Result map[string]string

//正式版本
var Resultv2 sync.Map
var Rulemsg sync.Map
var Md5com sync.Map
var GroupCache sync.Map
var GroupipCache sync.Map


func loadgroup(){
	f2, err := ioutil.ReadFile("cache/group.json")
	if err != nil {
		fmt.Println("read fail", err)
		os.Exit(-1)
	}
	Groupsjson := Groupjson{}
	err = json.Unmarshal([]byte(f2),&Groupsjson)
	if err != nil {
		fmt.Println("Load leg.json error : " ,err)
		os.Exit(-1)
	}
	Groups = Groupsjson

	for _,v := range Groups.Groups{
		GroupCache.Store(v.Name,v.Name)
	}
	for _,v := range Groups.Distribution{
		data := strings.Split(v.Dt,":")
		groupname := data[0]
		ip := data[1]
		GroupipCache.Store(ip,groupname)
	}
}

func load2(){
	f2, err := ioutil.ReadFile("cache/leg.json")
	if err != nil {
		fmt.Println("read fail", err)
		os.Exit(-1)
	}
	Legjson := Legst{}
	err = json.Unmarshal([]byte(f2),&Legjson)
	if err != nil {
		fmt.Println("Load leg.json error : " ,err)
		os.Exit(-1)
	}
	Legconf = Legjson

}


func LoadjsonFromLocal() {
	Rule = make(map[string]string)
	Conf =make(map[string]string)

	f, err := ioutil.ReadFile("rule.json")
	if err != nil {
		fmt.Println("read fail", err)
		os.Exit(-1)
	}

	Rulejsons := Rulejson{}

	err = json.Unmarshal([]byte(f),&Rulejsons)
	if err != nil {
		fmt.Println("Load rule.json error : " ,err)
		os.Exit(-1)
	}

	if Rulejsons.Role == "master" {
		load2()
		loadgroup()
	}


	Conf["Tag"] = Rulejsons.Tag
	Conf["AllowsPushdataFromlocal"] = Rulejsons.AllowsPushdataFromlocal
	Conf["Matser"] = Rulejsons.Matser
	Conf["Port"]= Rulejsons.Port
	Conf["CommandLimit"] =Rulejsons.CommandLimit
	Conf["Passwd"]=Rulejsons.Passwd

	Rulejsons.Passwd = "********"
	Rulejsons.Port = "-1"
	Rulejsons.Matser = "******"

	ShareConfload = Rulejsons

	if Rulejsons.Matser == "master" {
		return
	}

	for _,v := range Rulejsons.Rulesource{
		time.Sleep(time.Millisecond * 300)

		c := Rule[v.Name]
		if  c != ""{
			fmt.Println(v.Name,"Name repeat,can't not load this rule")
		}

		if v.Enable == "stop" {
			continue
		}

		if v.ForTime < 15 {
			fmt.Println(v.Name,"ForTime lessThan 15,maybe this server will be highPress")
			os.Exit(0)
		}

		Name := v.Name
		DataType := v.DataType
		Form := v.From
		Alert := v.Alert
		Alertdata := v.Alertdata
		AlertTo := v.AlertTo
		ForTime := v.ForTime
		Rule[Name] = Form + "√" + fmt.Sprint(ForTime) + "√" + Alert + "√" + DataType + "√" + Alertdata + "√" + AlertTo
		Rulemsg.Store(Name,v.Msg)
	}
}

func calculateMD5(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

func MakeStart(){
	for k,v :=  range Rule{
    ydata := strings.Split(v,"√")
    cmdrun := ydata[0]
    fortime := ydata[1]
		alert := ydata[2]
		datatype := ydata[3]
		alertdata := ydata[4]
		alertto := ydata[5]
		Md5com.Store(calculateMD5(k+Conf["Tag"]),k+"√"+Conf["Tag"])
		Md5com.Store(k+Conf["Tag"],calculateMD5(k+Conf["Tag"]))
		if cmdrun == "apiPush" || cmdrun == "localPush"{
		Addresultv2(k,"null","unkonw")
		go TimeWatchMan(k,fortime,alertto)
    	continue
	}
    go Run(k,cmdrun,fortime,alert,datatype,alertdata,alertto)
    time.Sleep(time.Millisecond * 500)
	}
}


func TimeWatchMan(name string,alertTime,alerto string){
	alertTimes,_ := strconv.Atoi(alertTime)
	 for {
		 time.Sleep(time.Second * time.Duration(alertTimes))
	 	t,_ := Resultv2.Load(name+":Time")
	 	if t == nil {
	 		t = ""
		}
	 	ok := panduantime(t.(string),alertTimes)
	 	if !ok{
			Addresultv2(name,"Timeout,LastPush:" + t.(string),"bad")
				msgtemplate,_ := Rulemsg.Load(name)
				SendAlert(Conf["Matser"],ShareConfload.Tag,name,msgtemplate.(string),alerto)
		}
		time.Sleep(time.Millisecond * time.Duration(200))

	 }
}

func panduantime(times string,alerttime int)bool{
	t1 := time.Now()
	stringTime := times
	loc, _ := time.LoadLocation("Local")
	the_time, err := time.ParseInLocation("2006-01-02 15:04:05", stringTime, loc)
	if err != nil {
		return false
	}
	if t1.Sub(the_time) > time.Duration(time.Second * time.Duration(alerttime)) {
		return false
	}
	return true
}

func Run(name,cmdrun,fortime,alert,dataType,alerdata,alerto string){
	if Conf["CommandLimit"] == "yes" {
		if strings.Index(cmdrun,"rm") != -1 || strings.Index(cmdrun,"dd") != -1 || strings.Index(cmdrun,"reboot") != -1 || strings.Index(cmdrun,"init") != -1{
		 fmt.Println(cmdrun,"trigger Conf.CommandLimit")
			os.Exit(-1)
		}
	}
	Resultv2.Store(name,"")
	 sleeptime,_ := strconv.Atoi(fortime)

	 fmt.Println(name)

	 for {
		 cmd := exec.Command("/bin/bash", "-c", cmdrun)
		 //创建获取命令输出管道
		 stdout, err := cmd.StdoutPipe()
		 cmd.Stderr = cmd.Stdout
		 if err != nil {
			 fmt.Printf("Error:can not obtain stdout pipe for command:%s\n", err)
			 return
		 }
		 //执行命令
		 if err := cmd.Start(); err != nil {
			 fmt.Println("Error:The command is err,", err)
			 return
		 }
		 //读取所有输出
		 bytes, err := ioutil.ReadAll(stdout)
		 if err != nil {
			 fmt.Println("ReadAll Stdout:", err.Error())
			 return
		 }

		 //fmt.Println(string(bytes))

		 cmdrunReturndata := string(bytes)

		 if len(cmdrunReturndata) == 0{
		 	cmdrunReturndata = "无返回值"
		 }

		 status := "ok"

		 switch alert {
		 case ">":
		 	cmdrunReturndata = strings.ReplaceAll(strings.ReplaceAll(string(bytes),"\n","")," ","")
		 	rundata := cmdrunReturndata
		 	comparedata,err :=  strconv.Atoi(alerdata)
			if err != nil {
				cmdrunReturndata= "数值判断器运行错误(" + string(bytes) +")"

			}
		 	rundataNum,err := strconv.Atoi(rundata)
		 	if err != nil {
				cmdrunReturndata= "数值判断器运行错误(" + alerdata +")"
			}
			if rundataNum > comparedata {
				cmdrunReturndata = cmdrunReturndata +  "～"  + " Cmddata:"  + rundata +  " > " + alerdata
				status = "bad"

			}
		 case "<":
			 cmdrunReturndata = strings.ReplaceAll(strings.ReplaceAll(string(bytes),"\n","")," ","")
			rundata := cmdrunReturndata
			comparedata,err :=  strconv.Atoi(alerdata)
			if err != nil {

				cmdrunReturndata= "Alertdata数值判断器运行错误(" + string(alerdata) +")"
			}
			rundataNum,err := strconv.Atoi(rundata)
			if err != nil {
				cmdrunReturndata= "Rundata数值判断器运行错误(" + rundata + cmdrunReturndata +")"
			}
			if rundataNum < comparedata {
				cmdrunReturndata = cmdrunReturndata +  "～"  + " Cmddata:"  + rundata +  " < " + alerdata
				status = "bad"
			}
		 case "=":
			rundata := cmdrunReturndata
			if rundata == alerdata {
				cmdrunReturndata = cmdrunReturndata +  "～"  + " Cmddata:"  + rundata +  " = " + alerdata
				status = "bad"
			}
		 case "!=":
			rundata := cmdrunReturndata
			if rundata != alerdata {
				cmdrunReturndata = cmdrunReturndata +  "～"  + " Cmddata:"  + rundata +  " != " + alerdata
				status = "bad"
			}
		 case "index":

			 if strings.Index(alerdata,"|") != -1 {
				 lip := strings.Split(alerdata,"|")
				 msg := ""
				 for i:=0;i<cap(lip);i++{
					 if strings.Index(string(bytes),lip[i]) != -1 {
						 msg =  msg + "～"  + " Cmddata:"  + cmdrunReturndata +  " index " + lip[i]
						 status = "bad"
					 }
				 }
				 cmdrunReturndata = msg

			 }else {
				 if strings.Index(string(bytes),alerdata) != -1 {
					 cmdrunReturndata =  cmdrunReturndata +  "～"  + " Cmddata:"  + cmdrunReturndata +  " index " + alerdata
					 status = "bad"
				 }
			 }

		 case "!index":

			 if strings.Index(alerdata,"|") != -1 {
				 lip := strings.Split(alerdata,"|")
				 msg := ""
				 for i:=0;i<cap(lip);i++{
					 if strings.Index(string(bytes),lip[i]) == -1 {
						 msg =  msg + "～"  + " Cmddata:"  + cmdrunReturndata +  " !index " + lip[i]
						 status = "bad"
					 }
				 }
				 cmdrunReturndata = msg
			 }else {
				 if strings.Index(string(bytes),alerdata) == -1 {
					 cmdrunReturndata = cmdrunReturndata +  "～"  + " Cmddata:"  + cmdrunReturndata +  " !index " + alerdata
					 status = "bad"
				 }
			 }
		 }
		 Addresultv2(name,cmdrunReturndata,status)

		 if status != "ok" {
			 msgtemplate,_ := Rulemsg.Load(name)
			 SendAlert(Conf["Matser"],ShareConfload.Tag,name,msgtemplate.(string),alerto)
		 }
		 time.Sleep(time.Second *  time.Duration(sleeptime))
	 }
	//return string(bytes), err
}


func Addresultv2(name,rs,status string){
	 Resultv2.Store(name,rs+"√"+time.Now().Format("2006-01-02 15:04:05")+"√"+status)
}


func ContoMaster()bool{
	url := "http://" + Conf["Matser"] + "/master/add"
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	sendip := Conf["Tag"]
	if sendip == "" {
		sendip = "null"
	}

	_ = writer.WriteField("ip",sendip)
	err := writer.Close()
	if err != nil {
		return false
	}


	client := &http.Client {}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false
	}

	mj := Mastejson{}

	err = json.Unmarshal(body,&mj)
	if err != nil {
		fmt.Println("Establishing communication with master Failed !")
		return false
	}
	if mj.Master == "ok"{
		fmt.Println("Establishing communication with master SSSSuccess !")
		Conf["Tag"] = mj.Else
		ShareConfload.Tag = mj.Else
		fmt.Println(mj.Else,"jion to Master")
		return true
	}

	if mj.Master == "ip already bein"{
		fmt.Println("Establishing communication with master SSSSuccess !")
		Conf["Tag"] = mj.Else
		ShareConfload.Tag = mj.Else
		fmt.Println(mj.Else,"jion to Master")
		return true
	}

	return false


}

type Mastejson struct {
	Master string `json:"master"`
	Else string `json:"else"`
}

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






