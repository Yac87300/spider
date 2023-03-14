package main

import (
	"example.com/mod/api"
	"example.com/mod/loadconf"
	"example.com/mod/master"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"time"
)

func start(){
   versionInfo := `v-LIP 1.3`
   fmt.Println(versionInfo,"","START Spider-v1 !")
   time.Sleep(time.Millisecond * 500 )
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

		go master.CheckAlive() //开启agent拨测

		go master.Forcachespmid() //spmid加载

		gin.SetMode(gin.ReleaseMode)

		server := gin.Default()

		server.Use(CORSMiddleware())

		master.LoadAlertmethod()

		master.Signrouter(server) //启动总路由

		fmt.Println(" -> Master.now WorkPort > 0.0.0.0/TCP: " + loadconf.Conf["Port"])

		server.Run(":"+loadconf.Conf["Port"])

		fmt.Println("Start listenTcp error:maybe The port Used")
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
