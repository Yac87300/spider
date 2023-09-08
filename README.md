#spider  
自动化运维/数据采集/Api转发/活接口调用  


编译  
go build -o spider main.go  

Master 编辑配置文件  
./spider   


agent编辑配置文件  
./spider --agent   


----------------------------------------------
编译日志组件  
cd  mysql  
go build spidersql main.go  

启动日志组件 (需要自建mysql)  <- master需要修改配置文件适配日志组件  
user passwd host Port  
./spidersql -u -p -h  -P  



自建Grafana 导入大盘，修改mysql地址  


---------------------------------------------------------
命令行工具 
cd cli;go build spcli main.go  
spcli -g login -u  #登录 授权 获取之后复制token到本地环境  
spcli -g rule  #查看agent正在运行的规则 以及状态 （-o sr/s/r/sbad/sok）  
spcli -g ip    # 查看agent在线状态   -group 分组  
spcli -g sync #同步集群  
spcli -g leg -l legname -o leg中自定义的变量(逗号隔开) 例如： spcli -g leg -l autodf -o path=/data,logfile=/root/log  
spcli -g alert #查看master的告警规则  
spcli -g catproxy #查看api代理  
spcli -g catleg  #查看自定义脚本  
spcli -g sleep  #休眠告警 -id ruleid -time 休眠时间  
spcli -g open  #打开告警 -id ruleid  
spcli -g edit -ip ip #远程修改agent配置文件

---------------------------------------------------------
V2版本工具  
spcli -g port -s 0.0.0.0:80 -d 192.168.23.198:73  
spcli -g ai -e "如何排查k8s node notready的问题"    
spcli -g doit -msg "巡检表"                                                                                                                                                                                                     
spcli -g stroage -show info  查看存储的容量、状态、kv的节点分布  
spcli -g stroage -c 2all    所有存储机器保存全量数据            #更改会暂停集群服务 等待数据同  步
spcli -g storage -c 2slice  修改存储规则，1份全量，2份切片    #只会对后面增量数据生效，之前的全量数据不会再切片  
spcli -g storage -bak /mnt/xxxx.bak  备份  



V2 大版本  （24 Q1 发布）
---------------------------------------------------------
新增Dashboard                前端管理页面  
新增PortMan                  一键转发端口流量Port2Port，不依赖防火墙  
新增X2S                      自定义Mod、function、接口并通过网关管控、转发、记录  
新增高级Rule                  支持更多的rule规则，引入内置的正则表达式，全新的SpiderRule语法,不单单是执行命令  
新增WeArefamily              只用配置一个Master即可实现高可用，master失效时，node可以自动选举，并转为master，等待主master恢复  
新增Aiops                    引入训练好的ChatGpt作为运维助手，协助排查  
新增boatboxStorage           内置存储引擎，不再依赖mysql、etcd，存储支持容灾  
新增JustDoit                 支持生成巡检列表  











