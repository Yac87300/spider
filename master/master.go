package master

import (
	"bufio"
	"bytes"
	"encoding/json"
	"example.com/mod/loadconf"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)


var blacklist sync.Map
var iplist sync.Map
var Offline sync.Map
var proxycomp sync.Map
var spmids sync.Map


func Bindcheck(c *gin.Context){
	spid := c.Query("spid")
	data,err := Jiemi("19a7251c679b22ccf83bd8a9709910be",spid)
	fmt.Println(data)
	if err != nil{
		c.JSON(503,gin.H{
			"error" : "Your cliTools no spid/spid notright to make token,stop your cliTools,Please call the admin to Add.",
		})
		c.Abort()
		return
	}

	lip := strings.Split(data,":")
	if c.ClientIP() != lip[0] {
		fmt.Println(c.ClientIP(),lip[0])
		c.JSON(503,gin.H{
			"error" : "Your spid not right.",
		})
		c.Abort()
		return
	}

	k,ok := spmids.Load(data)
	fmt.Println("[cliTools]:",k,data,"From:",c.ClientIP(),"To:",c.FullPath())
	if ok {
		c.Next()
	}else {
		c.JSON(503,gin.H{
			"error" : "Your cliTools no spid/spid notright to make token,stop your cliTools,Please call the admin to Add.",
		})
		c.Abort()
		return
	}
}

func testjiami(c *gin.Context){
	data,err := Jiami("19a7251c679b22ccf83bd8a9709910be",c.PostForm("id"))
	if err != nil{
		c.JSON(200,gin.H{
			"error": err,
		})
		return
	}
	c.JSON(200,gin.H{
		"msg": data,
	})
}

func Signrouter(server *gin.Engine){

	server.POST("/master/sm4",testjiami)
	server.POST("/master/add",Addip)
	server.POST("/master/alert",RunAlertMethod)
	masterGroup := server.Group("/master",Bindcheck)
	{
		masterGroup.POST("/ping", func(c *gin.Context) {
			msg := c.PostForm("msg")
			token := c.PostForm("token")
			q := c.Query("test")
			if msg == ""{
				c.JSON(200,gin.H{
					"msg" : "null",
				})
				return
			}
			c.JSON(200,gin.H{
			 	"msg" : msg + token + q,
			 })
		})

		masterGroup.POST("/show",Showiplist)
		masterGroup.POST("/getrule",GetallruleFromAgent)
		masterGroup.POST("/getrulers",GetallruleAndRsFromAgent)

		masterGroup.POST("/delayAlert",SetBlackAlert)
		masterGroup.POST("/delatAlertwithMd5",RunAlertMethodWithMd5)
		masterGroup.POST("/openalertWithMd5",OpendelayAlertMethodWithMd5)
		masterGroup.POST("/startAlert",StartBlackAlert)
		masterGroup.POST("/showdelayAlert",Showdefer)
		masterGroup.POST("/proxy/:proxyname",proxy)
		masterGroup.GET("/proxy/:proxyname",proxy)
		masterGroup.POST("/showmethod",Showmethod)
		masterGroup.POST("/syncstatus",Sync)
		masterGroup.POST("/leg/:legname",Leg)
		masterGroup.POST("/group",Managegroup)
		masterGroup.POST("/groupip",ManageGroupip)
	}

}




