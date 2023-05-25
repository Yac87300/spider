package master

import (
    "bufio"
    "encoding/base64"
    "encoding/hex"
    "example.com/mod/loadconf"
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/tjfoc/gmsm/sm4"
    "io/ioutil"
    "os"
    "os/exec"
    "reflect"
    "strconv"
    "strings"
    "sync"
    "syscall"
    "time"
)



var AlertMethod sync.Map

//Name      string `json:"Name"`
//DataType  string `json:"DataType"`
//From      string `json:"From"`
//Alert     string `json:"Alert"`
//Alertdata string `json:"Alertdata"`
//AlertTo   string `json:"AlertTo"`
//ForTime   int    `json:"ForTime"`
//Status string `json:"status"`
//Time string `json:"time"`
//Rs string `json:"rs"`

var Msgstrings = []string{"Name","DataType","From","Alert","Alertdata","AlertTo","ForTime","Status","Time","Rs"}

func LoadAlertmethod(){
  for _,v := range loadconf.ShareConfload.AlertMethod {
       AlertMethod.Store(v.Path,v.RunMode)
  }
}


func RunAlertMethodWithMd5(c *gin.Context){
    k := Tranceforinterface(VenusPost(c,"id","timex"))
    id,timex := k[0],k[1]
    msg := ""
    iplist.Range(func(key, value interface{})bool{
        _,ok := Offline.Load(key.(string))
        if ok{

        }else {
            rule,ok := Getrule(key.(string) +":" +loadconf.ShareConfload.AgentPort,loadconf.Conf["Passwd"],"/rule/getrs")
            if ok {
                for _,v := range rule.Rulesource{
                    if v.Md5c == id{

                        for {
                            _,ok := blacklist.Load(rule.Tag+":"+v.Name+"skey")
                            _,ok2 := blacklist.Load(rule.Tag+":"+v.Name)
                            if ok {
                                blacklist.Delete(rule.Tag+":"+v.Name+"skey")

                            }else if !ok && !ok2{
                                break
                            }else {
                                fmt.Println("等待消息队列删除")
                            }
                            time.Sleep(time.Second * 1)
                        }

                        msg ="ok"
                        blacklist.Store(rule.Tag+":"+v.Name+"skey","open")
                        go deferDelete(rule.Tag,v.Name,timex)
                        break
                    }
                }

            }
        }
        return true
    })

    if msg != "ok" {
        msg = "Can't find This ID in SpiderPool"
    }

    c.JSON(200,gin.H{
        "master" :msg,
    })

}

func OpendelayAlertMethodWithMd5(c *gin.Context){
    k := Tranceforinterface(VenusPost(c,"id"))
    id := k[0]
    msg := ""
    iplist.Range(func(key, value interface{})bool{
        _,ok := Offline.Load(key.(string))
        if ok{

        }else {
            rule,ok := Getrule(key.(string) +":" +loadconf.ShareConfload.AgentPort,loadconf.Conf["Passwd"],"/rule/getrs")
            if ok {
                for _,v := range rule.Rulesource{
                    if v.Md5c == id{
                        for {
                            _,ok := blacklist.Load(rule.Tag+":"+v.Name+"skey")
                            _,ok2 := blacklist.Load(rule.Tag+":"+v.Name)
                            if ok {
                                blacklist.Delete(rule.Tag+":"+v.Name+"skey")
                            }else if !ok && !ok2{
                                break
                            }else {
                                fmt.Println("等待消息队列删除")
                            }
                            time.Sleep(time.Second * 1)
                        }
                        msg ="ok"
                        break
                    }
                }

            }
        }
        return true
    })

    if msg != "ok" {
        msg = "Can't find This ID in SpiderPool"
    }

    c.JSON(200,gin.H{
        "master" :msg,
    })

}


func RunAlertMethod(c *gin.Context){

    k := Tranceforinterface(VenusPost(c,"methodName","ip","name","msg"))
    methodname,ip,name,msg := k[0],k[1],k[2],k[3]
    if _,ok := blacklist.Load(ip+":"+name);ok {
        c.JSON(200,gin.H{
            "master" :name + "The rule is asleep ",
        })
        return
    }
    if _,ok := iplist.Load(ip);!ok {
        c.JSON(200,gin.H{
            "master" : "0 failed",
        })
        return
    }

    _,ok := AlertMethod.Load(methodname)

    if !ok {
        c.JSON(200,gin.H{
            "master" : "1 failed",
        })
        return
    }
    rule,ok := Getrule(ip +":" +loadconf.ShareConfload.AgentPort,loadconf.Conf["Passwd"],"/rule/getrs")
    if !ok {
        return
    }
    msg = strings.Replace(msg,"Ip",ip,1)
    for _,repString := range Msgstrings{
        for _,v := range rule.Rulesource{
            if v.Name != name{
                continue
            }
            ret := reflect.ValueOf(v)
            bs := ret.FieldByName(repString)
            if ok {
                msg = strings.Replace(msg,repString,bs.String(),1) //生成消息
            }
        }
    }

    cmdrun,_ := AlertMethod.Load(methodname)

    end := methodname + "" + strings.ReplaceAll(cmdrun.(string),"{{msg}}",msg)

    if strings.Index(methodname,",") != -1 {
        lip := strings.Split(methodname,",")
        for i:=0;i<cap(lip);i++{
            ALertRun(lip[i])
        }
    }else {
        ALertRun(end)
    }


}


