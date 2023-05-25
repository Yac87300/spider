package mysqlpool

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)



var DB *gorm.DB

func Con_mysql(userandpasswd,ip string)  (*gorm.DB,error){
	dsn := userandpasswd + "@tcp("+ ip + ")/spider?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local"
	db,err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil{
		return nil,err
	}
	db.AutoMigrate(Tasks{}) //
	db.AutoMigrate(Gin{})
	db.AutoMigrate(Iplist{})
	db.AutoMigrate(Grouplist{})
	return db,nil
}

type Tasks struct {
	Taskid      string `json:"Taskid"`
	Fromip      string `json:"fromip"`
	Taskname    string `json:"Taskname" gorm:"charset=utf8mb4"`
	Taskresult  string `json:"Taskresult"`
	Taskstatus  string `json:"Taskstatus"`
	Datafrom    string `json:"Datafrom"`
	Tasktime    string `json:"Tasktime"`
	Alertmethod string `json:"Alertmethod"`
	AlertData   string `json:"AlertData"`
	Alerttype   string `json:"Alerttype"`
	Aletstatus  string `json:"aletstatus"`
}

type Gin struct {
	StatusCode int `json:"status_code"`
	Timex time.Time `json:"timex"`
	Timeconsuming float64 `json:"timeconsuming"`
	Fromip string `json:"fromip"`
	Path string `json:"path"`
	Who string `json:"who"`
}

type Iplist struct {
	Ip string `json:"ip"`
	Status string `json:"status"`
	IPGroup string `json:"ip_group"`
}

type Grouplist struct {
	Groupname string
}
