# 房间抽象模块

    用于mqant内房间类游戏的开发,此模块为抽象出来的房间牌桌逻辑

# 依赖模块

    go get github.com/liangdas/mqant
    go get github.com/yireyun/go-queue


# 使用方法

### 1. 导入项目

    go get github.com/liangdas/mqant-modules

### 2. 房间/牌桌结构

一个游戏模块进程一个room(一个游戏模块可以配置为多个进程,即代表多个room)

一个room中可以创建多个table,根据游戏服务器进程性能设定最大房间数

### 3. 牌桌结构

牌桌(table)是最重要的设计,游戏开发主要就是实现牌桌的具体逻辑

1. tableId

    牌桌唯一ID (进程内tableId从0开始自增)
2.  游戏帧
    每一个游戏都是一帧一帧驱动的,牌桌后端也是如此

### 4. 牌桌生命周期

    OnCreate()  //可以进行一些初始化的工作在table第一次被创建的时候调用
	OnStart()   //table创建完成，但还不可与用户交互，无法接收用户消息 开始：onCreate()->onStart() onStop()->onRestart()->onStart()
	OnRestart() //在table停止后，在再次启动之前被调用 重启  onStop()->onRestart()
	OnResume()  //取得控制权，可接受用户输入。 恢复：onCreate()->onStart()->onResume() onPause()->onResume() onStop()->onRestart()->onStart()->onResume()
	OnPause()   //table内暂停，可接收用户消息，此方法主要用来将未保存的变化进行持久化，停止游戏时钟等 暂停：onStart()->onPause()
	OnStop()    //当table不再提供服务时调用此方法，将无法再接收用户消息 停止:onPause()->onStop()
	OnDestroy() //在table销毁时调用 销毁：onPause()->onStop()->onDestroy()
	OnTimeOut() //当table超时了

### 5. 玩家状态

1. 加入游戏
2. 准备好(坐下) (游戏一般需要满足一定数量的玩家准备好才会开始)
3. 临时断线
4. 重连回
5. 玩家断线一段时间（超时）
6. 立刻房间 (游戏完成/退出)

### 6. 接收玩家输入
牌桌内实现了玩家输入缓存功能

简单的说就是玩家输入的消息*不会立刻处理*

而是会先放到玩家输入队列中

等待牌桌下一帧运行时*统一执行*

     玩家1 ---> 消息1--->

                       |消息队列|---> 下一帧统一处理

     玩家2 ---> 消息2--->

这样做的好处是可以避免牌桌内多协程消息输入造成资源锁问题,通过这种方式可以避免使用资源锁

### 7. 玩家管理

1. 玩家加入
2. 玩家准备好
3. 玩家超时管理

> 由于时间关系先不深入讲解,自己去看示例和源码吧

### 8. 玩家消息统一发送

暂不说明


### 9. 示例

    // 多人猜数字游戏
    // 游戏周期:
    //第一阶段 空闲期，什么都不做
    //第一阶段 空档期，等待押注
    //第二阶段 押注期  可以押注
    //第三阶段 开奖期  开奖
    //第四阶段 结算期  结算
    //玩法:
    //玩家可以押 0-9 中的一个数字,每押一次需要消耗500金币
    //押注完成后牌桌内随机开出一个 0-9 的数字
    //哪个玩家猜中的数字与系统开出的数字最近就算赢,可以赢取所有玩家本局押注的金币*80%
    //玩家金币用完后将被踢出房间

[猜数字游戏](https://github.com/liangdas/mqantserver/tree/master/src/server/xaxb)

[猜数字游戏机器人](https://github.com/liangdas/mqantserver/tree/master/src/robot)