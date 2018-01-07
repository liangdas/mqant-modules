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
	"container/list"
	"fmt"
	"github.com/liangdas/mqant/gate"
	"github.com/yireyun/go-queue"
)

type CallBackMsg struct {
	notify  bool     //是否是广播
	players []string //如果不是广播就指定session
	topic   *string
	body    *[]byte
}
type TableImp interface {
	OnNetBroken(BasePlayer)
	GetSeats() []BasePlayer
	GetViewer() *list.List
}
type UnifiedSendMessageTable struct {
	queue_message *queue.EsQueue
	tableimp      TableImp
}

func (this *UnifiedSendMessageTable) UnifiedSendMessageTableInit(tableimp TableImp) {
	this.queue_message = queue.NewQueue(256)
	this.tableimp = tableimp
}
func (this *UnifiedSendMessageTable) GetBindPlayer(session gate.Session) BasePlayer {
	for _, player := range this.tableimp.GetSeats() {
		if (player != nil) && (player.Session() != nil) {
			if player.Session().IsGuest() {
				if player.Session().GetSessionid() == session.GetSessionid() {
					player.OnRequest(session)
					return player
				}
			} else {
				if player.Session().GetUserid() == session.GetUserid() {
					player.OnRequest(session)
					return player
				}
			}

		}
	}
	return nil
}

func (this *UnifiedSendMessageTable) SendCallBackMsg(players []string, topic string, body []byte) error {
	ok, quantity := this.queue_message.Put(&CallBackMsg{
		notify:  false,
		players: players,
		topic:   &topic,
		body:    &body,
	})
	if !ok {
		return fmt.Errorf("Put Fail, quantity:%v\n", quantity)
	} else {
		return nil
	}
}

func (this *UnifiedSendMessageTable) NotifyCallBackMsg(topic string, body []byte) error {
	ok, quantity := this.queue_message.Put(&CallBackMsg{
		notify:  true,
		players: nil,
		topic:   &topic,
		body:    &body,
	})
	if !ok {
		return fmt.Errorf("Put Fail, quantity:%v\n", quantity)
	} else {
		return nil
	}
}

/**
【每帧调用】统一发送所有消息给各个客户端
*/
func (this *UnifiedSendMessageTable) ExecuteCallBackMsg() {
	ok := true
	queue := this.queue_message
	index := 0
	for ok {
		val, _ok, _ := queue.Get()
		index++
		if _ok {
			msg := val.(*CallBackMsg)
			if msg.notify {
				for _, role := range this.tableimp.GetSeats() {
					if role != nil && role.Session() != nil {
						netBroken, _ := role.GetNetBroken()
						if !netBroken {
							e := role.Session().Send(*msg.topic, *msg.body)
							if e != "" {
								if this.tableimp != nil {
									this.tableimp.OnNetBroken(role)
								}
							} else {
								role.OnResponse(role.Session())
							}
						}

					}

				}

				//通知观众玩家
				for e := this.tableimp.GetViewer().Front(); e != nil; e = e.Next() {
					err := e.Value.(gate.Session).Send(*msg.topic, *msg.body)
					if err != "" {
						this.tableimp.GetViewer().Remove(e)
					}
				}
			} else {
				for _, sessionId := range msg.players {
					for _, role := range this.tableimp.GetSeats() {
						if role != nil {
							if (role.Session().GetSessionid() == sessionId) && (role.Session() != nil) {
								netBroken, _ := role.GetNetBroken()
								if !netBroken {
									e := role.Session().Send(*msg.topic, *msg.body)
									if e != "" {
										if this.tableimp != nil {
											this.tableimp.OnNetBroken(role)
										}
									} else {
										role.OnResponse(role.Session())
									}
								}
							}
						}

					}
				}
			}
		}
		ok = _ok
	}
}