func AlertLog(){

}

func ALertRun(cmds string){
    fmt.Println(cmds)
    cmd := exec.Command("/bin/bash", "-c", cmds)
    //创建获取命令输出管道
    cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
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

    err = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
    cmd.Process.Signal(syscall.SIGKILL)
    err = cmd.Wait()
    fmt.Println(string(bytes))
}

func Cmdrun(cmds string)string{
    cmd := exec.Command("/bin/bash", "-c", cmds)
    //创建获取命令输出管道
    stdout, err := cmd.StdoutPipe()
    cmd.Stderr = cmd.Stdout
    if err != nil {
        fmt.Printf("Error:can not obtain stdout pipe for command:%s\n", err)
        return fmt.Sprint("Error:can not obtain stdout pipe for command:%s\n", err)
    }

    //执行命令
    if err := cmd.Start(); err != nil {
        fmt.Println("Error:The command is err,", err)
        return  fmt.Sprint("Error:The command is err,", err)
    }

    //读取所有输出
    bytes, err := ioutil.ReadAll(stdout)
    if err != nil {
        fmt.Println("ReadAll Stdout:", err.Error())
        return  fmt.Sprint("ReadAll Stdout:", err.Error())
    }
    fmt.Println("[LegsCMD]",string(bytes))
    return string(bytes)
}


func SetBlackAlert(c *gin.Context){
    k := Tranceforinterface(VenusPost(c,"ip","alertname","time"))
    ip,alertnmae,times := k[0],k[1],k[2]
    blacklist.Delete(ip+":"+alertnmae)
    go deferDelete(ip,alertnmae,times)
    c.JSON(200,gin.H{
        "master" : "ok",
    })

}

func StartBlackAlert(c *gin.Context){
    k := Tranceforinterface(VenusPost(c,"ip","alertname"))
    ip,alertnmae := k[0],k[1]
    blacklist.Delete(ip+":"+alertnmae)
    blacklist.Store(ip+":"+alertnmae+"skey","open")
    c.JSON(200,gin.H{
        "master" : "ok",
    })
}

func deferDelete(ip,alertname,times string){
     blacklist.Store(ip+":"+alertname,time.Now().Format("2006-01-02 15:04:05")+" -> After: "+times+"s")
     timess,_ := strconv.Atoi(times)
     //time.Sleep(time.Second * time.Duration(timess))
     for i:=0;i<timess;i++{
         _,ok := blacklist.Load(ip+":"+alertname+"skey")
         if !ok{
             blacklist.Delete(ip+":"+alertname)
             return
         }
         time.Sleep(time.Second * 1)
     }
     blacklist.Delete(ip+":"+alertname)
}

func Showdefer(c *gin.Context){
    ciplist := Iplist{}
    blacklist.Range(func(key, value interface{}) bool {
            ciplist.Iplists = append(ciplist.Iplists,Data{DelayName:key.(string),DelayData: value.(string)})
        return true
    })

    c.JSON(200,ciplist)
}


func Jiami(hexKey, raw string) (string, error) {
    key, err := hex.DecodeString(hexKey)
    if err != nil {
        return "", err
    }
    out, err := sm4.Sm4Ecb(key, []byte(raw), true)
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(out), nil
}

func Jiemi(hexKey, base64Raw string) (string, error) {
    if base64Raw == "" {
        return "", nil
    }

    key, err := hex.DecodeString(hexKey)

    if err != nil {
        return "", err
    }

    raw, err := base64.StdEncoding.DecodeString(base64Raw)

    if err != nil {
        return "", err
    }

    out, err := sm4.Sm4Ecb(key, raw, false)
    if err != nil {
        return "", err
    }
    return string(out), nil

}

func cachespmid(){
    file, err := os.Open("./cache/spmid")
    if err != nil {
        fmt.Println("rule err:",err)
        os.Exit(-1)
    }
    fileScanner := bufio.NewScanner(file)
    for fileScanner.Scan() {
        if fileScanner.Text() == ""{
            continue
        }
        //ip:name:time
        data,err := Jiemi("19a7251c679b22ccf83bd8a9709910be",fileScanner.Text())
        if err != nil{
            fmt.Println(fileScanner.Text(),"is not a spmid")
            continue
        }
        spmids.Store(data,"_load#")
    }
}

func Forcachespmid(){
    for {
        cachespmid()
        time.Sleep(time.Second * 10)
    }
}



