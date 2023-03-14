url=$1
msg=$2
curl "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=$url" \
   -H 'Content-Type: application/json' \
   -d '
   {
    	"msgtype": "text",
    	"text": {
        	"content": "'"$msg"'"
    	}
   }'
