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
	"github.com/liangdas/mqant/gate"
	"time"
)

type BasePlayerImp struct {
	session         gate.Session
	sitDown         bool  //是否已坐下 ,如果网络断开会设置为false,当网络连接成功以后需要重新坐下\
	netBroken       bool  //网络中断
	netBrokenDate   int64 //网络中断时间，超过60秒就踢出房间或者做其他处理 单位/秒
	lastRequestDate int64 //玩家最后一次请求时间	单位秒
}

func (self *BasePlayerImp) Bind() bool {
	if self.session == nil {
		return false
	} else {
		return true
	}
}

func (self *BasePlayerImp) OnBind(session gate.Session) BasePlayer {
	self.session = session
	self.netBroken = false
	return self
}
func (self *BasePlayerImp) OnUnBind() BasePlayer {
	self.session = nil
	self.OnSitUp()
	self.lastRequestDate = 0
	self.OnNetBroken()
	return self
}

/**
玩家主动发请求时间
*/
func (self *BasePlayerImp) OnRequest(session gate.Session) {
	self.session = session
	self.lastRequestDate = time.Now().Unix()
	self.netBroken = false
}

/**
服务器主动发送消息给客户端的时间
*/
func (self *BasePlayerImp) OnResponse(session gate.Session) {
	self.session = session
	self.lastRequestDate = time.Now().Unix()
	self.netBroken = false
}
func (self *BasePlayerImp) OnSitDown() {
	self.sitDown = true
}

func (self *BasePlayerImp) OnSitUp() {
	self.sitDown = false
}

func (self *BasePlayerImp) OnNetBroken() {
	self.netBrokenDate = time.Now().Unix()
	self.netBroken = true
}
func (self *BasePlayerImp) GetNetBroken() (bool, int64) {
	return self.netBroken, self.netBrokenDate
}
func (self *BasePlayerImp) GetLastRequestDate() int64 {
	return self.lastRequestDate
}

func (self *BasePlayerImp) Session() gate.Session {
	return self.session
}

func (self *BasePlayerImp) SitDown() bool {
	return self.sitDown
}
