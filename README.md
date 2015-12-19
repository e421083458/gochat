# golang聊天室实例

本实例基于websocket和jQuery开发。
[websocket](https://github.com/gorilla/websocket)
[jQuery](http://jquery.com) 

本实例特点如下：
1. 支持浏览器客户端和命令行客户端两种方式。
2. 支持私聊。

## 运行实例

实例运行运行在go环境中，安装go环境请参照(http://golang.org/doc/install)

启动服务器。

    $ git clone github.com/gorilla/websocket
    $ go get github.com/gorilla/websocket
    $ cd server
    $ go run *.go

运行命令行客户端   

    $ cd client
    $ go run *.go

运行浏览器客户端
    http://127.0.0.1:8080/

登陆聊天室输入：

    username=niuyufu&token=123456

跟某人私聊输入：

    @dengyanlu 这是私聊信息！
