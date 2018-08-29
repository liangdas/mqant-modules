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
	"github.com/liangdas/mqant/module"
	"fmt"
	"github.com/liangdas/mqant/gate"
	"github.com/yireyun/go-queue"
	"strings"
	"github.com/liangdas/mqant/log"
)

type CallBackMsg struct {
	notify  	bool     	//是否是广播
	needReply	bool 		//是否需要回复
	players 	[]string 	//如果不是广播就指定session
	topic   	*string
	body    	*[]byte
}
type TableImp interface {
	OnNetBroken(BasePlayer)
	GetSeats() []BasePlayer
	GetViewer() *list.List
	GetModule() module.RPCModule
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
				if player.Session().GetSessionId() == session.GetSessionId() {
					player.OnRequest(session)
					return player
				}
			} else {
				if player.Session().GetUserId() == session.GetUserId() {
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
		needReply:true,
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
		needReply:true,
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

func (this *UnifiedSendMessageTable) SendCallBackMsgNR(players []string, topic string, body []byte) error {
	ok, quantity := this.queue_message.Put(&CallBackMsg{
		notify:  false,
		needReply:false,
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

func (this *UnifiedSendMessageTable) NotifyCallBackMsgNR(topic string, body []byte) error {
	ok, quantity := this.queue_message.Put(&CallBackMsg{
		notify:  true,
		needReply:false,
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
合并玩家所在网关
 */
func (this *UnifiedSendMessageTable) mergeGate() (map[string][]string){
	merge:=map[string][]string{}
	for _, role := range this.tableimp.GetSeats() {
		if role != nil && role.Session() != nil {
			netBroken, _ := role.GetNetBroken()
			if !netBroken {
				//未断网
				if _,ok:=merge[role.Session().GetServerId()];ok{
					merge[role.Session().GetServerId()]=append(merge[role.Session().GetServerId()],role.Session().GetSessionId())
				}else{
					merge[role.Session().GetServerId()]=[]string{role.Session().GetSessionId()}
				}
			}
		}
	}
	//通知观众玩家
	//for e := this.tableimp.GetViewer().Front(); e != nil; e = e.Next() {
	//	session:=e.Value.(gate.Session)
	//	if session != nil {
	//		if plist,ok:=merge[session.GetServerid()];ok{
	//			plist=append(plist,session.GetSessionid())
	//		}else{
	//			plist=[]string{session.GetSessionid()}
	//			merge[session.GetServerid()]=plist
	//		}
	//	}
	//}
	return merge
}

/**
【每帧调用】统一发送所有消息给各个客户端
*/
func (this *UnifiedSendMessageTable) ExecuteCallBackMsg() {
	var merge map[string][]string;
	ok := true
	queue := this.queue_message
	index := 0
	for ok {
		val, _ok, _ := queue.Get()
		index++
		if _ok {
			msg := val.(*CallBackMsg)
			if msg.notify {
				if merge==nil{
					merge=this.mergeGate()
				}
				for serverid,plist:=range merge{
					sessionids:=strings.Join(plist,",")
					server, e := this.tableimp.GetModule().GetApp().GetServerById(serverid)
					if e != nil {
						log.Warning("SendBatch error %v", e);
					}
					if msg.needReply{
						result,err:=server.Call("SendBatch",nil,sessionids,*msg.topic, *msg.body)
						if err != "" {
							log.Warning("SendBatch error %v %v",serverid,err);
						} else {
							if int(result.(int64))<len(plist){
								//有连接断了，牌桌可以广播一个心跳包去查一下
								this.NotifyHeartbeat()
							}
							for _, role := range this.tableimp.GetSeats() {
								if role != nil && role.Session() != nil {
									netBroken, _ := role.GetNetBroken()
									if !netBroken {
										//更新一下
										role.OnResponse(role.Session())
									}
								}
							}
						}
					}else{
						err:=server.CallNR("SendBatch",nil,sessionids,*msg.topic, *msg.body)
						if err != nil {
							log.Warning("SendBatch error %v %v",serverid,err.Error());
						}
					}

				}
				//for _, role := range this.tableimp.GetSeats() {
				//	if role != nil && role.Session() != nil {
				//		netBroken, _ := role.GetNetBroken()
				//		if !netBroken {
				//			var e string=""
				//			if msg.needReply{
				//				e= role.Session().Send(*msg.topic, *msg.body)
				//			}else{
				//				e = role.Session().SendNR(*msg.topic, *msg.body)
				//			}
				//			if e != "" {
				//				if this.tableimp != nil {
				//					this.tableimp.OnNetBroken(role)
				//				}
				//			} else {
				//				role.OnResponse(role.Session())
				//			}
				//		}
				//
				//	}
				//
				//}

				//通知观众玩家
				//for e := this.tableimp.GetViewer().Front(); e != nil; e = e.Next() {
				//	var err string=""
				//	if msg.needReply{
				//		err = e.Value.(gate.Session).Send(*msg.topic, *msg.body)
				//	}else{
				//		err = e.Value.(gate.Session).SendNR(*msg.topic, *msg.body)
				//	}
				//
				//	if err != "" {
				//		this.tableimp.GetViewer().Remove(e)
				//	}
				//}
			} else {
				for _, sessionId := range msg.players {
					for _, role := range this.tableimp.GetSeats() {
						if role != nil {
							if (role.Session().GetSessionId() == sessionId) && (role.Session() != nil) {
								netBroken, _ := role.GetNetBroken()
								if !netBroken {
									var e string=""
									if msg.needReply{
										e = role.Session().Send(*msg.topic, *msg.body)
									}else{
										e = role.Session().SendNR(*msg.topic, *msg.body)
									}

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
/**
给房间内的所有客户端发送一个心跳包检测
 */
func (this *UnifiedSendMessageTable) NotifyHeartbeat() error {
	for _, role := range this.tableimp.GetSeats() {
		if role != nil && role.Session() != nil {
			netBroken, _ := role.GetNetBroken()
			if !netBroken {
				e:= role.Session().Send("Table/HB", []byte("hb"))
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
	return nil
}