func Managegroup(c *gin.Context) {
	 if c.PostForm("manage") == "add" {
        name := c.PostForm("name")
	 	loadconf.Groups.Groups = append(loadconf.Groups.Groups,loadconf.GroupAdd{
	 		Name: name,
		})

	 	ysdata,err := json.Marshal(loadconf.Groups)
	 	if err !=nil{
			c.JSON(200,gin.H{
				"error" : "add failed,try later",
			})
			return
		}
	 	err =Writefile("cache/group.json",string(ysdata))
        if err != nil{
        	c.JSON(200,gin.H{
        		"error" : "add failed,try later",
			})
			return
		}
		loadconf.GroupCache.Store(name,name)
		 c.JSON(200,gin.H{
			 "msg" : "ok",
		 })
	 }else if c.PostForm("manage") == "delete" {
		 var newgroups  []loadconf.GroupAdd

		 dname := c.PostForm("name")
	 	for _,v := range loadconf.Groups.Groups{
	 		if v.Name == dname {
               continue
			}
			newgroups = append(newgroups,loadconf.GroupAdd{Name: v.Name})
		}
		 loadconf.Groups.Groups = newgroups

		 var newgroupsip  []loadconf.GroupipAdd
		 for _,v := range loadconf.Groups.Distribution{
			lip := strings.Split(v.Dt,":")
			if lip[0] == dname {
				continue
			}
			newgroupsip = append( newgroupsip,loadconf.GroupipAdd{Dt: v.Dt})
		 }
		 loadconf.Groups.Distribution = newgroupsip

		ysdata,err := json.Marshal(loadconf.Groups)
		 if err !=nil{
			 c.JSON(200,gin.H{
				 "error" : "add failed,try later",
			 })
			 return
		 }
		 err =Writefile("cache/group.json",string(ysdata))
		 if err != nil{
			 c.JSON(200,gin.H{
				 "error" : "add failed,try later",
			 })
			 return
		 }
		 c.JSON(200,gin.H{
			 "msg" : "ok",
		 })
	 }else if c.PostForm("manage") == "show" {
		 c.JSON(200, loadconf.Groups.Groups)
	 }
}

func ManageGroupip(c *gin.Context){
	 if c.PostForm("manage") == "show"{
	 	c.JSON(200,loadconf.Groups.Distribution)
	 }else if c.PostForm("manage")== "add" {

	 	group := c.PostForm("group")
	 	_,ok := loadconf.GroupCache.Load(group)
	 	if !ok {
			c.JSON(200,gin.H{
				"error" : group + " Not Found",
			})
			return
		}


	 	ip := c.PostForm("ip")
		 _,ok = iplist.Load(ip)
		 if !ok {
			 c.JSON(200,gin.H{
				 "error" : ip + " Not Found",
			 })
			 return
		 }

	 	loadconf.Groups.Distribution = append(loadconf.Groups.Distribution,loadconf.GroupipAdd{
			 Dt: group+":"+ip,
	 	})

	 	loadconf.GroupipCache.Store(ip,group)

	 	ysdata,err := json.Marshal(loadconf.Groups)
	 	if err !=nil{
			 c.JSON(200,gin.H{
				 "error" : "add failed,try later",
			 })
			 return
		 }
		 err =Writefile("cache/group.json",string(ysdata))
		 if err != nil{
			 c.JSON(200,gin.H{
				 "error" : "add failed,try later",
			 })
			 return
		 }
		 c.JSON(200,gin.H{
			 "msg" : "ok",
		 })
	 }else if c.PostForm("manage")== "delete" {

		 group := c.PostForm("group")
		 _,ok := loadconf.GroupCache.Load(group)
		 if !ok {
			 c.JSON(200,gin.H{
				 "error" : group + " Not Found",
			 })
			 return
		 }


		 ip := c.PostForm("ip")
		 _,ok = iplist.Load(ip)
		 if !ok {
			 c.JSON(200,gin.H{
				 "error" : ip + " Not Found",
			 })
			 return
		 }

		 var newgroupsip  []loadconf.GroupipAdd

		 for _,v := range loadconf.Groups.Distribution{
			 lip := strings.Split(v.Dt,":")
			 if lip[1] == ip && lip[0] == group{
				 continue
			 }
			 newgroupsip = append( newgroupsip,loadconf.GroupipAdd{Dt: v.Dt})
		 }
		 loadconf.Groups.Distribution = newgroupsip

		 ysdata,err := json.Marshal(loadconf.Groups)
		 if err !=nil{
			 c.JSON(200,gin.H{
				 "error" : "add failed,try later",
			 })
			 return
		 }
		 err =Writefile("cache/group.json",string(ysdata))
		 if err != nil{
			 c.JSON(200,gin.H{
				 "error" : "add failed,try later",
			 })
			 return
		 }
		 c.JSON(200,gin.H{
			 "msg" : "ok",
		 })
	 }
}


