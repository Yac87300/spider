package pool

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"example.com/mod/master"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

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

func GetmasterAddress()string{
	add := os.Getenv("spmadd")
	return add
}

func Getmasterspid()string{
	add := os.Getenv("apiToken")
	return add
}

func GetUser()string{
	add := os.Getenv("username")
	return add
}

func Setenv2(cmds string){
	fmt.Println(cmds)
	cmd := exec.Command("bash", "-c", "source ~/.sl && exec $SHELL")
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

func Setenv(cmds string)string{
	fmt.Println(cmds)
	cmd := exec.Command("bash","-c",cmds)
	//创建获取命令输出管道
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		fmt.Printf("Error:can not obtain stdout pipe for command:%s\n", err)
		return ""
	}

	//执行命令
	if err := cmd.Start(); err != nil {
		fmt.Println("Error:The command is err,", err)
		return ""
	}

	//读取所有输出
	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Println("ReadAll Stdout:", err.Error())
		return ""
	}

	err = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	cmd.Process.Signal(syscall.SIGKILL)
	err = cmd.Wait()

	 return string(bytes)

}

func SetLogin(passwd,username string){

	url := "http://" + GetmasterAddress() + "/master/auth"
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("passwd", passwd)
	_ = writer.WriteField("username", username)
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
	x := authjson{}
	err = json.Unmarshal(body,&x)

	if x.Msg != "认证成功"{
        fmt.Println("认证失败,请重新登录")
		return
	}
	token := strconv.Itoa(int(x.ApiToken))
    //err =	os.Setenv("apiToken",token)
    // fmt.Println(err)
    //err =	os.Setenv("username",username)
    fmt.Println("")
    master.ALertRun("export apiToken=" + token)
	master.ALertRun("export username=" + username)
	//Setenv("echo " + "export \"apiToken=\"" + token + "> ~/.sl")
	//Setenv("echo " + "export \"username=\"" + username + ">> ~/.sl")
	//paths := strings.ReplaceAll(Setenv("echo ~"),"\n","")
	//Setenv2(  paths + "/.sl")
	fmt.Println("Please Copy ^^^ Env in Your Bash to LoginAuth.")
}

type authjson struct {
	Authcode int    `json:"Authcode"`
	ApiToken int64  `json:"apiToken"`
	Msg      string `json:"msg"`
}


func calculateMD5(input string) string {

	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}
