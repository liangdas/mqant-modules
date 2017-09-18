# 分布式跟踪服务

    基于【Appdash，用Go实现的分布式系统跟踪神器】的分布式跟踪服务模块

[如何在mqant中使用分布式跟踪功能](http://www.mqant.com/topic/59463345bf9668524f4ed685)
# 依赖模块

    go get github.com/liangdas/mqant
    go get sourcegraph.com/sourcegraph/appdash
    go get sourcegraph.com/sourcegraph/appdash/traceapp


# 使用方法

### 1，导入项目

    go get github.com/liangdas/mqant-modules

### 2，将模块加入启动列表

    app.Run(true, //只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
    		tracing.Module(),
    		。。。。
    	)

### 3，配置文件中加入模块配置

    {
        //.....
        "Module":{
            //.....
            "Tracing":[
                {
                    "Id":"Tracing001",
                    "ProcessID":"development",
                    "Settings":{
                        //日志数据保存路径
                        "StoreFile":     "/tmp/appdash.gob",
                        //控制台访问路径
                        "URL":     	 "http://localhost:7700",
                        //日志收集监控端口
                        "CollectorAddr":":7701",
                        //控制台HTTP服务监控端口与URL对应
                        "HTTPAddr":   ":7700"
                    }
                }

            ]
        }
    }

[如何在mqant中使用分布式跟踪功能](http://www.mqant.com/topic/59463345bf9668524f4ed685)