func Sync(c *gin.Context){
	iplist.Range(func(key, value interface{}) bool {
		if !ping(key.(string)){
			Offline.Store(key.(string),"Offonline")
		}else {
			Offline.Delete(key)
		}
		return true
	})
	c.JSON(200,gin.H{
		"master" : "ok",
	})
}

func Showmethod(c *gin.Context){
	 c.JSON(200,loadconf.ShareConfload.AlertMethod)
}

func Addip(c *gin.Context){

	ip := c.PostForm("ip")
	if ip == "null"{
		ip = c.ClientIP()
	}
	fmt.Println(ip)
	_,ok := iplist.Load(ip)

	if ok{
		c.JSON(200,gin.H{
			"master" :"ip already bein",
			"else":ip,
		})
		return
	}


	v := Readfile("./cache/signfile")
	err := Writefile("./cache/signfile",v+"\n"+ip)
	if err != nil{
		c.JSON(200,gin.H{
			"err": err,
		})
		return
	}
	iplist.Store(ip,"sign")
	c.JSON(200,gin.H{
		"master":"ok",
		"else":ip,
	})
}



func Writefile(filename string,nei string) error{
	fileName := filename
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("file create failed. err: " + err.Error())
	} else {

		content := nei
		n, _ := f.Seek(0, os.SEEK_END)
		_, err = f.WriteAt([]byte(content), n)

		defer f.Close()
	}
	return err
}

func Readfile(filename string)string{
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("read fail", err)
		os.Exit(-1)
	}
	return  string(f)
}

func Cacheip(){
	file, err := os.Open("./cache/signfile")
	if err != nil {
		fmt.Println("rule err:",err)
		os.Exit(-1)
	}
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		if fileScanner.Text() == ""{
			continue
		}
		iplist.Store(fileScanner.Text(),"sign")
	}
}

func CacheProxy(){
	for _,v := range loadconf.ShareConfload.GatewayProxy {
		proxycomp.Store(v.OutPath,"")
		proxycomp.Store(v.OutPath+"outpath",v.OutPath)
		proxycomp.Store(v.OutPath+"InsidePath",v.InsidePath)
		proxycomp.Store(v.OutPath+"InsidePathPort",v.InsidePathPort)
		proxycomp.Store(v.OutPath+"Servers",v.Servers)
		proxycomp.Store(v.OutPath+"H",v.H)
		proxycomp.Store(v.OutPath+"Tokencheck",v.Tokencheck)
	}
}

type Iplist struct {
	Iplists []Data `json:"iplists"`
}

type Data struct {
	Ip string `json:"ip"`
	Status string `json:"status"`
	Group string `json:"group"`
	DelayName string `json:"delay"`
	DelayData string `json:"delaydata"`
}


type Rulelist struct {
	 Data []loadconf.Rulejson `json:"data"`
}

func Showiplist(c *gin.Context){
	ciplist := Iplist{}

	grouptag := c.Query("group")
	if grouptag != "" {
		iplist.Range(func(key, value interface{}) bool {

			_,ok := Offline.Load(key.(string))
			if ok {
				ipk,ok := loadconf.GroupipCache.Load(key.(string))
				if ok {
					if ipk == grouptag {
						ciplist.Iplists = append(ciplist.Iplists,Data{Ip: key.(string),Status:"offonline",Group: ipk.(string)})
					}
				}
			}else {
				ipk,ok := loadconf.GroupipCache.Load(key.(string))
				if ok {
					if ipk == grouptag {
						ciplist.Iplists = append(ciplist.Iplists,Data{Ip: key.(string),Status:"online",Group: ipk.(string)})
					}
				}
			}
			return true
		})
	}else {
		iplist.Range(func(key, value interface{}) bool {

			_,ok := Offline.Load(key.(string))
			if ok {
				ipk,ok := loadconf.GroupipCache.Load(key.(string))
				if ok {
					ciplist.Iplists = append(ciplist.Iplists,Data{Ip: key.(string),Status:"offonline",Group: ipk.(string)})
				}else {
					ciplist.Iplists = append(ciplist.Iplists,Data{Ip: key.(string),Status:"offonline",Group:"NULL"})
				}

			}else {
				ipk,ok := loadconf.GroupipCache.Load(key.(string))
				if ok {
					ciplist.Iplists = append(ciplist.Iplists,Data{Ip: key.(string),Status:"online",Group: ipk.(string)})
				}else {
					ciplist.Iplists = append(ciplist.Iplists,Data{Ip: key.(string),Status:"online",Group:"NULL"})
				}

			}
			return true
		})
	}



	c.JSON(200,ciplist)
}



