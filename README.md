# spider
自动化运维/数据采集/Api转发/活接口调用


#编译
go build -o spider main.go

#Master 编辑配置文件
./spider 


#agent编辑配置文件
./spider --agent 

#编译日志组件
cd  mysql
go build spidersql main.go

# 启动日志组件 (需要自建mysql)  <- master需要修改配置文件适配日志组件
user passwd host Port
./spidersql -u -p -h  -P

#自建Grafana 导入大盘，修改mysql地址



#命令行工具 
cd cli;go build spcli main.go


spcli -g login -u  #登录 授权 获取之后复制token到本地环境

spcli -g rule  #查看agent正在运行的规则 以及状态 （-o sr/s/r/sbad/sok）
spcli -g ip    # 查看agent在线状态   -group 分组
spcli -g sync #同步
spcli -g leg -l legname -o leg中自定义的变量(逗号隔开) 例如： spcli -g leg -l autodf -o path=/data,logfile=/root/log
spcli -g alert #查看master的告警规则
spcli -g catproxy #查看api代理
spcli -g catleg  #查看自定义脚本
spcli -g sleep  #休眠告警 -id ruleid -time 休眠时间
spcli -g open  #打开告警 -id ruleid
spcli -g edit -ip ip #远程修改agent配置文件



