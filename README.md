# ShortUrlApi
使用Redis作为后端存储。  
返回json格式数据。  

-  创建短链接接口：POST  http://127.0.0.1:8000/api/shorten  

```
{
	"url": "http://www.baidu.com",
	"expiration_in_minutes": 10
}
```  

Response:  
```
{
    "shortlink": "4"
}
```
- 短链接详细信息: GET  http://127.0.0.1:8000/api/info?shortlink=1  

Response:  
```
"{\"url\":\"https://www.baidu.com\",\"created_at\":\"2020-02-21 17:30:40.8550107 +0800 CST m=+78.665463301\",\"expiration_in_minutes\":10}"
```

- 短链接跳转真实地址: GET  http://127.0.0.1:8000/4 




