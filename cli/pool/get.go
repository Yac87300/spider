package pool

import (
	"bufio"
	"bytes"
	"encoding/json"
	"example.com/mod/loadconf"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Rulelist struct {
	Msg string `json:"msg"`
	Error string `json:"error"`
	Data []loadconf.Rulejson `json:"data"`
}


type noLogTransport struct{}

func (t *noLogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Disable logging by setting error output to ioutil.Discard
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		resp.Body.Close()
	}
	return resp, err
}

func Getleg(){
	fmt.Printf("%-20s%-40s%-40s\n","Name","Vars","HandleMode")
	url := "http://"+ GetmasterAddress() + "/master/show/leg" + "?apiToken=" + Getmasterspid() +"&username=" + GetUser()
	method := "GET"
	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
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
	list := loadconf.Legst{}
	err = json.Unmarshal(body,&list)
	for _,v := range list.Legs{
		fmt.Printf("%-20s%-40s%-40s\n",v.Legname,v.VarsFromApi,v.HandleMode)
	}
}
func Getproxy(){
    	fmt.Printf("%-30s%5s\n","Method","Path -> Servers")
		url := "http://"+ GetmasterAddress() + "/master/show/proxy" + "?apiToken=" + Getmasterspid() +"&username=" + GetUser()
		method := "GET"
		client := &http.Client {
		}
		req, err := http.NewRequest(method, url, nil)

		if err != nil {
			fmt.Println(err)
			return
		}
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
		list := []loadconf.Proxyshow{}
		err = json.Unmarshal(body,&list)
	    for _,v := range list{

		fmt.Printf("%-30s%5s\n",v.H,"/master/proxy/"+v.OutPath+" -> "+v.Servers+v.InsidePath)
	   }
}


func Download(filepath string){
	url := "http://"+ GetmasterAddress() +"/master/file/download/" + "?spid=" + Getmasterspid() + "&filepath=" + filepath
	method := "GET"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
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
	fmt.Println("Send Key to CryptoDownloadFile")

	fs := strings.Split(res.Header.Get("Content-Disposition"),";")
	fs2 := strings.Split(fs[1],"=")

	// 打开文件以供写入
	output, err := os.Create(fs2[1])
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer output.Close()
	reader := bufio.NewReader(res.Body)
	var wg sync.WaitGroup
	go Getsize(filepath,fs2[1],&wg)

	start := time.Now()
	buf := make([]byte,1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading from response body:", err)
			return
		}

		if n > 0 {
			_, err = output.Write(buf[:n])
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
		}
	}

	wg.Wait()
	end := time.Now()
	// 计算函数的处理时间
	duration := end.Sub(start)
	fmt.Println("\nDownLoadFile Times:",duration)

	//body, err := ioutil.ReadAll(res.Body)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//list := Mastejson{}
	//err = json.Unmarshal(body,&list)
	//if err != nil {
	//	//fmt.Println("err1:",err)
	//	//os.Exit(-1)
	//}
	//if list.Error != ""{
	//	fmt.Println("err2:",list.Error)
	//	os.Exit(-1)
	//}
	//_ = Writefile(fs2[1],"")

}

func Getsize(filepath,filename string, wg *sync.WaitGroup){
	url := "http://"+ GetmasterAddress() +"/master/file/download/" + "?spid=" + Getmasterspid() + "&filepath=" + filepath +"&tip=size"
	method := "GET"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
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
	list := Mastejson{}
	err = json.Unmarshal(body,&list)
	if err != nil {
		fmt.Println("err2:",err)
		fmt.Println(string(body))
		os.Exit(-1)
	}
	if list.Error != ""{
		fmt.Println("err:",list.Error)
		os.Exit(-1)
	}

	wg.Add(1)
	old := float64(1)

	var bytesWritten float64 = 0
	start := time.Now()
	func(){
		for {
			fileinfo, err := os.Stat(filename)
			if err != nil {
                  continue
			}
			filesize := fileinfo.Size()
			sd := float64(filesize) - old
			total,err := strconv.Atoi(list.Msg)
			totalx := total / 1024 / 1024
			bfs := float64(filesize) / float64(total) * 100
			xz := float64(filesize) - float64(total)

			bytesWritten += sd
			elapsed := time.Since(start).Seconds()
			speed := float64(bytesWritten) / elapsed / 1024 /1024

			//fmt.Printf("\rWriting speed: %.2f bytes/sec", speed)

			//fmt.Printf("\rFileName:%-8s Total:%sMb -> [%.2f%%]",filename,fmt.Sprint(totalx),bfs)
			msg := fmt.Sprintf("FileName: %-10s Total:%8sMb  -> [%6.1f%%] <- %6.1f mb/s\r", filename,fmt.Sprint(totalx),bfs,speed)
			fmt.Print(msg)
			old = float64(filesize)


			//msg := fmt.Sprintf("FileName:%-8s Total:%sMb -> [%.2f%%]", filename, fmt.Sprint(totalx), bfs)
			//fmt.Print("\rS:" + msg)

			if xz == 0 || xz == 5400{
				wg.Done()
				break
			}
			time.Sleep(time.Millisecond * 50)
		}
		fmt.Println("..")
	}()

}


