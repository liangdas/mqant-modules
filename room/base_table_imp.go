// Copyright 2014 loolgame Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package room

import (
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/modules/timer"
	"time"
)

type SubTable interface {
	BaseTable
	TableImp
}

type QTable struct {
	BaseTableImp
	QueueTable
	UnifiedSendMessageTable
	TimeOutTable
	last_time_update time.Time
	opts             Options
}

func (this *QTable) GetSeats() map[string]BasePlayer {
	panic("implement func GetSeats() map[string]BasePlayer")
}
func (this *QTable) GetModule() module.RPCModule {
	panic("implement func GetModule() module.RPCModule")
}
func (this *QTable) update(arge interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("update error %v", r)
		}
	}()
	this.ExecuteEvent(arge) //执行这一帧客户端发送过来的消息
	if this.opts.Update != nil {
		this.opts.Update(time.Now().Sub(this.last_time_update))
	}
	this.ExecuteCallBackMsg(this.Trace()) //统一发送数据到客户端
	this.CheckTimeOut()
	if this.Runing() {
		timewheel.GetTimeWheel().AddTimer(50*time.Millisecond, nil, this.update)
	}
}

//可以进行一些初始化的工作在table第一次被创建的时候调用
func (this *QTable) OnCreate() {
	this.ResetTimeOut()
	this.last_time_update = time.Now()
	timewheel.GetTimeWheel().AddTimer(50*time.Millisecond, nil, this.update)
}

func (this *QTable) OnDestroy() {
	if this.opts.DestroyCallbacks != nil {
		err := this.opts.DestroyCallbacks(this)
		if err != nil {
			log.Error("DestroyCallbacks %v", err)
		}
	}
}

func (this *QTable) OnInit(subtable SubTable, opts ...Option) error {
	subtable.GetSeats()
	subtable.GetModule()
	this.opts = newOptions(opts...)
	this.last_time_update = time.Now()
	this.BaseTableImpInit(subtable, opts...)
	this.QueueInit(opts...)
	this.UnifiedSendMessageTableInit(subtable, this.opts.SendMsgCapaciity)
	this.TimeOutTableInit(subtable, this.opts.TimeOut)
	return nil
}

type BaseTableImp struct {
	opts     Options
	trace    log.TraceSpan
	state    int //当前写的队列
	subtable BaseTable
}

func (this *BaseTableImp) BaseTableImpInit(subtable BaseTable, opts ...Option) {
	this.opts = newOptions(opts...)
	this.state = Uninitialized
	this.subtable = subtable
	this.trace = log.CreateRootTrace()
}

func (this *BaseTableImp) Options() Options {
	return this.opts
}

func (this *BaseTableImp) TableId() string {
	return this.opts.TableId
}
func (this *BaseTableImp) Trace() log.TraceSpan {
	return this.trace
}
func (this *BaseTableImp) SetTrace(span log.TraceSpan) {
	this.trace = span
}

//uninitialized active paused stoped destroyed
func (this *BaseTableImp) State() int {
	return this.state
}

func (this *BaseTableImp) Runing() bool {
	if this.state != Uninitialized {
		return true
	}
	return false
}

//初始化table
func (this *BaseTableImp) Create() {
	if this.state == Uninitialized {
		this.state = Initialized
		this.subtable.OnCreate()
	}
}

//开始一次游戏
func (this *BaseTableImp) Start() {
	if this.state == Uninitialized {
		this.state = Initialized
		this.subtable.OnCreate()
		this.state = Active
		this.subtable.OnStart()
		this.subtable.OnResume()
	} else if this.state == Initialized {
		this.state = Active
		this.subtable.OnStart()
		this.subtable.OnResume()
	} else if this.state == Stoped {
		this.state = Active
		this.subtable.OnRestart()
		this.subtable.OnStart()
		this.subtable.OnResume()
	}
}

//暂停
func (this *BaseTableImp) Pause() {
	if this.state == Active {
		this.state = Paused
		this.subtable.OnPause()
	}
}

//停止
func (this *BaseTableImp) Stop() {
	if this.state == Active {
		this.state = Paused
		this.subtable.OnPause()
		this.state = Stoped
		this.subtable.OnStop()
	} else if this.state == Paused {
		this.state = Stoped
		this.subtable.OnStop()
	}
}

//重新开始
func (this *BaseTableImp) Restart() {
	if this.state == Stoped {
		this.state = Initialized
		this.subtable.OnRestart()
		this.subtable.OnStart()
		this.state = Active
		this.subtable.OnResume()
	}
}

//重新开始
func (this *BaseTableImp) Resume() {
	if this.state == Paused {
		this.state = Active
		this.subtable.OnResume()
	}
}

//停止table
func (this *BaseTableImp) Finish() {
	if this.state == Initialized {
		this.state = Uninitialized
		this.subtable.OnDestroy()
	} else if this.state == Active {
		this.subtable.OnPause()
		this.subtable.OnStop()
		this.state = Uninitialized
		this.subtable.OnDestroy()
	} else if this.state == Paused {
		this.subtable.OnStop()
		this.state = Uninitialized
		this.subtable.OnDestroy()
	} else if this.state == Stoped {
		this.state = Uninitialized
		this.subtable.OnDestroy()
	} else if this.state == Initialized {
		this.state = Uninitialized
		this.subtable.OnDestroy()
	}
}

//可以进行一些初始化的工作在table第一次被创建的时候调用
func (this *BaseTableImp) OnCreate() {
	panic("implement func OnCreate()")
}

//table创建完成，一次游戏开始，可以在这里初始化游戏数据 开始：onCreate()->onStart() onStop()->onRestart()->onStart()
func (this *BaseTableImp) OnStart() {
	//log.TInfo(this.Trace(),"BaseTableImp %v", "OnStart")
}

//在table停止后，在再次启动之前被调用 重启  onStop()->onRestart()
func (this *BaseTableImp) OnRestart() {
	//log.TInfo(this.Trace(),"BaseTableImp %v", "OnRestart")
}

//取得控制权，可接受用户输入。 恢复：onCreate()->onStart()->onResume() onPause()->onResume() onStop()->onRestart()->onStart()->onResume()
func (this *BaseTableImp) OnResume() {
	//log.TInfo(this.Trace(),"BaseTableImp %v", "OnResume")
}

//table内暂停，可接收用户消息,此方法主要用在游戏过程中的游戏时钟暂停,不销毁本次游戏的数据 暂停：onStart()->onPause()
func (this *BaseTableImp) OnPause() {
	//log.TInfo(this.Trace(),"BaseTableImp %v", "OnPause")
}

//当本次游戏完成时调用,这里需要销毁游戏数据，对游戏数据做本地化处理，比如游戏结算等 停止:onPause()->onStop()
func (this *BaseTableImp) OnStop() {
	//log.TInfo(this.Trace(),"BaseTableImp %v", "OnStop")
}

//在table销毁时调用,将无法再接收用户消息 销毁：onPause()->onStop()->onDestroy()
func (this *BaseTableImp) OnDestroy() {
	panic("implement func OnDestroy()")
}

//在table超时是调用
func (this *BaseTableImp) OnTimeOut() {
	this.Finish()
}
