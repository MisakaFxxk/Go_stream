import flask,requests,hashlib
from flask import request, redirect
from flask_cors import CORS
import urllib.parse

emby_url = ""
emby_key = ""
local_dir = ""  # 目录头
remote_api = "http://ip:12180/stream"
remote_token = ""

app = flask.Flask(__name__)

@app.route('/emby/videos/<item_Id>/stream.<type>',methods=["GET"])
def handle_request(item_Id,type):
    MediaSourceId = request.args.get("MediaSourceId")
    api_key = request.args.get("api_key")

    if api_key:
        # 非Infuse
        itemInfoUri = f"{emby_url}/Items/{item_Id}/PlaybackInfo?MediaSourceId={MediaSourceId}&api_key={api_key}"
        print(itemInfoUri)
        emby_path = fetchEmbyFilePath(itemInfoUri,MediaSourceId)

        local = str(emby_path).replace(local_dir,"")
        local = urllib.parse.quote(local) 
        raw_string = "dir=" + str(local) + "&MediaSourceId=" + str(MediaSourceId) + "&remote_token=" + str(remote_token)
        md5_verify = hashlib.md5((raw_string).encode(encoding='UTF-8')).hexdigest()
        raw_url = remote_api + "?dir=" + str(local) + "&MediaSourceId=" + str(MediaSourceId) + "&key=" + str(md5_verify)

        return redirect(raw_url)
    else:
        # Infuse
        itemInfoUri = f"{emby_url}/Items/{item_Id}/PlaybackInfo?MediaSourceId={MediaSourceId}&api_key={emby_key}"
        emby_path = fetchEmbyFilePath(itemInfoUri,MediaSourceId)

        local = str(emby_path).replace(local_dir,"")
        local = urllib.parse.quote(local) 
        raw_string = "dir=" + str(local) + "&MediaSourceId=" + str(MediaSourceId) + "&remote_token=" + str(remote_token)
        md5_verify = hashlib.md5((raw_string).encode(encoding='UTF-8')).hexdigest()
        raw_url = remote_api + "?dir=" + str(local) + "&MediaSourceId=" + str(MediaSourceId) + "&key=" + str(md5_verify)

        return redirect(raw_url)

@app.route('/Videos/<item_Id>/stream',methods=["GET"])
def handle_request2(item_Id):
    MediaSourceId = request.args.get("MediaSourceId")
    api_key = request.args.get("api_key")

    if api_key:
        # 非Infuse
        itemInfoUri = f"{emby_url}/Items/{item_Id}/PlaybackInfo?MediaSourceId={MediaSourceId}&api_key={api_key}"
        emby_path = fetchEmbyFilePath(itemInfoUri,MediaSourceId)

        local = str(emby_path).replace(local_dir,"")
        local = urllib.parse.quote(local) 
        raw_string = "dir=" + str(local) + "&MediaSourceId=" + str(MediaSourceId) + "&remote_token=" + str(remote_token)
        md5_verify = hashlib.md5((raw_string).encode(encoding='UTF-8')).hexdigest()
        raw_url = remote_api + "?dir=" + str(local) + "&MediaSourceId=" + str(MediaSourceId) + "&key=" + str(md5_verify)

        return redirect(raw_url)
    else:
        # Infuse
        itemInfoUri = f"{emby_url}/Items/{item_Id}/PlaybackInfo?MediaSourceId={MediaSourceId}&api_key={emby_key}"
        emby_path = fetchEmbyFilePath(itemInfoUri,MediaSourceId)

        local = str(emby_path).replace(local_dir,"")
        local = urllib.parse.quote(local) 
        raw_string = "dir=" + str(local) + "&MediaSourceId=" + str(MediaSourceId) + "&remote_token=" + str(remote_token)
        md5_verify = hashlib.md5((raw_string).encode(encoding='UTF-8')).hexdigest()
        raw_url = remote_api + "?dir=" + str(local) + "&MediaSourceId=" + str(MediaSourceId) + "&key=" + str(md5_verify)

        return redirect(raw_url)



def fetchEmbyFilePath(itemInfoUri,MediaSourceId):
    # 获取Emby内文件路径
    req = requests.get(itemInfoUri)
    resjson = req.json()
    for i in resjson['MediaSources']:
        if i['Id'] == MediaSourceId:
            mount_path = i['Path']
    return mount_path


# 在Flask应用中启用CORS
CORS(app, resources={r"/*": {"origins": "*"}})

if __name__ == '__main__':
    app.run(port=60001,debug=True,host='0.0.0.0',threaded=True) 