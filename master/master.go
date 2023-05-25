package master

import (
	"bufio"
	"bytes"
	"encoding/json"
	"example.com/mod/loadconf"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)


var blacklist sync.Map
var iplist sync.Map
var Offline sync.Map
var proxycomp sync.Map
var spmids sync.Map
var download sync.Map

var DB *gorm.DB


func Bindcheck(c *gin.Context){
	c.Next()
	//spid := c.Query("spid")
	//data,err := Jiemi("19a7251c679b22ccf83bd8a9709910be",spid)
	//fmt.Println(data,spid)
	//
	//if err != nil{
	//	c.JSON(503,gin.H{
	//		"error" : "Your cliTools no spid/spid notright to make token,stop your cliTools,Please call the admin to Add.",
	//	})
	//	c.Abort()
	//	return
	//}
	//
	//lip := strings.Split(data,":")
	//if c.ClientIP() != lip[0] {
	//	fmt.Println(c.ClientIP(),lip[0])
	//	c.JSON(503,gin.H{
	//		"error" : "Your spid not right.",
	//	})
	//	c.Abort()
	//	return
	//}
	//
	//k,ok := spmids.Load(data)
	//fmt.Println("[cliTools]:",k,data,"From:",c.ClientIP(),"To:",c.FullPath())
	//if ok {
	//	c.Next()
	//}else {
	//	c.JSON(503,gin.H{
	//		"error" : "Your cliTools no spid/spid notright to make token,stop your cliTools,Please call the admin to Add.",
	//	})
	//	c.Abort()
	//	return
	//}
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

func isInternalIP(ip string) bool {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return false
	}

	// IPv4的内网IP地址范围
	privateIPBlocks := []*net.IPNet{
		// 10.0.0.0 - 10.255.255.255
		{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		// 172.16.0.0 - 172.31.255.255
		{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		// 192.168.0.0 - 192.168.255.255
		{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
	}

	for _, block := range privateIPBlocks {
		if block.Contains(netIP) {
			return true
		}
	}

	// IPv6的内网IP地址范围（根据实际情况进行配置）
	// privateIPv6Blocks := []*net.IPNet{
	// 	// ...
	// }

	// for _, block := range privateIPv6Blocks {
	// 	if block.Contains(netIP) {
	// 		return true
	// 	}
	// }

	return false
}

func CheckIP(c *gin.Context){
	 if isInternalIP(c.ClientIP()) == false && c.ClientIP() != "127.0.0.1"{
		 fmt.Println(c.ClientIP(),"c.abort")
	 	c.JSON(200,gin.H{
	 		"code" : "无效的路由",
		})
	 	c.Abort()
		 return
	 }
	 fmt.Println(c.ClientIP(),"c.next()")
	 c.Next()
}





func Signrouter(server *gin.Engine){


	if loadconf.ShareConfload.OpenUiMode == "yes"{
		server.POST("/master/auth",Loginvaild)
	}

	//server.POST("/master/sm4",testjiami)
	server.POST("/master/add",CheckIP,Addip)
	server.POST("/master/alert",CheckIP,RunAlertMethod)
	server.GET("/master/keep/alive",CheckIP,Keepalived)
	server.POST("/master/proxy/:proxyname",proxy)
	server.GET("/master/proxy/:proxyname",proxy)

	masterGroup := server.Group("/master",Checkapitoken)
	{
		masterGroup.GET("/ping", func(c *gin.Context) {

			c.JSON(200,gin.H{
			 	"msg" : "1",
			 })
		})

		masterGroup.POST("/ping2", func(c *gin.Context) {
			msg := c.PostForm("msg")
			token := c.PostForm("token")
			rand.Seed(time.Now().UnixNano())
			sleepTime := rand.Intn(3) + 1

			// 睡眠随机秒数
			time.Sleep(time.Duration(sleepTime) * time.Second)

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
		masterGroup.POST("/showmethod",Showmethod)
		masterGroup.POST("/syncstatus",Sync)
		masterGroup.POST("/leg/:legname",Leg)
		masterGroup.POST("/group",Managegroup)
		masterGroup.POST("/groupip",ManageGroupip)
		masterGroup.GET("/file/create",handleWebSocket)
		masterGroup.GET("/file/download",Downdload)
		masterGroup.GET("/show/proxy",ShowPorxy)
		masterGroup.GET("/show/leg",ShowLeg)
		masterGroup.POST("/master/proxy/edit/:proxyname")
	}
}




func ShowLeg(c *gin.Context){
	local := loadconf.Legconf
	c.JSON(200,local)
}

func ShowPorxy(c *gin.Context){

	local := loadconf.ShareConfload.GatewayProxy

	show := []loadconf.Proxyshow{}

	for _,v := range local{
		show = append(show,loadconf.Proxyshow{H: v.H,OutPath: v.OutPath,InsidePath: v.InsidePath,Servers: v.Servers,
			Tokencheck: v.Tokencheck,Status: "ok"})
	}
	c.JSON(200,show)
}


func SendFunctionFileToagent(filename string, targetUrl string) error {
	// 打开要上传的文件
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	// 创建一个带有文件内容的缓冲区
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err = io.Copy(part, file); err != nil {
		return err
	}
	writer.Close()

	// 创建HTTP请求
	req, err := http.NewRequest("POST", targetUrl, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// 发送HTTP请求
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	// 检查响应状态码
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload file, status code: %d", res.StatusCode)
	}
	return nil
}

func SyncSAgentStatus(ip,group,status,add string){
	url := add
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("ip", ip)
	_ = writer.WriteField("group", group)
	_ = writer.WriteField("status", status)
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
		fmt.Println("syncipstatus",err)
		return
	}
	//fmt.Println(string(body))
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}


func Keepalived(c *gin.Context) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 在这里判断请求来源是否允许
			return true
		},
	}
	sip := c.ClientIP()
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	// 在这里处理 WebSocket 连接逻辑
	err = conn.WriteMessage(websocket.TextMessage,[]byte("Master: established a connection with you"))
	if err != nil {
	}
	Offline.Delete(sip)

	data , ok := loadconf.GroupipCache.Load(sip)
	if ok {
		SyncSAgentStatus(sip,data.(string),"online",loadconf.ShareConfload.Spsqlipadd)
	}else {
		SyncSAgentStatus(sip,"NULL","online",loadconf.ShareConfload.Spsqlipadd)
	}

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(sip,"断开")
			Offline.Store(sip,"Offonline")
			data , ok := loadconf.GroupipCache.Load(sip)
			if ok {
				SyncSAgentStatus(sip,data.(string),"offonline",loadconf.ShareConfload.Spsqlipadd)
			}else {
				SyncSAgentStatus(sip,"NULL","offonline",loadconf.ShareConfload.Spsqlipadd)
			}
			break
		}
		if messageType == websocket.TextMessage {
			message := string(p)

			if message == "update"{

			}

		} else if messageType == websocket.BinaryMessage {

		}

		err = conn.WriteMessage(websocket.TextMessage,[]byte("established"))
		if err != nil {
			break
		}
		fmt.Println("keep")
	}
}

