package main

import (
	"encoding/json"
	"example.com/mod/cli/pool"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"log"
	"os"
)

func main(){
	get := flag.String("g", "rule", "-g ip/ipok/ipbad/rule/sync/alert")
	o := flag.String("o","s","r,s,rs")
	e := flag.String("e","","-e key=xxx,key=xxx")
	u := flag.String("u","","username")
	l := flag.String("l","","legname")
	ip := flag.String("ip","","ip to get")
	id := flag.String("id","","id from rule")
	times := flag.String("time","600","sleep time/s")
	Group := flag.String("group","","group of ip")
	//alert := flag.String("alert", "", "defer/open")
	//timex := flag.Int("time",60,"sleep time")
	if pool.GetmasterAddress() == ""{
		fmt.Println("no masterAddress,please set MasterAddress Frist,set ENV(spmadd) for systemctl")
		os.Exit(-1)
	}

	//if pool.Getmasterspid() == ""{
	//	fmt.Println("no spid,please set spid Frist,set ENV(spid) for systemctl")
	//	os.Exit(-1)
	//}

	flag.Parse()

	switch *get{
	case "ip":
		Getiplist("ip",*Group)
	case "ipbad":
		 Getiplist("ipbad",*Group)
	case "ipok":
		Getiplist("ipok",*Group)
	case "rule":
		pool.Getrulerslist(*o,*ip,*Group)
	case "alert":
		pool.Getmethod()
	case "sync":
		pool.Syncstatus()
	case "sleep":
		if *id != "" {
			pool.Sleepalert(*id,*times)
		}else {
			fmt.Println("-id can't be empty")
			return
		}
	case "open":
		if *id != "" {
			pool.Openalert(*id)
		}else {
			fmt.Println("-id can't be empty")
			return
		}
	case "leg":
		pool.Doleg(*l,*e)
	case "cfile":
		pool.Ws2(*e)
	case "dfile":
		pool.Download(*e)
	case "catproxy":
		pool.Getproxy()

	case "catleg":
		pool.Getleg()
	case "login":
		fmt.Print("Enter your password: ")
		bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println("Error reading password:", err)
			return
		}
		fmt.Print("\n")
		pool.SetLogin(string(bytePassword),*u)
	default:
		fmt.Println("-g ip/ipok/ipbad/rule/sync/alert")
	}
	flag.Parse()
}



func Getiplist(typs,group string){
	fmt.Printf("%-17s%-14s%-14s\n","IP","Status","Group")
	url := ""
	if group == "" {
		url = "http://" + pool.GetmasterAddress() + "/master/show" + "?apiToken=" + pool.Getmasterspid() +"&username=" + pool.GetUser()
	}else {
		url = "http://" + pool.GetmasterAddress() + "/master/show" + "?apiToken=" + pool.Getmasterspid() + "&group=" + group +"&username=" + pool.GetUser()
	}

	resp, err := pool.Post(url,nil)
	if err != nil {
		fmt.Println("Can't reqdo to ",pool.GetmasterAddress(),"check address and spmid")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	list := pool.Iplist{}
	err = json.Unmarshal(body,&list)

	if list.Error != ""{
		fmt.Println("err:",list.Error)
		os.Exit(-1)
	}

	if err != nil {
		fmt.Println("err:",err)
		os.Exit(-1)
	}


		for _,v := range list.Iplists{
			if typs == "ipbad" {
				if v.Status == "offonline" {
					//fmt.Printf("%-17s%-10s\n",v.Ip,v.Status)
					fmt.Printf("%-17s%-14s%-14s\n",v.Ip,v.Status,v.Group)
				}
			}else if typs== "ipok"{
				if v.Status == "online" {
					//fmt.Printf("%-17s%-10s\n",v.Ip,v.Status)
					fmt.Printf("%-17s%-14s%-14s\n",v.Ip,v.Status,v.Group)
				}
			}else {
				fmt.Printf("%-17s%-14s%-14s\n",v.Ip,v.Status,v.Group)
			}

		}


}


