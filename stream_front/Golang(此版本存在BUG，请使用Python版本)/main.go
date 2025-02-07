package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
	"github.com/spf13/viper"
)

var configFile string // 配置文件路径

var Config struct {
	Emby struct {
		url    string
		apikey string
	}
	Remote struct {
		url    string
		apikey string
	}
	Local struct {
		dir string
	}
}

// 获取URL请求
func stream(c *gin.Context) {
	// 从URL中获取参数
	MediaSourceId := c.Query("MediaSourceId")
	api_key := c.Query("api_key")

	if len(api_key) != 0 {
		itemId := strings.Split(c.Param("path"), "/")[1]

		// 获取Emby内文件路径
		var itemInfoUri string = fmt.Sprintf(`%s/Items/%s/PlaybackInfo?MediaSourceId=%s&api_key=%s`, Config.Emby.url, itemId, MediaSourceId, api_key)
		emby_path := fetchEmbyFilePath(itemInfoUri)

		// 拼接鉴权key
		local := strings.Replace(emby_path, Config.Local.dir, "", -1)
		raw_string := "dir=" + local + "&MediaSourceId=" + MediaSourceId + "&remote_token=" + Config.Remote.apikey
		hash_1 := md5.Sum([]byte(raw_string))
		hash := hex.EncodeToString(hash_1[:])
		raw_url := Config.Remote.url + "?dir=" + local + "&MediaSourceId=" + MediaSourceId + "&key=" + hash
		// 302跳转
		c.Header("Content-Type", "video/mp4")
		c.Header("Content-Disposition", "inline")
		c.Redirect(302, raw_url)

	} else {
		// Infuse请求
		itemId := strings.Split(c.Param("path"), "/")[1]

		// 获取Emby内文件路径
		var itemInfoUri string = fmt.Sprintf(`%s/Items/%s/PlaybackInfo?MediaSourceId=%s&api_key=%s`, Config.Emby.url, itemId, MediaSourceId, Config.Emby.apikey)
		emby_path := fetchEmbyFilePath(itemInfoUri)

		// 拼接鉴权key
		local := strings.Replace(emby_path, Config.Local.dir, "", -1)
		raw_string := "dir=" + local + "&MediaSourceId=" + MediaSourceId + "&remote_token=" + Config.Remote.apikey
		hash_1 := md5.Sum([]byte(raw_string))
		hash := hex.EncodeToString(hash_1[:])
		raw_url := Config.Remote.url + "?dir=" + local + "&MediaSourceId=" + MediaSourceId + "&key=" + hash
		// 302跳转
		c.Header("Content-Type", "video/mp4")
		c.Header("Content-Disposition", "inline")
		c.Redirect(302, raw_url)

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

// 获取Emby内文件信息
func fetchEmbyFilePath(itemInfoUri string) string {
	req := gorequest.New()
	_, body, err := req.Post(itemInfoUri).End()
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	var resjson map[string]interface{}
	err1 := json.Unmarshal([]byte(body), &resjson)
	if err1 != nil {
		panic(err1)
	}
	mount_path := resjson["MediaSources"].([]interface{})[0].(map[string]interface{})["Path"].(string)

	return mount_path
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
	Config.Emby.url = viper.GetString("Emby.url")
	Config.Emby.apikey = viper.GetString("Emby.apikey")
	Config.Remote.url = viper.GetString("Remote.url")
	Config.Remote.apikey = viper.GetString("Remote.apikey")
	Config.Local.dir = viper.GetString("Local.dir")

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(corsMiddleware())
	r.GET("/emby/videos/*path", stream)
	r.GET("/Videos/*path", stream)
	r.Run(":60001")

}
