package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var remote_token, configFile, mount_dir string

func remote(c *gin.Context) {
	MediaSourceId := c.Query("MediaSourceId")
	dir := c.Query("dir")
	key := c.Query("key")

	// 鉴权
	raw_string := "dir=" + dir + "&MediaSourceId=" + MediaSourceId + "&remote_token=" + remote_token
	hash_1 := md5.Sum([]byte(raw_string))
	hash := hex.EncodeToString(hash_1[:])
	if key == hash {
		// 鉴权成功
		local_dir := mount_dir + dir
		c.File(local_dir)
	} else {
		// 鉴权失败
		c.AbortWithStatus(403)
	}
}

// 定义中间件处理跨域请求
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	// 读取配置文件
	args := os.Args[1:]
	configFile = args[0]
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("read config file error:", err)
		return
	}
	remote_token = viper.GetString("Remote.apikey")
	mount_dir = viper.GetString("Mount.dir")
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(corsMiddleware())
	r.GET("/stream", remote)
	r.Run(":12180")
}
