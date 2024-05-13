# web-mysql


## 目录结构：

```
.  
├── README.md  
├── go.mod  
├── go.sum  
├── docker-compose.yaml  # compose文件，通过在当前目录执行 `docker-compose up` 可以运行代码
├── Dockerfile           # 用于镜像编译 
├── Makefile             # Makefile，定义了 run、build、test 等命令   
├── bin  
│   └── server  # 项目编译后的二进制文件
├── cmd  
│   └── server  
│       └── main.go  
├── config  # 配置目录  
│   ├── default.properties  # 默认配置  
│   ├── local.properties    # 本地开发配置，不设置环境变量，会使用该配置
│   ├── dev.properties      # 开发环境配置，可以用于部署到开发环境的服务器上
│   └── prod.properties     # 生产环境配置，用于部署到生产环境使用
├── internal  
│   ├── master.go           # 主 Preist 函数，用于加载框架基础依赖(本项目只依赖了 goner/gin ) 和 内部 Priest 函数
│   ├── priest.go           # 运行 make gone 生成的文件，包括一个Priest函数，用于安葬(加载)所有内部的Goner组件
│   ├── interface           # 接口包，可以包含 domain 、event、 service 等子包 
│   │   ├── domain          # 用于定义业务对象(实体对象、值对象、聚合根等)
│   │   │   ├── demo.go  
│   │   │   └── user.go  
│   │   └── service          # 内部接口包，用于管理内部接口定义
│   │       └── i_demo.go    # IDemo 接口
│   ├── middleware                 # 中间件包，用于定义中间件
│   │   ├── authorize.go  
│   │   └── pub.go  
│   ├── module                     # 模块
│   │   └── demo             # demo模块，一般情况下，一个模块与service包下的一个文件对应，实现其中定义的接口
│   │       ├── demo_svc.go  # 实现 IDemo 接口
│   │       └── error.go     # 模块内的错误编码定义
│   ├── controller                 # controller目录，用于定义http服务接口 
│   │   └── demo_ctr.go      # demo controller
│   ├── pkg  
│   │   └── utils  
│   │       └── error.go     # 定义全局错误编码
│   └── router  
│       ├── auth_router.go         # 定义需要鉴权的路由
│       └── pub_router.go          # 定义不需要鉴权的路由
└── tests                                # 全局测试目录
    └── api                              # 接口测试
        └── demo.http                    # demo接口测试
``` 

## 命令

> 假设已经安装了Golang环境  
> 下列命令使用Makefile定义，系统中需要安装Make(参考[安装make](https://cn.bing.com/search?q=%E5%AE%89%E8%A3%85+make))

- `make gone`   
  用于生成Priest代码，在IED中运行前**⚠️必须先执行**该命令，否则代码不完整无法运行

- `make run`  
  运行代码，已经集成了`make gone`

- `make build`  
  编译项目
- `make build-docker`   
  编译docker镜像，假设已经安装 [docker](https://www.docker.com/)
  和 [docker-compose](https://docs.docker.com/compose/install/)
- `make run-in-docker`  
  在docker中运行，假设已经安装 [docker](https://www.docker.com/)
  和 [docker-compose](https://docs.docker.com/compose/install/)
