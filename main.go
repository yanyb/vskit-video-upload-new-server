package main

import (
	"flag"
	"fmt"
	"go-app/app"
	"os"
	"path"
	"runtime"
)

var (
	profile = flag.String("profile", "dev", "profile: dev, test, prod ")
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	if os.Getenv("vskitGoEnv") != "" {
		*profile = os.Getenv("vskitGoEnv")
	}

	fmt.Println("启动的profile: ", *profile)
	configPath := path.Join("config", "config-"+*profile+".yml")

	// 初始化配置文件
	app.InitApp(configPath, *profile)

	// 启动
	app.Start()
}
