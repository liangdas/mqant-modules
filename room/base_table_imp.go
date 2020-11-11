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
	"runtime"
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
func (this *QTable) GetApp() module.App {
	panic("implement func GetApp() module.App")
}
func (this *QTable) update(arge interface{}) {
	defer func() {
		if r := recover(); r != nil {
			buff := make([]byte, 1024)
			runtime.Stack(buff, false)
			log.Error("Update panic(%v)\n info:%s", r, string(buff))
			this.Finish()
		}
	}()
	if this.opts.ProUpdate != nil {
		this.opts.ProUpdate(time.Now().Sub(this.last_time_update))
	}
	this.ExecuteEvent(arge) //执行这一帧客户端发送过来的消息
	if this.opts.Update != nil {
		this.opts.Update(time.Now().Sub(this.last_time_update))
	}
	this.ExecuteCallBackMsg(this.Trace()) //统一发送数据到客户端
	this.CheckTimeOut()
	if this.opts.PostUpdate != nil {
		this.opts.PostUpdate(time.Now().Sub(this.last_time_update))
	}
	this.last_time_update = time.Now()
	if this.Runing() {
		timewheel.GetTimeWheel().AddTimer(this.opts.RunInterval, nil, this.update)
	}
}

func (this *QTable) OnCreate() {
	this.ResetTimeOut()
	this.last_time_update = time.Now()
	timewheel.GetTimeWheel().AddTimer(this.opts.RunInterval, nil, this.update)
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
	subtable.GetApp()
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

func (this *BaseTableImp) Runing() bool {
	if this.state == Active {
		return true
	}
	return false
}

//初始化table
func (this *BaseTableImp) Run() {
	if this.state != Active {
		this.state = Initialized
		this.subtable.OnCreate()
		this.state = Active
	}
}

//停止table
func (this *BaseTableImp) Finish() {
	if this.state == Initialized {
		this.subtable.OnDestroy()
		this.state = Finished
	} else if this.state == Active {
		this.subtable.OnDestroy()
		this.state = Finished
	} else if this.state == Uninitialized {
		this.subtable.OnDestroy()
		this.state = Finished
	}
}

//可以进行一些初始化的工作在table第一次被创建的时候调用
func (this *BaseTableImp) OnCreate() {
	panic("implement func OnCreate()")
}

//在table销毁时调用,将无法再接收用户消息 销毁：onPause()->onStop()->onDestroy()
func (this *BaseTableImp) OnDestroy() {
	panic("implement func OnDestroy()")
}

//在table超时是调用
func (this *BaseTableImp) OnTimeOut() {
	this.Finish()
}
