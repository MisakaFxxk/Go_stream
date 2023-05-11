# Emby播放分离

### 1、前期准备

至少两台服务器（否则你这分离似乎没有什么意义）、由Nginx反向代理的Emby服务端、一定的动手能力。

### 2、原理实现

为了更好地讲解，我们将Emby提供用户交互的部分称之为**前端**，负责播放的部分称之为**后端**。

通过Nginx正则匹配，拦截原Emby的播放请求`/videos/\d*/stream`至本地搭建的服务接口（代码结构中为stream_front的部分），在stream_front部分中，我们首先提取了URL内的各项参数，然后通过`api_key`这一参数的是否存在来判断否是使用了Infuse客户端（万恶之源）。

对于Infuse客户端而言，其请求的播放链接中仅包含了`MediaSourceId`与没用的`static`参数，这样我们无法通过用户的`api_key`来请求Emby服务端获取播放文件真实地址，需要使用Emby后台生成的`Admin_api_key`。

对于非Infuse客户端，我们通过拼接`%s/Items/%s/PlaybackInfo?MediaSourceId=%s&api_key=%s`来向Emby服务端获取播放文件的真实地址。例如我是通过docker运行的Emby服务端，创建docker时使用了`--volume /mnt:/mnt/share`这样一个映射，获取到的docker内文件路径为：**/mnt/share/local1/google2/动漫追更/总之就是非常可爱/Season 02/S02E05 - 哔哩哔哩东南亚.2160P-MisakaF.mkv**，对应的物理机文件路径为：**/mnt/local1/google2/动漫追更/总之就是非常可爱/Season 02/S02E05 - 哔哩哔哩东南亚.2160P-MisakaF.mkv**。

接着进行鉴权，在此之前我们需要手动定义一个key，例如VRX1BVYHLEGP5ERMU2GC，程序会将dir，MediaSourceId，remote_token(自己定义的KEY)进行拼接，然后计算MD5，附加在请求播放后端时的URL内。

------

进入播放后端部分，对应代码结构中的stream_backend部分。程序接收了前端302重定向而来的URL后，将dir，MediaSourceId，remote_token(事先约定好的KEY)进行拼接，然后计算MD5，与参数中携带的MD5值进行比对，若相同则通过鉴权，返回本地的视频文件给用户客户端，开始播放。

### 3、着手搭建

#### 3.1 Nginx

进入反向代理，在`location / `的大括号内添加

```
if ($request_uri ~* /videos/\d*/stream)
    {
    proxy_pass http://127.0.0.1:60001;
    }
```

#### 3.2 前端

##### 3.2.1 Golang版本

**此版本苹果系统下的Emby官方客户端无法正常播放，疑似与请求头缺失有关，但抓包分析后无法修复。**

下载最新release，将其中的stream_front文件夹上传到服务器内，修改config.yaml文件

```
# Emby配置
Emby:
  url: "http://xxx:xxx/"
  apikey: "xxx"

# 播放端配置
Remote:
  url: "http://xxx:12180/stream"
  apikey: "xxx"

# 目录头配置
Local: 
  dir: "xxx"
```

Emby配置不必多说，播放端配置就是你的后端地址和你自己生成的一个apikey。

目录头配置很重要！首先你要确保你的所有视频文件有一个公共的上层（除根目录），比如在我的服务器内，我挂载了三个文件夹，分别在

```
/mnt/google
/mnt/local1
/mnt/local2
```

很明显他们都有公共的上层目录`/mnt`，此时如果你是非docker搭建的Emby，目录头配置中的dir填写`/mnt`即可。如果你是docker搭建的emby，使用了目录映射，比如我将外部目录/mnt映射到docker内部目录/mnt/share，那么目录头配置中的dir填写`/mnt/share`。

**总之你需要保证Emby内的所有视频路径都可以减去这个目录头。**

然后cd到stream_front文件夹内，执行下面命令开启程序：

```
chmod +x ./StreamFront
nohup ./StreamFront ./config.yaml  > StreamFront.log 2>&1 &
```

##### 3.2.2 Python版本

**此版本所有客户端均可正常使用**

下载最新release，将其中的stream_front/main_os.py上传到服务器内，打开文件修改5-9行。具体每个配置该怎么填请参考上方Golang部分。

运行程序：

```
nohup python3 /root/stream_front/main.py  > streampy.log 2>&1 &
```

------

#### 3.3后端

和前端相同目录结构挂载所有文件夹。

下载最新release，将其中的stream_backend文件夹上传到服务器内，修改config.yaml文件

```
# 播放端配置
Remote:
  apikey: "xxx"

# 目录头配置
Mount: 
  dir: "xxx"
```

Remote.apikey填写你在前端设置的一串秘钥。这里的目录头配置同样重要，非docker用户直接和前端的config里填一样的即可。

docker用户填`/mnt`，我来举个例子帮助你理解：

```
在emby内看到的文件路径：/mnt/share/local1/google2/动漫追更/总之就是非常可爱/Season 02/S02E05 - 哔哩哔哩东南亚.2160P-MisakaF.mkv
前端填写的目录头：/mnt/share
经过前端处理，传输到后端的路径：/local1/google2/动漫追更/总之就是非常可爱/Season 02/S02E05 - 哔哩哔哩东南亚.2160P-MisakaF.mkv
此时，我们需要这个影片在后端服务器内的真实路径，由于我们是相同目录挂载，所以他的真实路径应该为：/mnt/local1/google2/动漫追更/总之就是非常可爱/Season 02/S02E05 - 哔哩哔哩东南亚.2160P-MisakaF.mkv
相比前端传过来的路径缺少了/mnt部分，所以我们需要在后端的配置文件内将目录头设置为/mnt来拼接获得真实路径。
```

然后cd到stream_backend文件夹内，执行下面命令开启程序：

```
chmod +x ./StreamBackend
nohup ./StreamBackend ./config.yaml  > StreamBackend.log 2>&1 &
```

### 4、端口占用

前端与Emby交互使用60001端口，前端与后端交互使用12180端口，请确保后端服务器12180保持开放，目前无法通过配置文件更改端口。更改端口需要下载源代码修改后重新编译。
