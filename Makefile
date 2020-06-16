build:
	go build -o WesternQueen
docker:
	docker login -u a2osdocker@1443039390876007 -p a2osdocker registry.cn-hangzhou.aliyuncs.com
	docker build -t "registry.cn-hangzhou.aliyuncs.com/a2os/tianchi:1.0" .
	docker push registry.cn-hangzhou.aliyuncs.com/a2os/tianchi:1.0



