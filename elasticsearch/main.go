package main

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
)

var Client *elastic.Client
var host = "http://192.168.0.88:9200"

//初始化
func init() {
	//errorlog := log.New(os.Stdout, "APP", log.LstdFlags)
	var err error
	//这个地方有个小坑 不加上elastic.SetSniff(false) 会连接不上
	Client, err = elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(host))
	if err != nil {
		panic(err)
	}
	_,_,err = Client.Ping(host).Do(context.Background())
	if err != nil {
		panic(err)
	}
	//fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	esversion,err := Client.ElasticsearchVersion(host)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch version %s\n", esversion)
}

func main() {

	if !ExistsIndex(){
		CreateIndex()
	}
	//ExistsIndex()
	//CloseIndex()
	//DelIndex()
	OpenIndex()
	fmt.Println("处理完成。")
}

/*
创建索引
*/
func CreateIndex()  {
	mapping := `{
   "mappings":{
   "properties":{
   "taskid":     { "type": "long" },
   "taskname":   { "type": "text" },
   "taskresult":   { "type": "text" },
   "taskstatus":   { "type": "text" },
   "datafrom":    { "type": "text" },
   "tasktime":  { "type": "text" },
   "alertmethod":         { "type": "text" },
   "alertData":  { "type": "text" },
   "alerttype":  { "type": "text" }
   }
  }
 }`
	createIndex,err := Client.CreateIndex("spider").BodyString(mapping).Do(context.Background())

	if err != nil{
		fmt.Println(err)
	}else{
		if !createIndex.Acknowledged{
			fmt.Println("创建失败")
		}else {
			fmt.Println("创建成功")
		}
	}
	AddAliasses()
}

/*
删除索引
*/
func DelIndex()  {
	Client.DeleteIndex("spider").Do(context.Background())
}

/*
判断索引是否存在
*/
func ExistsIndex()bool  {
	exist,err:=Client.IndexExists("spider").Do(context.Background())
	if err!= nil{

	}
	if !exist{
		fmt.Println("不存在")
		return false
	}else {
		fmt.Println("存在")
		return true
	}
	return true
}

/*
索引添加别名
*/
func AddAliasses()  {
	Client.Alias().Add("spider","ses").Do(context.Background())
}

/*
关闭索引
*/
func CloseIndex()  {
	Client.CloseIndex("enIndex_20211129").Do(context.Background())
}

/*
打开索引
*/
func OpenIndex()  {
	Client.OpenIndex("enIndex_20211129").Do(context.Background())
}

