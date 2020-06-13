# 测试使用
## 测试文件下载
下载界面
https://tianchi.aliyun.com/competition/entrance/231790/information

下载
```bash
wget https://tianchi-competition.oss-cn-hangzhou.aliyuncs.com/231790/trace1_data.tar.gz?spm=5176.12281978.0.0.6ae21022DkMU0X&OSSAccessKeyId=LTAI4G7mrxYb7QrcXkTr3zzt&Expires=1593247540&Signature=YZLVJiJiJOga/KY4Z8dKVPyGJAQ=
wget https://tianchi-competition.oss-cn-hangzhou.aliyuncs.com/231790/trace2_data.tar.gz?spm=5176.12281978.0.0.6ae21022DkMU0X&OSSAccessKeyId=LTAI4G7mrxYb7QrcXkTr3zzt&Expires=1593247613&Signature=Z7PZrVG1UXnXdbO08bPv3hBCfyA=
```

解压
```bash
tar -xvf trace1_data.tar.gz
tar -xvf trace2_data.tar.gz
rm trace1_data.tar.gz
rm trace2_data.tar.gz
```

## 临时文件服务器(非官方流程)
配置 nginx 文具
file_server.conf
```
server {
    listen  9971;
    server_name  localhost;
    charset utf-8;
    # set your path
    root /Users/arcosx/tianchi/WesternQueen/test; 
    location / {
        autoindex on;
        autoindex_exact_size on;
        autoindex_localtime on;
    }
}
```


## 官方流程测试
详见 https://tianchi.aliyun.com/competition/entrance/231790/information