func Ws2(filepath string){
	url := "ws://" + GetmasterAddress() + "/master/file/create" + "?spid=" + Getmasterspid() + "&filepath=" + filepath
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("连接错误：", err)
		return
	}
	go func() {
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("读取消息错误：", err)
				return
			}
			if messageType == websocket.TextMessage {
				message := string(p)
				fmt.Println("fileid：", message)
			} else if messageType == websocket.BinaryMessage {
				fmt.Println("else")
			}
		}
	}()
	time.Sleep(time.Minute * 60)
	defer conn.Close()
}


func Doleg(legname,cs string) {
	url :=  "http://"+ GetmasterAddress() +"/master/leg/" + legname + "?apiToken=" + Getmasterspid() +"&username=" + GetUser()
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	if strings.Index(cs,",") != -1 {
		lip := strings.Split(cs,",")
		for i:=0;i<cap(lip);i++{
			data := strings.Split(lip[i],"=")
			key := data[0]
			va := data[1]
			_ = writer.WriteField(key, va)
		}
	}else {
		data := strings.Split(cs,"=")
		key := data[0]
		va := data[1]
		_ = writer.WriteField(key, va)
	}


	err := writer.Close()
	if err != nil {
		fmt.Println("00",err)
		return
	}

	client := &http.Client {}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println("01",err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)

	if err != nil {
        fmt.Println("Can't reqdo to ",GetmasterAddress(),"check address and spmid")
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {

		return
	}

	legs := Leg{}
	_ = json.Unmarshal(body,&legs)
	if legs.Error != ""{
		fmt.Println(legs.Error)
		return
	}
	fmt.Println(legs.Msg)
}

func Getrulelist(){
	fmt.Printf("%-17s%6s\n","IP","Rule")
	url := "http://" + GetmasterAddress() + "/master/getrule"
	resp, err := Post(url,nil)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	list := Rulelist{}
	err = json.Unmarshal(body,&list)
	if err != nil {
		fmt.Println("err:",err)
		os.Exit(-1)
	}

	for _,v := range list.Data{
		fmt.Printf("%-15s\n",v.Tag)
		for _,v2 := range v.Rulesource {
			fmt.Printf("%-19s%-15s\n","",v2.Name)
			fmt.Printf("%-19s%s%-15s%-s%-s%-s%-s%-s\n","","-- From: ",v2.From,"  ",v2.Alert,"  ",v2.Alertdata,"  ")
			fmt.Printf("%26s%s%s%-v\n\n","-- To: ",v2.AlertTo,"  TIME/S: ",v2.ForTime)
		}
	}
}

var deferdata sync.Map
func GetDeferrule()error{
	url := "http://" + GetmasterAddress() + "/master/showdelayAlert" + "?apiToken=" + Getmasterspid() +"&username=" + GetUser()
	resp, err := Post(url,nil)
	if err != nil {
		fmt.Println("Can't reqdo to ",GetmasterAddress(),"check address and spmid")
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	list := Iplist{}
	err = json.Unmarshal(body,&list)
	if err != nil {
		fmt.Println("err:",err)
		os.Exit(-1)
	}
	for _,v:= range list.Iplists{
		deferdata.Store(v.DelayName,v.DelayData)
	}
	return nil
}

func Getrulerslist(o,ips,group string){

	if GetDeferrule() != nil{
		os.Exit(0)
	}

	fmt.Printf("%-17s%6s\n","IP","Rule")

	url := ""
	if group != "" {
		url = "http://" + GetmasterAddress() + "/master/getrulers" + "?apiToken=" + Getmasterspid() + "&group=" + group +"&username=" + GetUser()
	}else {
		url = "http://" + GetmasterAddress() + "/master/getrulers" + "?apiToken=" + Getmasterspid() +"&username=" + GetUser()
	}

	resp, err := Post(url,nil)
	if err != nil {
		fmt.Println("Can't reqdo to ",GetmasterAddress(),"check address and spmid")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	list := Rulelist{}
	err = json.Unmarshal(body,&list)

	if list.Error != ""{
		fmt.Println(list.Error)
	}

	if err != nil {
		fmt.Println("err:",err)
		os.Exit(-1)
	}

	for _,v := range list.Data{
		if ips != ""{
			if v.Tag != ips{
				continue
			}
		}
		fmt.Printf("%-10s\n",v.Tag)
		for _,v2 := range v.Rulesource {
            NAME := ""
			k,ok := deferdata.Load(v.Tag+":"+v2.Name)
			if ok {
				NAME = v2.Name + "  (暂停告警) Start[" + k.(string) +"]"
			}else {
				NAME = v2.Name
			}

			if o == "sbad"  || o == "sbadr" {
				if v2.Status == "bad" {
					fmt.Printf("%-19s%-20s\n","",NAME)
					fmt.Printf("%-19s%-3s%s\n","","ID:",v2.Md5c)
					fmt.Printf("%-19s%s%-15s%-s%-s%-s%-s%-s\n","","-- From: ",v2.From,"  ",v2.Alert,"  ",v2.Alertdata,"  ")
					fmt.Printf("%-19s%-3s%s\n","","-- Time: ",v2.Time)
					if o == "" {
						fmt.Printf("%26s%s%s%-v\n\n","-- To: ",v2.AlertTo,"  TIME/S: ",v2.ForTime)
					}else {
						fmt.Printf("%26s%s%s%-v\n","-- To: ",v2.AlertTo,"  TIME/S: ",v2.ForTime)
					}



					if o == "r" {
						fmt.Printf("%26s%s\n\n","-- Rs: ",strings.ReplaceAll(v2.Rs,"\n"," "))
					}



					if o == "s" || o == "sbad"{
						fmt.Printf("%26s%s\n\n","-- Ss: ",v2.Status)
					}

					if o == "rs" || o == "sr" || o == "sbadr"{
						fmt.Printf("%26s%s\n","-- Ss: ",v2.Status)
						fmt.Printf("%26s%s\n\n","-- Rs: ",strings.ReplaceAll(v2.Rs,"\n"," "))
					}
				}
			}else if o == "sok" || o == "sokr"{
				if v2.Status == "ok" {
					fmt.Printf("%-19s%-20s\n","",NAME)
					fmt.Printf("%-19s%-3s%s\n","","ID:",v2.Md5c)
					fmt.Printf("%-19s%-3s%s\n","","-- Time: ",v2.Time)
					fmt.Printf("%-19s%s%-15s%-s%-s%-s%-s%-s\n","","-- From: ",v2.From,"  ",v2.Alert,"  ",v2.Alertdata,"  ")
					if o == "" {
						fmt.Printf("%26s%s%s%-v\n\n","-- To: ",v2.AlertTo,"  TIME/S: ",v2.ForTime)
					}else {
						fmt.Printf("%26s%s%s%-v\n","-- To: ",v2.AlertTo,"  TIME/S: ",v2.ForTime)
					}



					if o == "r" {
						fmt.Printf("%26s%s\n\n","-- Rs: ",strings.ReplaceAll(v2.Rs,"\n"," "))
					}
					if o == "s" || o == "sok"{
						fmt.Printf("%26s%s\n\n","-- Ss: ",v2.Status)
					}

					if o == "rs" || o == "sr"||o == "sokr" {
						fmt.Printf("%26s%s\n","-- Ss: ",v2.Status)
						fmt.Printf("%26s%s\n\n","-- Rs: ",strings.ReplaceAll(v2.Rs,"\n"," "))
					}
				}
			}else {
				fmt.Printf("%-19s%-20s\n","",NAME)
				fmt.Printf("%-19s%-3s%s\n","","ID:",v2.Md5c)
				fmt.Printf("%-19s%-3s%s\n","","-- Time: ",v2.Time)
				fmt.Printf("%-19s%s%-15s%-s%-s%-s%-s%-s\n","","-- From: ",v2.From,"  ",v2.Alert,"  ",v2.Alertdata,"  ")
				if  o != "s" && o != "r"&& o != "sr"&& o != "rs"&& o != "sok"&& o != "sbad"&& o != "sokr"&& o != "sbadr"{
					fmt.Printf("%26s%s%s%-v\n\n","-- To: ",v2.AlertTo,"  TIME/S: ",v2.ForTime)
				}else {
					fmt.Printf("%26s%s%s%-v\n","-- To: ",v2.AlertTo,"  TIME/S: ",v2.ForTime)
				}

				if o == "r" {
					fmt.Printf("%26s%s\n\n","-- Rs: ",strings.ReplaceAll(v2.Rs,"\n"," "))
				}

				if o == "s"{
					fmt.Printf("%26s%s\n\n","-- Ss: ",v2.Status)
				}

				if o == "rs" || o == "sr" {
					fmt.Printf("%26s%s\n","-- Ss: ",v2.Status)
					fmt.Printf("%26s%s\n\n","-- Rs: ",strings.ReplaceAll(v2.Rs,"\n"," "))
				}

			}




		}
	}
}

func Getmethod(){
	fmt.Printf("%-30s%5s\n","Name","Mode")
	url := "http://" + GetmasterAddress() + "/master/showmethod" + "?apiToken=" + Getmasterspid() +"&username=" + GetUser()
	resp, err := Post(url,nil)
	if err != nil {
		fmt.Println("Can't reqdo to ",GetmasterAddress(),"check address and spmid")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	list := loadconf.ShareConfload.AlertMethod
	err = json.Unmarshal(body,&list)
	if err != nil {
		fmt.Println("err:","failed with unkonw error")
		os.Exit(-1)
	}
	for _,v := range list{
		fmt.Printf("%-30s%-6s\n",v.Path,v.RunMode)
	}
}

func Syncstatus(){
	fmt.Printf("%-30s%5s\n","Sync AgentStatus ","Running...")
	url := "http://" + GetmasterAddress() + "/master/syncstatus" + "?apiToken=" + Getmasterspid()  +"&username=" + GetUser()
	resp, err := Post(url,nil)
	if err != nil {
		fmt.Println("Can't reqdo to ",GetmasterAddress(),"check address and spmid")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	list := Mastejson{}
	err = json.Unmarshal(body,&list)
	if err != nil {
		fmt.Println("err:",err)
		os.Exit(-1)
	}

	if list.Error != ""{
		fmt.Println(list.Error)
		os.Exit(-1)
	}
	if list.Master == "ok" {
		fmt.Println("\nAll agent Status are Sync OK!   -get ip to see")
	}else {
		fmt.Println("error",list.Master)
	}
}


func Sleepalert(id,time string){
	fmt.Printf("%-s%5s\n","- "," Marshal from pool...")
	url := "http://" + GetmasterAddress() + "/master/delatAlertwithMd5" + "?apiToken=" + Getmasterspid()  +"&username=" + GetUser()
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("id", id)
	_ = writer.WriteField("timex", time)
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
		fmt.Println("Can't reqdo to ",GetmasterAddress(),"check address and spmid")
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	list := Mastejson{}
	err = json.Unmarshal(body,&list)
	if err != nil {
		fmt.Println("err:",err)
		os.Exit(-1)
	}
	if list.Error != ""{
		fmt.Println(list.Error)
		os.Exit(-1)
	}
	if list.Master == "ok" {
		fmt.Println(id,"Sleep Alert ",time,"s  ok\n" )
	}else {
		fmt.Println("RunPool Error:",list.Master)
	}
}

func Openalert(id string){
	fmt.Printf("%-s%5s\n","- "," Marshal from pool...")
	url := "http://" + GetmasterAddress() + "/master/openalertWithMd5" + "?apiToken=" + Getmasterspid() +"&username=" + GetUser()
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("id", id)
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
		fmt.Println("Can't reqdo to ",GetmasterAddress(),"check address and spmid")
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	list := Mastejson{}
	err = json.Unmarshal(body,&list)
	if err != nil {
		fmt.Println("err:",err)
		os.Exit(-1)
	}
	if list.Error != ""{
		fmt.Println(list.Error)
		os.Exit(-1)
	}
	if list.Master == "ok" {
		fmt.Println(id,"Open Alert "," ok\n" )
	}else {
		fmt.Println("RunPool Error:",list.Master)
	}
}













func Post(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}

func Get(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("Get", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}
