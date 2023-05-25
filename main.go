package main

import (
	"bytes"
	"example.com/mod/api"
	"example.com/mod/loadconf"
	"example.com/mod/master"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

func start(){
   versionInfo := `v-LIP 1.5`
   fmt.Println(versionInfo,"","START Spider-v1 !")
   time.Sleep(time.Millisecond * 500 )
}


func Con_mysql(userandpasswd,ip string)  (*gorm.DB,error){
	dsn := userandpasswd + "@tcp("+ ip + ")/spider?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local"
	db,err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil{
		return nil,err
	}

	db.AutoMigrate(loadconf.Userinfo{}) //注册信息表

	return db,nil
}



func main(){

	start()
	loadconf.LoadjsonFromLocal()

	if loadconf.ShareConfload.Role != "master" {
		time.Sleep(time.Second * 1)
		if ! loadconf.ContoMaster(){
			fmt.Println("Establishing communication with master Failed !")
			os.Exit(-1)
		}


		go loadconf.Ws2()

		fmt.Println("Load Success")
		loadconf.MakeStart()

		fmt.Println("Injection rule Success")
		gin.SetMode(gin.ReleaseMode)
		server := gin.Default()
		server.Use(CORSMiddleware())
		api.Signrouter(server) //启动总路由
		fmt.Println(" -> Agent.now WorkPort > 0.0.0.0/TCP: " + loadconf.Conf["Port"])
		server.Run(":"+loadconf.Conf["Port"])
		fmt.Println("Start listenTcp error:maybe The port Used")

	}else if loadconf.ShareConfload.Role == "master"{

		master.Cacheip() //读取本地ip列表

		master.CacheProxy() //读取代理配置

		if loadconf.ShareConfload.OpenUiMode == "yes" {
			DBX, err :=	Con_mysql(loadconf.ShareConfload.UiModeStorage,loadconf.ShareConfload.UiModeStorageadd)
			if err != nil {
				fmt.Println("Conn Uistorage failed:",err)
				os.Exit(-1)
			}
			fmt.Println("OpenUi init success")
			master.DB = DBX
			DBX.Create(loadconf.Userinfo{
				Username: "test",
				Passwd: "873012614",
				IndexId: master.CalculateMD5("test"),
				Role: "role",
				Ncalls: 1,
			})
		}

		//go master.CheckAlive() //开启agent拨测

		go master.Forcachespmid() //spmid加载

		gin.SetMode(gin.ReleaseMode)

		server := gin.Default()

		server.Use(CORSMiddleware())

		if loadconf.ShareConfload.RewriteLogAndip == "yes" {
			server.Use(Logger(loadconf.ShareConfload.Spsqladd))
		}

		master.LoadAlertmethod()

		master.Signrouter(server) //启动总路由

		if os.Getenv("sslopen") == "yes" {

			go server.Run(":"+loadconf.Conf["Port"])
			fmt.Println(" -> Master.now WorkPort > 0.0.0.0/TCP: " + loadconf.Conf["Port"])

			fmt.Println(" -> Master.now WorkPort Https > 0.0.0.0/TCP: " + "443")
			server.Use(TlsHandler(443))
			server.RunTLS(":443","proxyai.cn_bundle.pem","proxyai.cn.key")

		}else {
			server.Run(":"+loadconf.Conf["Port"])
			fmt.Println(" -> Master.now WorkPort > 0.0.0.0/TCP: " + loadconf.Conf["Port"])
		}

		fmt.Println("Start listenTcp error:maybe The port Used")
	}
}


func TlsHandler(port int) gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     "0.0.0.0:" + strconv.Itoa(port),
		})
		err := secureMiddleware.Process(c.Writer, c.Request)
		if err != nil {
			return
		}

		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func Logger(add string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始请求的时间
		start := time.Now()
		// 处理请求
		c.Next()
		// 记录结束请求的时间
		end := time.Now()
		// 记录访问日志
		url := add
		method := "POST"
		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)
		_ = writer.WriteField("Timex",start.Format("2006-01-02 15:04:05"))
		_ = writer.WriteField("statuscode",strconv.Itoa(c.Writer.Status()))
		duration := end.Sub(start).Seconds()
		_ = writer.WriteField("Timec",fmt.Sprintf("%.6f",duration))
		_ = writer.WriteField("Fromip", c.ClientIP())
		_ = writer.WriteField("Path", c.Request.URL.Path)
		_ = writer.WriteField("who", "master")
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
}