func GetallruleFromAgent(c *gin.Context){
	crulelist := Rulelist{}
	grouptag := c.Query("group")

	if grouptag != ""{
		iplist.Range(func(key, value interface{}) bool {
			_,ok := Offline.Load(key.(string))
			if ok {

			}else {
				ipk , _ := loadconf.GroupipCache.Load(key.(string))
				if  ipk == grouptag{
					rule,ok := Getrule(key.(string) +":" +loadconf.ShareConfload.AgentPort,loadconf.Conf["Passwd"],"/rule/get")
					if ok {
						crulelist.Data = append(crulelist.Data,rule)
						return true
					}
				}

			}
			return ok
		})
	}else {
		iplist.Range(func(key, value interface{}) bool {
			_,ok := Offline.Load(key.(string))
			if ok {
			}else {
				rule,ok := Getrule(key.(string) +":" +loadconf.ShareConfload.AgentPort,loadconf.Conf["Passwd"],"/rule/get")
				if ok {
					crulelist.Data = append(crulelist.Data,rule)
					return true
				}
			}
			return ok
		})
	}


	c.JSON(200,crulelist)
}


func GetallruleAndRsFromAgent(c *gin.Context){
	crulelist := Rulelist{}
	grouptag := c.Query("group")

	if grouptag != ""{
		iplist.Range(func(key, value interface{}) bool {
			_,ok := Offline.Load(key.(string))
			if ok {

			}else {
				ipk , _ := loadconf.GroupipCache.Load(key.(string))
				if  ipk == grouptag{
					rule,ok := Getrule(key.(string) +":" +loadconf.ShareConfload.AgentPort,loadconf.Conf["Passwd"],"/rule/getrs")
					if ok {
						crulelist.Data = append(crulelist.Data,rule)
						return true
					}
				}

			}
			return ok
		})
	}else {
		iplist.Range(func(key, value interface{}) bool {
			_,ok := Offline.Load(key.(string))
			if ok {
			}else {
				rule,ok := Getrule(key.(string) +":" +loadconf.ShareConfload.AgentPort,loadconf.Conf["Passwd"],"/rule/getrs")
				if ok {
					crulelist.Data = append(crulelist.Data,rule)
					return true
				}
			}
			return ok
		})
	}


	c.JSON(200,crulelist)
}

func CheckAlive(){
        for {
		iplist.Range(func(key, value interface{}) bool {
			if !ping(key.(string)){
				Offline.Store(key.(string),"Offonline")
			}else {
				Offline.Delete(key)
			}
			return true
		})
		time.Sleep(time.Second * 30)
	}
}

func ping(ip string)bool{
	url := "http://" + ip + ":" + loadconf.ShareConfload.AgentPort + "/ping"
	method := "GET"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("key", loadconf.Conf["Passwd"])
	err := writer.Close()
	if err != nil {
		return false
	}


	client := &http.Client {Timeout: time.Second * 2}
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
		fmt.Println(err)
		return false
	}
	agentjson := agentmsg{}
	err = json.Unmarshal(body,&agentjson)
	if err != nil{
		return false
	}

	if agentjson.Msg != "pong" {
		return false
	}
	return true
}

type agentmsg struct {
	Msg string `json:"msg"`
}


