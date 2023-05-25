package main

import (
	"example.com/mod/mysql/mysqlpool"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

var DBX *gorm.DB

func main(){

	u := flag.String("u", "", "user of mysql")
	p := flag.String("p","","passwd of mysql")
	h := flag.String("h","127.0.0.1","remote con2 mysql,address")
	P := flag.String("P","3306","mysql Port")

	flag.Parse()
	up := *u + ":" + *p
	add := *h +":" + *P

	DBX, err := mysqlpool.Con_mysql(up, add)
	if err != nil {
		return
	}

	result := DBX.Exec("select * from gins")
	fmt.Println(result)

	mysqlpool.DB = DBX

	server := gin.Default()
	server.Use(CORSMiddleware())
	server.POST("/wtask",mysqlpool.Taskswirte)
	server.POST("/wginlog",mysqlpool.GetGinlogs)
	server.POST("/wip",mysqlpool.GetIplistAndUpdateStatus)
	server.GET("/ping", func(c *gin.Context) {
		c.JSON(200,gin.H{
			"msg":"pong",
		})
	})
	fmt.Println("> Start WlB OK √ ！")
	server.Run(":1212")
}






