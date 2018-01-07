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

import ()
import (
	"github.com/liangdas/mqant/log"
	"time"
)

type BaseTableImp struct {
	BaseTable
	state         int //当前写的队列
	tableId       int
	transactionId int
	subtable      BaseTable
}

func (this *BaseTableImp) BaseTableImpInit(tableId int, subtable BaseTable) {
	this.state = Uninitialized
	this.subtable = subtable
	this.tableId = tableId
	this.transactionId = 0
}
func (this *BaseTableImp) TableId() int {
	return this.tableId
}
func (this *BaseTableImp) AllowJoin() bool {
	//默认运行加入条件
	if this.State() == Uninitialized {
		return true
	}
	return false
}
func (this *BaseTableImp) TransactionId() int {
	return this.transactionId
}

//uninitialized active paused stoped destroyed
func (this *BaseTableImp) State() int {
	return this.state
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
	if this.state == Active {
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
	this.transactionId = int(time.Now().Unix())
	log.Debug("BaseTableImp", "OnCreate")
}

//table创建完成，一次游戏开始，可以在这里初始化游戏数据 开始：onCreate()->onStart() onStop()->onRestart()->onStart()
func (this *BaseTableImp) OnStart() {
	log.Debug("BaseTableImp", "OnStart")
}

//在table停止后，在再次启动之前被调用 重启  onStop()->onRestart()
func (this *BaseTableImp) OnRestart() {
	log.Debug("BaseTableImp", "OnRestart")
}

//取得控制权，可接受用户输入。 恢复：onCreate()->onStart()->onResume() onPause()->onResume() onStop()->onRestart()->onStart()->onResume()
func (this *BaseTableImp) OnResume() {
	log.Debug("BaseTableImp", "OnResume")
}

//table内暂停，可接收用户消息,此方法主要用在游戏过程中的游戏时钟暂停,不销毁本次游戏的数据 暂停：onStart()->onPause()
func (this *BaseTableImp) OnPause() {
	log.Debug("BaseTableImp", "OnPause")
}

//当本次游戏完成时调用,这里需要销毁游戏数据，对游戏数据做本地化处理，比如游戏结算等 停止:onPause()->onStop()
func (this *BaseTableImp) OnStop() {
	log.Debug("BaseTableImp", "OnStop")
}

//在table销毁时调用,将无法再接收用户消息 销毁：onPause()->onStop()->onDestroy()
func (this *BaseTableImp) OnDestroy() {
	this.transactionId = 0
	log.Debug("BaseTableImp", "OnDestroy")
}

//在table超时是调研
func (this *BaseTableImp) OnTimeOut() {
	log.Debug("BaseTableImp", "OnTimeOut")
	this.Finish()
}
