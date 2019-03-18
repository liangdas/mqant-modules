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
	"github.com/liangdas/mqant/module"
	"sync"
)

type Room struct {
	module      module.RPCModule
	lock        *sync.RWMutex
	tables      map[int]BaseTable
	newTable    func(module module.RPCModule, tableId int) (BaseTable, error)
	usableTable func(BaseTable) bool
	roomId      int
	index       int
	max         int
}

func NewRoom(module module.RPCModule, roomId int, newTable func(module module.RPCModule, tableId int) (BaseTable, error), usableTable func(BaseTable) bool) *Room {
	room := &Room{
		module:      module,
		lock:        new(sync.RWMutex),
		tables:      map[int]BaseTable{},
		newTable:    newTable,
		usableTable: usableTable,
		roomId:      roomId,
		index:       0,
		max:         0,
	}
	return room
}
func (self *Room) RoomId() int {
	return self.roomId
}
func (self *Room) Create(module module.RPCModule) (BaseTable, error) {
	self.lock.Lock()
	self.index++
	if table, ok := self.tables[self.index]; ok {
		self.lock.Unlock()
		return table, nil
	}
	self.lock.Unlock()
	table, err := self.CreateById(module, self.index)
	if err != nil {
		return nil, err
	}
	self.lock.Lock()
	self.tables[table.TableId()] = table
	self.lock.Unlock()
	return table, nil
}

func (self *Room) CreateById(module module.RPCModule, tableId int) (BaseTable, error) {
	self.lock.Lock()
	if table, ok := self.tables[tableId]; ok {
		self.lock.Unlock()
		return table, nil
	}
	self.lock.Unlock()
	table, err := self.newTable(module, tableId)
	if err != nil {
		return nil, err
	}
	self.lock.Lock()
	self.tables[tableId] = table
	self.lock.Unlock()
	return table, nil
}

func (self *Room) GetTable(tableId int) BaseTable {
	self.lock.Lock()
	if table, ok := self.tables[tableId]; ok {
		self.lock.Unlock()
		return table
	}
	self.lock.Unlock()
	return nil
}

/**
获取一个可用的桌
*/
func (self *Room) GetUsableTable() (BaseTable, error) {
	//先尝试获取没有满的房间
	for _, table := range self.tables {
		if self.usableTable(table) {
			return table, nil
		}
	}
	//没有找到已创建的空房间,新创建一个
	table, err := self.Create(self.module)
	if err != nil {
		return nil, err
	}
	return table, nil
}

//func (self *Room)GetEmptyTable()(BaseTable,error){
//	for _,table:=range self.tables{
//		if table.State()==Uninitialized{
//			return table,nil
//		}else if table.State()==Stoped{
//			return table,nil
//		}
//	}
//	//没有找到已创建的空房间,新创建一个
//	table,err:= self.Create(self.module)
//	if err!=nil{
//		return nil,err
//	}
//	return table,nil
//}
///**
//获取一个未满的桌
// */
//func (self *Room)GetNoFullTable()(BaseTable,error){
//	//先尝试获取没有满的房间
//	for _,table:=range self.tables{
//		if !table.Empty()&&!table.Full(){
//			return table,nil
//		}
//	}
//	//再尝试获取可能是空的房间
//	for _,table:=range self.tables{
//		if !table.Full(){
//			return table,nil
//		}
//	}
//	//没有找到已创建的空房间,新创建一个
//	table,err:= self.Create(self.module)
//	if err!=nil{
//		return nil,err
//	}
//	return table,nil
//}
