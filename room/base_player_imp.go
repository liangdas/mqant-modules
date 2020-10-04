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
	session      gate.Session
	lastNewsDate int64 //玩家最后一次成功通信时间	单位秒
	body         interface{}
}

func (self *BasePlayerImp) Type() string {
	return "BasePlayer"
}

func (self *BasePlayerImp) IsBind() bool {
	if self.session == nil {
		return false
	} else {
		return true
	}
}

func (self *BasePlayerImp) Bind(session gate.Session) BasePlayer {
	self.lastNewsDate = time.Now().Unix()
	self.session = session
	return self
}

func (self *BasePlayerImp) UnBind() error {
	self.session = nil
	return nil
}

/**
玩家主动发请求时间
*/
func (self *BasePlayerImp) OnRequest(session gate.Session) {
	self.session = session
	self.lastNewsDate = time.Now().Unix()
}

/**
服务器主动发送消息给客户端的时间
*/
func (self *BasePlayerImp) OnResponse(session gate.Session) {
	self.session = session
	self.lastNewsDate = time.Now().Unix()
}

func (self *BasePlayerImp) GetLastReqResDate() int64 {
	return self.lastNewsDate
}
func (self *BasePlayerImp) Body() interface{} {
	return self.body
}

func (self *BasePlayerImp) SetBody(body interface{}) {
	self.body = body
}

func (self *BasePlayerImp) Session() gate.Session {
	return self.session
}
