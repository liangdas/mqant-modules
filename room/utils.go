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
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

//生成随机字符串
func GetRandomString(lenght int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < lenght; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

//## bigRoomId
//BR:{serverId}:{tableId}:{TransactionId}
//BR:Tacit@Tacit001:100:1235233
func BuildBigRoomId(serverId string, tableId int, transactionId int) string {
	return fmt.Sprintf("BR:%s:%d:%d", serverId, tableId, transactionId)
}
func ParseBigRoomId(bigroomId string) (string, int, int, error) {
	s := strings.Split(bigroomId, ":")
	if len(s) != 4 {
		return "", -1, 0, fmt.Errorf("The bigroomId data structure is incorrect")
	}
	tableId, error := strconv.Atoi(s[2])
	if error != nil {
		return "", -1, 0, error
	}
	transactionId, error := strconv.Atoi(s[3])
	if error != nil {
		return "", -1, 0, error
	}
	return s[1], tableId, transactionId, nil
}
