# 短信验证码模块

    可以在mqant任何模块中方便的使用阿里云和sendcloud发送短信验证码

# 依赖模块

    go get github.com/liangdas/mqant

# 外部依赖

    1. redis
    2. http

# 使用方法

### 1，导入项目

    go get github.com/liangdas/mqant-modules

### 2，将模块加入启动列表

    app.Run(true, //只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
    		smscode.Module(),
    		。。。。
    	)

### 3，配置文件中加入模块配置

    {
        //.....
        "Module":{
            //.....
            "SMS":[
                {
                    "Id":"SMS001",
                    "ProcessID":"development",
                    "Settings":{
                        //用于限制每一个手机号码发送短信时间间隔(TTL)
                        "RedisUrl":  "redis://:[user]@[ip]:[port]/[db]",
                        "TTL":60,
                        //sendcloud后台申请参数
                        "SendCloud":{
                            "SmsUser":    "xxx",
                            "SmsKey":    "xxx",
                            "TemplateId":"xxx"
                        },
                        //阿里云后台申请参数
                        "Ailyun":{
                            "AccessKeyId":"xxx",
                            "AccessSecret":"xxx",
                            "SignName":"xxx",
                            "TemplateCode":"xxx"
                        }
                     },
                     //mqant  rpc 通信配置
                    "Redis":{
                        "Uri"          :"redis://:[user]@[ip]:[port]/[db]",
                        "Queue"        :"SMS001"
                    }
                }
            ]
        }
    }

### 4，在mqant模块中通过rpc远程调用发送验证码

    extra:=map[string]interface{}{}
    m,err:=self.module.RpcInvoke("SMS","SendVerifiycode",phone,purpose,extra)
    if err!=""{
    	//验证码发送失败
    }


传入参数:

| 参数     | 参数类型 |   是否必选  | 说明  |
| :-------- |:--:| --------:| :--: |
| phone  | string|是 |  手机号码   |
| purpose  | string |是 |  验证码用途   |
| extra  | map |是 |  额外参数   |


