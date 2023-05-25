package mysqlpool

import (
	"example.com/mod/master"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
)

func GetGinlogs(c *gin.Context){
	data := master.Tranceforinterface(master.VenusPost(c,"Timex","statuscode","Timec","Fromip","Path","who"))
	Timex,statuscode,Timec,Fromip,Path,who := data[0],data[1],data[2],data[3],data[4],data[5]
	fmt.Println(Timex,statuscode,Timec,Fromip,Path,who)
    fmt.Println(Timex)
	layout := "2006-01-02 15:04:05"


	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		fmt.Println(err)
		return
	}

	timex, err := time.ParseInLocation(layout, Timex,loc)
	if err != nil {
		fmt.Println(err)
		return
	}
	//timec , _ := strconv.Atoi(Timec)
	//fmt.Println(timec)
	floatvar, err := strconv.ParseFloat(Timec, 64)
	statuscodes,_ := strconv.Atoi(statuscode)


	mdata := Gin{
			StatusCode:    statuscodes,
			Timex:         timex,
			Timeconsuming: floatvar,
			Fromip:        Fromip,
			Path:          Path,
			Who:           who,
		}
	err = DB.Model(&Gin{}).Create(mdata).Error

	if err != nil{
		c.JSON(503,gin.H{
			"msg" : "log failed",
		})
		return
	}

	c.JSON(200,gin.H{
		"msg" : "ok",
	})
}

func GetIplistAndUpdateStatus( c *gin.Context){
	data := master.Tranceforinterface(master.VenusPost(c,"ip","group","status"))
	ip,group,status := data[0],data[1],data[2]

	list := []Iplist{}
	lists := ""
	DB.Where("ip = ?",ip).First(&list)
	for _,v := range append(list){
		lists = v.Ip
	}

	mdata := Iplist{
		Ip: ip,
		IPGroup: group,
		Status: status,
	}
	if len(lists) == 0 {

		err := DB.Model(&Iplist{}).Create(mdata).Error

		if err != nil{
			c.JSON(503,gin.H{
				"msg" : "log failed",
			})
			return
		}
		c.JSON(200,gin.H{
			"msg" : "ok",
		})
	}else {
		err := DB.Model(&Iplist{}).Clauses(clause.Returning{}).Where("ip = ?", ip).Updates(Iplist{
			Status: status,
			IPGroup: group,
		}).Error
		if err != nil{
			c.JSON(503,gin.H{
				"msg" : "log failed",
			})
			return
		}
		c.JSON(200,gin.H{
			"msg" : "ok",
		})
	}
}


func Taskswirte(c *gin.Context){
	data := master.Tranceforinterface(master.VenusPost(c,"taskid","taskname","taskresult","Taskstatus","Datafrom","Tasktime","Alertmethod","AlertData","Alerttype","Aletstatus","Fromip"))
	taskid,taskname,taskresult,Taskstatus,Datafrom,Tasktime,Alertmethod,AlertData,Alerttype,Aletstatus,Fromip := data[0],data[1],data[2],data[3],data[4],data[5],data[6],data[7],data[8],data[9],data[10]
    fmt.Println(taskid,taskname,taskresult,Taskstatus,Datafrom,Tasktime,Alertmethod,AlertData,Alerttype,Aletstatus,Fromip)
	mdata := Tasks{
		Taskid:     taskid,
		Taskname:    taskname,
		Taskresult:  taskresult,
		Taskstatus:  Taskstatus,
		Datafrom:    Datafrom,
		Tasktime:    Tasktime,
		Alertmethod: Alertmethod,
		AlertData:   AlertData,
		Alerttype:   Alerttype,
		Aletstatus:  Aletstatus,
		Fromip: Fromip,
	}

	list := []Tasks{}
	lists := ""
	DB.Where("taskid = ?",taskid).First(&list)
	for _,v := range append(list){
		lists = v.Taskid
	}
	fmt.Println(lists)

	if len(lists) == 0 {

		err := DB.Model(&Tasks{}).Create(mdata).Error

		if err != nil{
			c.JSON(503,gin.H{
				"msg" : "log failed",
			})
			return
		}
		c.JSON(200,gin.H{
			"msg" : "ok",
		})
	}else {
		err := DB.Model(&Tasks{}).Clauses(clause.Returning{}).Where("taskid = ?",taskid).Updates(Tasks{
			Taskresult:  taskresult,
			Taskstatus:  Taskstatus,
			Datafrom:    Datafrom,
			Tasktime:    Tasktime,
			Alertmethod: Alertmethod,
			AlertData:   AlertData,
			Alerttype:   Alerttype,
			Aletstatus:  Aletstatus,
			Fromip: Fromip,
		}).Error
		if err != nil{
			c.JSON(503,gin.H{
				"msg" : "log failed",
			})
			return
		}
		c.JSON(200,gin.H{
			"msg" : "ok",
		})
	}
}


func Update(ip string,data interface{}) error {
	return DB.Model(&data).Where("ip = ?", ip).Updates(data).Error
}

func Create(data interface{}) error {
	result := DB.Create(data.(Gin))
	return result.Error
	return nil
}
