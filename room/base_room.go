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
	app    module.App
	tables sync.Map
	roomId int
}

type NewTableFunc func(module module.App, tableId string) (BaseTable, error)

func NewRoom(module module.App) *Room {
	room := &Room{
		app: module,
	}
	return room
}
func (self *Room) RoomId() int {
	return self.roomId
}

func (self *Room) CreateById(app module.App, tableId string, newTablefunc NewTableFunc) (BaseTable, error) {
	if table, ok := self.tables.Load(tableId); ok {
		table.(BaseTable).Run()
		return table.(BaseTable), nil
	}
	table, err := newTablefunc(app, tableId)
	if err != nil {
		return nil, err
	}
	self.tables.Store(table.TableId(), table)
	return table, nil
}

func (self *Room) GetTable(tableId string) BaseTable {
	if table, ok := self.tables.Load(tableId); ok {
		table.(BaseTable).Run()
		return table.(BaseTable)
	}
	return nil
}

func (self *Room) DestroyTable(tableId string) error {
	self.tables.Delete(tableId)
	return nil
}