func proxy(c *gin.Context){
	proxynames := c.Param("proxyname")
	Pathproxy := ""
	_,ok := proxycomp.Load(proxynames)
	if ok {
	   	H,_ := proxycomp.Load(proxynames+"H")
	   	Server ,_ := proxycomp.Load(proxynames+"Servers")
			InsidePathPort,_:= proxycomp.Load(proxynames+"InsidePathPort")
			InsidePath,_:= proxycomp.Load(proxynames+"InsidePath")
			Pathproxy = H.(string) + "://" +  Server.(string) + ":" + InsidePathPort.(string) + InsidePath.(string)
	}else {
		c.JSON(200,gin.H{
			"msg" : "请求不合法 p1101",
		})
		 return
	}
	reqProxy(c,Pathproxy,proxynames)
}
func reqProxy(c *gin.Context,Pathproxy string,Proxyname string){
	targetURL := Pathproxy
	// 创建HTTP客户端
	client := &http.Client{}
	// 创建一个新的请求对象
	req, err := http.NewRequest(c.Request.Method, targetURL, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	query := req.URL.Query()
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	req.URL.RawQuery = query.Encode()

	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(buf)
	req.Body = ioutil.NopCloser(buf)


	if strings.Index(fmt.Sprint(req.Body),"token") == -1 {
		c.JSON(200,gin.H{
			"msg" : "请求不合法 p1102",
		})
		return
	}

	x := strings.Split(fmt.Sprint(req.Body),"\"token\"")
	v := strings.Split(x[1],"-")
	tokens := strings.ReplaceAll(v[0]," ","")
	usertoken,_ := proxycomp.Load(Proxyname+"Tokencheck")
	if strings.Index(tokens,usertoken.(string)) == -1 {
		c.JSON(200,gin.H{
			"msg" : "请求不合法 p1103",
		})
		return
	}

	resp, err := client.Do(req)
	defer req.Body.Close()

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Status(resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if _, err := c.Writer.Write(body); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

func reqProxy2(c *gin.Context,Pathproxy string){
	targetURL := Pathproxy
	// 创建HTTP客户端
	client := &http.Client{}
	// 创建一个新的请求对象
	req, err := http.NewRequest(c.Request.Method, targetURL, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	query := req.URL.Query()
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	req.URL.RawQuery = query.Encode()

	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(buf)
	req.Body = ioutil.NopCloser(buf)
	resp, err := client.Do(req)

	defer req.Body.Close()

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Status(resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if _, err := c.Writer.Write(body); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

var Maxtry int
func Leg( c *gin.Context){

	if Maxtry > 3 {
		c.JSON(200,gin.H{
			"msg" : "Mode Locked !",
		})
		return
	}

	legname := c.Param("legname");keysfind :=[]interface{}{};var HandleMode,switchon,Proxyto string
	for _,v := range loadconf.Legconf.Legs {
		if v.Legname != legname {
			continue
		}
		if v.Token != c.PostForm("token"){
			c.JSON(200,gin.H{
				"msg" : "请求不合法 l1100",
			})
			return
		}
		HandleMode = v.HandleMode;switchon = v.ReturnResult;Proxyto = v.Proxyto
		if strings.Index(v.VarsFromApi,",") != -1 {
			lip := strings.Split(v.VarsFromApi,",")
			for i:=0;i<cap(lip);i++{
				keysfind = append(keysfind,lip[i])
			}
			break
		}
	}
	if HandleMode == ""{
		c.JSON(200,gin.H{
			"msg" : "请求不合法 l1101 (Max try 3times,then will be lock this mode)",
		})
		Maxtry = Maxtry + 1
		return
	}

	for _,keys := range keysfind {
		varname := keys.(string) + "vars"
		varname  = c.PostForm(keys.(string))
		if varname == ""{
			c.JSON(503,gin.H{
				"error" : "定义了变量:(" + fmt.Sprint(keys.(string)) + "),请求中未找到此变量",
			})
			return
		}
		old := "{{" + keys.(string) + "}}"
		HandleMode = strings.ReplaceAll(HandleMode,old,varname)
	}
	if Proxyto == ""{
		msg := Cmdrun(HandleMode)
		if switchon == "yes" {
			c.JSON(200,gin.H{
				"msg" : msg,
			})
		}else {
			c.JSON(200,gin.H{
				"msg" : "ok",
			})
		}
	}else {
		Pathproxy := "http://" + Proxyto + ":"  +loadconf.ShareConfload.AgentPort + "/rule/legproxy?runmode=" + HandleMode
		reqProxy2(c,Pathproxy)
	}
}