func handleWebSocket(c *gin.Context) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 在这里判断请求来源是否允许
			return true
		},
	}
	filex :=  c.Query("filepath")
	data,_:=Jiami("19a7251c679b22ccf83bd8a9709910be",filex)
	download.Store(data,filex)
	fmt.Println(data,filex)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	// 在这里处理 WebSocket 连接逻辑

	err = conn.WriteMessage(websocket.TextMessage,[]byte(data))
	if err != nil {
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(filex,"断开")
			download.Delete(data)
			break
		}

		err = conn.WriteMessage(websocket.TextMessage,[]byte(filex))
		if err != nil {

			break
		}
		fmt.Println("keep")
	}
}

func Downdload(c *gin.Context){
	 tip := c.Query("tip")
	 filex := c.Query("filepath")
	 filex2 := strings.ReplaceAll(filex," ","+")
	 fmt.Println(filex,filex2)
	 key,ok := download.Load(filex2)
	 if !ok{
	 	 c.JSON(200,gin.H{
	 	 	"error" : "can't find this path",
		 })
		 return
	 }
	filename := filepath.Base(key.(string))
	fmt.Println(filename,key)

	if tip == "size" {
		fileinfo, err := os.Stat(key.(string))
		if err != nil {
			fmt.Println("获取文件信息失败:", err)
			c.JSON(200,gin.H{
				"error" : err,
			})
			return
		}
		filesize := fileinfo.Size()
		c.JSON(200,gin.H{
			"msg" : fmt.Sprint(filesize),
		})
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+filename) // 用来指定下载下来的文件名
	c.Header("Content-Transfer-Encoding", "binary")
	c.File(key.(string))
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
		Offline.Store(fileScanner.Text(),"offonline")
		data , ok := loadconf.GroupipCache.Load(fileScanner.Text())
		if ok {
			if loadconf.ShareConfload.RewriteLogAndip == "yes" {
				SyncSAgentStatus(fileScanner.Text(),data.(string),"offonline",loadconf.ShareConfload.Spsqlipadd)
			}
		}else {
			if loadconf.ShareConfload.RewriteLogAndip == "yes" {
				SyncSAgentStatus(fileScanner.Text(),"null","offonline",loadconf.ShareConfload.Spsqlipadd)
			}
		}

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
				ipk , ok2 := loadconf.GroupipCache.Load(key.(string))
				if  ipk == grouptag{
					rule,ok := Getrule(key.(string) +":" +loadconf.ShareConfload.AgentPort,loadconf.Conf["Passwd"],"/rule/getrs")
					if ok {
						crulelist.Data = append(crulelist.Data,rule)
						return true
					}
				}else {
					return ok2
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
	proxynames := c.Param("proxyname");Pathproxy := ""
	_,ok := proxycomp.Load(proxynames)
	if ok {
	   	H,_ := proxycomp.Load(proxynames+"H")
	   	Server ,_ := proxycomp.Load(proxynames+"Servers")
			InsidePathPort,_:= proxycomp.Load(proxynames+"InsidePathPort")
			InsidePath,_:= proxycomp.Load(proxynames+"InsidePath")
			Pathproxy = H.(string) + "://" +  Server.(string) + ":" + InsidePathPort.(string) + InsidePath.(string)
	}else {
		c.JSON(503,gin.H{
			"msg" : "请求不合法 p1101",
		})
		fmt.Printf("请求不合法 p1101")
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
		c.JSON(503,gin.H{
			"msg" : "请求不合法 p1102",
		})
		fmt.Printf("请求不合法 p1102")
		return
	}

	x := strings.Split(fmt.Sprint(req.Body),"\"token\"")
	v := strings.Split(x[1],"-")
	tokens := strings.ReplaceAll(v[0]," ","")
	usertoken,_ := proxycomp.Load(Proxyname+"Tokencheck")
	if strings.Index(tokens,usertoken.(string)) == -1 {
		c.JSON(503,gin.H{
			"msg" : "请求不合法 p1103",
		})
		fmt.Printf("请求不合法 p1103")
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

