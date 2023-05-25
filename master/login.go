package master

import (
	"crypto/md5"
	"encoding/hex"
	"example.com/mod/loadconf"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"strconv"
	"sync"
	"time"
)


func Checkapitoken(c *gin.Context){
	if loadconf.ShareConfload.OpenUiMode != "yes" {
	   c.Next()
	}
	user := c.Query("username")
	token := c.Query("apiToken")

	if token == "" {
		token = c.PostForm("apiToken")
	}

	if token== ""{
		token = c.PostForm("spid")
	}

	tokennum,err := strconv.Atoi(token)
	if err != nil{
		c.JSON(200,gin.H{
			"code" : 04,
			"msg" : "This Token is incorrect Or Timeout,You can login reback",
			"error" : "This Token is incorrect Or Timeout,You can login reback",
		})
		c.Abort()
		return
	}

	if  GetapiToken(user) == 0 || GetapiToken(user) != tokennum {
		fmt.Println( GetapiToken(user),user,token)
		c.JSON(200,gin.H{
			"code" : 03,
			"msg" : "This Token is incorrect Or Timeout,You can login reback ",
			"error" : "This Token is incorrect Or Timeout,You can login reback",
		})
		c.Abort()
		return
	}

	c.Next()
}

func Loginvaild(c *gin.Context){
	data := Tranceforinterface(VenusPost(c,"username","passwd"))
	username,passwd := data[0],data[1]
	userinfo := loadconf.Userinfo{}
	DB.Where("username = ? and passwd = ?",username,passwd).First(&userinfo)
	if userinfo.Username == "" || userinfo.Passwd == "" {
		c.JSON(200,gin.H{
			"code": 01,
			"msg" : "认证失败 请联系管理员",
		})
	}else {
		MakeapiToken(username)
		c.JSON(200,gin.H{
			"msg" : "认证成功",
			"apiToken" : GetapiToken(username),
			"Authcode" : MakeauthCode(),
		})
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

func TimeForapiToken(user string)bool{
	tiemx , ok := TokenTime.Load(user)
	if ok {
	  if !	panduantime(tiemx.(string),20000){
		  return false
	  }
	  return true
	}
	return false
}

func GetapiToken(user string)int{
	token,ok :=Fastsearch.Load(user)
	if ok {
		if TimeForapiToken(user){
			return token.(int)
		}
		return 0
	}
	return  0
}

var Fastsearch sync.Map
var TokenTime  sync.Map




func Sign(c *gin.Context){

	//if c.ClientIP() != "127.0.0.1" {
	//	c.JSON(404,gin.H{})
	//	return
	//}

	data := Tranceforinterface(VenusPost(c,"username","passwd","token"))
	username,passwd,token := data[0],data[1],data[2]
	fmt.Println(username,passwd,token)
	if token != "null" {

		userindo := loadconf.Userinfo{}
		DB.Where("index_id = ?",CalculateMD5(username)).First(&userindo)
		if userindo.Username == ""{
			DB.Create(loadconf.Userinfo{
				Username: username,
				Passwd: passwd,
				IndexId: CalculateMD5(username),
				Role: "role",
				Ncalls: 0,
			})
			c.JSON(200,gin.H{
				"msg" :"success",
			})
			return
		}else {
			c.JSON(200,gin.H{
				"code": 05,
				"msg" :"This User already exists",
			})
		}
	}else {
		userindo := loadconf.Userinfo{}
		DB.Where("index_id = ?",CalculateMD5(username)).First(&userindo)
		if userindo.Username == ""{
			DB.Create(loadconf.Userinfo{
				Username: username,
				Passwd: passwd,
				IndexId: CalculateMD5(username),
				Role: "role",
				Ncalls: 1,
			})

			c.JSON(200,gin.H{
				"msg" :"success",
			})
			return
		}else {
			c.JSON(200,gin.H{
				"code": 051,
				"msg" :"This User already exists",
			})
		}
	}
}

func CalculateMD5(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}



func MakeauthCode()int{
	rand.Seed(time.Now().UnixNano()) // 设置随机种子
	num := rand.Intn(900) + 100
	return num
}

func MakeapiToken(user string) {
	rand.Seed(time.Now().UnixNano()) // 设置随机种子
	num := rand.Intn(9000000000000000) + 1000000000000000
	Fastsearch.Store(user,num)
	TokenTime.Store(user,time.Now().Format("2006-01-02 15:04:05"))
}