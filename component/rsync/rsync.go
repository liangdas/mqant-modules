// Copyright 2012 Julian Gutierrez Oschmann (github.com/julian-gutierrez-o).
// All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// A Golang implementation of the rsync algorithm.
// This package contains the algorithm for both client and server side.
package rsync

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"github.com/liangdas/mqant/log"
	"github.com/pkg/errors"
	"io"
	"math"
)

const (
	M                 = 1 << 16
	DeltaMagic uint16 = 0x7273
)

type LRsync struct {
	BlockSize int
}

type BlockHash struct {
	index      []int
	strongHash []byte
	weakHash   uint32
}

func (this *BlockHash) AddIndex(index int) {
	if this.index == nil {
		this.index = make([]int, 1)
		this.index[0] = index
	} else {
		for _, i := range this.index {
			if i == index {

				return
			}
		}
		this.index = append(this.index, index)
	}
}

/**
hu获取index,尽量选取大于且接近pro的
*/
func (this *BlockHash) GetIndex(pro int) int {
	var max int = 0
	if len(this.index) > 0 {
		max = this.index[0]
	}
	for _, i := range this.index {
		if math.Abs(float64(i-pro)) <= math.Abs(float64(max-pro)) {
			max = i
		}
	}
	//log.Info("index=%v pro=%v max=%v",this.index,pro,max)
	return max
}

// There are two kind of operations: BLOCK and DATA.
// If a block match is found on the server, a BLOCK operation is sent over the channel along with the block index.
// Modified data between two block matches is sent like a DATA operation.
const (
	BLOCK = iota
	DATA
)

// An rsync operation (typically to be sent across the network). It can be either a block of raw data or a block index.
type RSyncOp struct {
	// Kind of operation: BLOCK | DATA.
	opCode int
	// The raw modificated (or misaligned) data. Iff opCode == DATA, nil otherwise.
	data []byte
	// The index of found block. Iff opCode == BLOCK. nil otherwise.
	offset     int
	blockIndex int
}

// Returns weak and strong hashes for a given slice.
func (this *LRsync) CalculateBlockHashes(content []byte) map[string]*BlockHash {
	//blockHashes := make([]BlockHash, getBlocksNumber(content))
	//for i := range blockHashes {
	//	initialByte := i * BlockSize
	//	endingByte := min((i+1)*BlockSize, len(content))
	//	block := content[initialByte:endingByte]
	//	weak, _, _ := weakHash(block)
	//	blockHashes[i] = BlockHash{strongHash: strongHash(block), weakHash: weak}
	//	blockHashes[i].AddIndex(int32(i))
	//}
	//return blockHashes
	blockHashes := make(map[string]*BlockHash)
	num := this.getBlocksNumber(content)
	for i := 0; i < num; i++ {
		initialByte := i * this.BlockSize
		endingByte := min((i+1)*this.BlockSize, len(content))
		block := content[initialByte:endingByte]
		weak, _, _ := weakHash(block)
		if b, ok := blockHashes[string(strongHash(block))]; ok {
			b.AddIndex(i)
		} else {
			bb := &BlockHash{strongHash: strongHash(block), weakHash: weak}
			bb.AddIndex(i)
			blockHashes[string(strongHash(block))] = bb
		}
	}
	return blockHashes
}

// Returns the number of blocks for a given slice of content.
func (this *LRsync) getBlocksNumber(content []byte) int {
	blockNumber := (len(content) / this.BlockSize)
	if len(content)%this.BlockSize != 0 {
		blockNumber += 1
	}
	return blockNumber
}

// Applies operations from the channel to the original content.
// Returns the modified content.
func (this *LRsync) ApplyOps(content []byte, ops []RSyncOp, fileSize int) []byte {
	var offset int
	result := make([]byte, fileSize)
	for _, op := range ops {
		switch op.opCode {
		case BLOCK:
			copy(result[offset:offset+this.BlockSize], content[op.blockIndex*this.BlockSize:op.blockIndex*this.BlockSize+this.BlockSize])
			offset += this.BlockSize
		case DATA:
			copy(result[offset:], op.data)
			offset += len(op.data)
		}
	}
	return result
}

// Computes all the operations needed to recreate content.
// All these operations are sent through a channel of RSyncOp.
func (this *LRsync) CalculateDifferences(content []byte, hashes map[string]*BlockHash) (opsChannel []RSyncOp) {

	hashesMap := make(map[uint32][]BlockHash)
	opsChannel = make([]RSyncOp, 0)

	for _, h := range hashes {
		key := h.weakHash
		hashesMap[key] = append(hashesMap[key], *h)
	}

	var offset, previousMatch int
	var aweak, bweak, weak uint32
	var pro int
	var dirty, isRolling bool

	for offset < len(content) {
		//log.Info("hashesMap offset=%v content=%v",offset,len(content))
		endingByte := min(offset+this.BlockSize, len(content)-1)
		block := content[offset:endingByte]
		if !isRolling {
			weak, aweak, bweak = weakHash(block)
			isRolling = true
		} else {
			aweak = (aweak - uint32(content[offset-1]) + uint32(content[endingByte-1])) % M
			bweak = (bweak - (uint32(endingByte-offset) * uint32(content[offset-1])) + aweak) % M
			weak = aweak + (1 << 16 * bweak)
		}
		if l := hashesMap[weak]; l != nil {
			blockFound, blockHash := searchStrongHash(hashes, strongHash(block))
			if blockFound {
				if dirty {
					pro++
					opsChannel = append(opsChannel, RSyncOp{opCode: DATA, offset: pro, data: content[previousMatch:offset]})
					dirty = false
				}
				pro++
				opsChannel = append(opsChannel, RSyncOp{opCode: BLOCK, offset: pro, blockIndex: blockHash.GetIndex(pro)})
				previousMatch = endingByte
				isRolling = false
				offset += this.BlockSize
				continue
			}
		}
		dirty = true
		offset++
	}

	if dirty {
		pro++
		opsChannel = append(opsChannel, RSyncOp{opCode: DATA, offset: pro, data: content[previousMatch:]})
	}
	return opsChannel
}

func Int8ToBytes(x int8) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

const (
	RS_OP_END      uint8 = 0x00 //结束
	RS_OP_BLOCK_N1 uint8 = 0x01 //匹配块index 1字节 0-256 个BLOCK内的数据有效
	RS_OP_BLOCK_N2 uint8 = 0x02
	RS_OP_BLOCK_N4 uint8 = 0x03
	RS_OP_DATA_N1  uint8 = 0x04 //未匹配块  数据长度 0-256
	RS_OP_DATA_N2  uint8 = 0x05
	RS_OP_DATA_N4  uint8 = 0x06
)

/**
填充规则
mode   1 字节

BLOCK
START_INDEX
END_INDEX

DATA
LENGHT
DATA[]

*/
func (this *LRsync) CreateDelta(opsChannel []RSyncOp, modifiedSize int, crc32 uint32) []byte {
	b := bytes.NewBuffer(make([]byte, 0))
	writer := bufio.NewWriter(b)
	writer.Write(hton2(DeltaMagic))
	writer.Write(hton4(crc32))
	writer.Write(Htonl(uint32(this.BlockSize)))
	writer.Write(Htonl(uint32(modifiedSize)))
	totls := 0
	var opCode = -1
	var startblock int = 0
	var problock int = 0
	for _, op := range opsChannel {

		switch op.opCode {
		case BLOCK:
			totls++
			if opCode == BLOCK {
				if (problock + 1) == op.blockIndex {
					//上一个op也是匹配块,则继续累加
					problock = op.blockIndex
					//log.TInfo(nil,"累加 %v offset %v" ,op.blockIndex,op.offset)
				} else {
					//应该是重头开始
					if len(opsChannel) < 256 {
						writer.Write([]byte{RS_OP_BLOCK_N1}) //写块的index
						writer.Write(Int8ToBytes(int8(startblock)))
						writer.Write(Int8ToBytes(int8(problock)))
					} else if 256 <= len(opsChannel) && len(opsChannel) < 65535 {
						writer.Write([]byte{RS_OP_BLOCK_N2}) //写块的index
						writer.Write(hton2(uint16(startblock)))
						writer.Write(hton2(uint16(problock)))
					} else {
						writer.Write([]byte{RS_OP_BLOCK_N4}) //写块的index
						writer.Write(Htonl(uint32(startblock)))
						writer.Write(Htonl(uint32(problock)))
					}
					//log.TInfo(nil,"应该填充数据了 opCode=%v startblock=%v problock=%v offset %v",opCode,startblock,problock,op.offset)
					startblock = op.blockIndex
					problock = op.blockIndex
					//log.TInfo(nil,"重头开始 startblock %v problock %v offset %v" ,startblock,problock,op.offset)
				}
			} else if opCode == DATA {
				//上一个是未匹配块
				startblock = op.blockIndex
				problock = op.blockIndex
				//log.TInfo(nil,"重头开始 startblock %v  blockIndex %v offset %v" ,startblock,problock,op.offset)
			} else {
				//上一个op未知类型
				startblock = op.blockIndex
				problock = op.blockIndex

				//log.TInfo(nil,"重头开始 startblock %v  blockIndex %v offset %v" ,startblock,problock,op.offset)
			}
			opCode = BLOCK
			//log.TInfo(nil,"blockIndex %v",op.blockIndex)
			//writer.Write(Int16ToBytes(int16(op.blockIndex))) //写块的index
		case DATA:
			totls++
			if opCode == BLOCK {
				//log.TInfo(nil,"填充数据 opCode=%v startblock=%v  blockIndex %v offset=%v",opCode,startblock,problock,op.offset)
				if len(opsChannel) < 256 {
					writer.Write([]byte{RS_OP_BLOCK_N1}) //写块的index
					writer.Write(Int8ToBytes(int8(startblock)))
					writer.Write(Int8ToBytes(int8(problock)))
				} else if 256 <= len(opsChannel) && len(opsChannel) < 65535 {
					writer.Write([]byte{RS_OP_BLOCK_N2}) //写块的index
					writer.Write(hton2(uint16(startblock)))
					writer.Write(hton2(uint16(problock)))
				} else {
					writer.Write([]byte{RS_OP_BLOCK_N4}) //写块的index
					writer.Write(Htonl(uint32(startblock)))
					writer.Write(Htonl(uint32(problock)))
				}
			}
			opCode = DATA
			if len(op.data) < 256 {
				writer.Write([]byte{RS_OP_DATA_N1})
				writer.Write(Int8ToBytes(int8(len(op.data)))) //写块的数据长度
			} else if 256 <= len(opsChannel) && len(opsChannel) < 65025 {
				writer.Write([]byte{RS_OP_DATA_N2})
				writer.Write(hton2(uint16(len(op.data)))) //写块的数据长度
			} else {
				writer.Write([]byte{RS_OP_DATA_N4})
				writer.Write(Htonl(uint32(len(op.data)))) //写块的数据长度
			}
			writer.Write(op.data) //写块的数据长度
			//log.TInfo(nil,"填充未匹配数据 opCode=%v blockIndex %v len %v offset=%v",opCode,op.blockIndex,int16(len(op.data)),op.offset)
		default:
			log.Info("未知类型 %v", op)
		}
	}
	writer.Write([]byte{RS_OP_END})
	//for _,op:=range opsChannel{
	//	switch op.opCode {
	//	case DATA:
	//		writer.Write(Int16ToBytes(int16(op.blockIndex))) //写块的index
	//		writer.Write(Int16ToBytes(int16(len(op.data)))) //写块的数据长度
	//		writer.Write(op.data) //写块的数据长度
	//	}
	//}
	writer.Flush()
	return b.Bytes()
}

func (this *LRsync) Patch(content []byte, delta []byte) (result []byte, err error) {
	var (
		offset int
		cmd    uint8
	)
	rd := bytes.NewReader(delta)
	//deltamagic
	if deltamagic, errr := ntoh2(rd); errr == io.EOF {
		return
	} else {
		if deltamagic != DeltaMagic {
			err = errors.New("DeltaMagic error")
			return
		}
	}
	//crc32
	if _, err = ntoh4(rd); err == io.EOF {
		return
	}
	var BlockSize int = 0
	if blocksize, errr := ntoh4(rd); errr == io.EOF {
		err = errr
		return
	} else {
		BlockSize = int(blocksize)
	}
	var modifiedSize uint32
	if modifiedSize, err = ntoh4(rd); err == io.EOF {
		return
	}
	result = make([]byte, modifiedSize)
	for {
		if cmd, err = readByte(rd); err == io.EOF {
			err = nil
			break
		}
		if cmd == 0 { // delta的结束命令
			break
		}
		switch uint8(cmd) {
		case RS_OP_BLOCK_N1:
			start_index, err := readByte(rd)
			if err != nil {
				return nil, err
			}
			end_index, err := readByte(rd)
			if err != nil {
				return nil, err
			}
			for i := start_index; i <= end_index; i++ {
				var cp = int(i) * int(BlockSize)
				var cc = int(i)*int(BlockSize) + int(BlockSize)
				//log.Info("start_index %v end_index %v i %v cp %v cc %v",start_index,end_index,i,cp,cc)
				copy(result[offset:offset+BlockSize], content[cp:cc])
				offset += BlockSize
			}
		case RS_OP_BLOCK_N2:
			var start_index uint16 = 0
			var end_index uint16 = 0
			if start_index, err = ntoh2(rd); err != nil {
				err = fmt.Errorf("read signature maigin failed: %s", err.Error())
				return nil, err
			}
			if end_index, err = ntoh2(rd); err != nil {
				err = fmt.Errorf("read signature maigin failed: %s", err.Error())
				return nil, err
			}
			for i := start_index; i <= end_index; i++ {
				var cp = int(i) * int(BlockSize)
				var cc = int(i)*int(BlockSize) + int(BlockSize)
				copy(result[offset:offset+BlockSize], content[cp:cc])
				offset += BlockSize
			}
		case RS_OP_BLOCK_N4:
			var start_index uint32 = 0
			var end_index uint32 = 0
			if start_index, err = ntoh4(rd); err != nil {
				err = fmt.Errorf("read signature maigin failed: %s", err.Error())
				return nil, err
			}
			if end_index, err = ntoh4(rd); err != nil {
				err = fmt.Errorf("read signature maigin failed: %s", err.Error())
				return nil, err
			}
			for i := start_index; i <= end_index; i++ {
				var cp = int(i) * int(BlockSize)
				var cc = int(i)*int(BlockSize) + int(BlockSize)
				copy(result[offset:offset+BlockSize], content[cp:cc])
				offset += BlockSize
			}
		case RS_OP_DATA_N1:
			var lenght uint8 = 0
			if lenght, err = readByte(rd); err != nil {
				err = fmt.Errorf("read signature maigin failed: %s", err.Error())
				return nil, err
			}
			var (
				n   int
				buf []byte = make([]byte, lenght)
			)
			n, err = io.ReadFull(rd, buf)
			if uint8(n) != lenght {
				err = fmt.Errorf("uint8(n) %v != lenght %v", uint8(n), lenght)
				return nil, err
			}
			copy(result[offset:], buf)
			offset += len(buf)
		case RS_OP_DATA_N2:
			var lenght uint16 = 0
			if lenght, err = ntoh2(rd); err != nil {
				err = fmt.Errorf("read signature maigin failed: %s", err.Error())
				return nil, err
			}
			var (
				n   int
				buf []byte = make([]byte, lenght)
			)
			n, err = io.ReadFull(rd, buf)
			if uint16(n) != lenght {
				err = fmt.Errorf("int16(n) %v != lenght %v", uint8(n), lenght)
				return nil, err
			}
			copy(result[offset:], buf)
			offset += len(buf)
		case RS_OP_DATA_N4:
			var lenght uint32 = 0
			if lenght, err = ntoh4(rd); err != nil {
				err = fmt.Errorf("read signature maigin failed: %s", err.Error())
				return nil, err
			}
			var (
				n   int
				buf []byte = make([]byte, lenght)
			)
			n, err = io.ReadFull(rd, buf)
			if uint32(n) != lenght {
				err = fmt.Errorf("uint32(n) %v != lenght %v", uint8(n), lenght)
				return nil, err
			}
			copy(result[offset:], buf)
			offset += len(buf)
		}
	}

	return result, nil
}

// Searches for a given strong hash among all strong hashes in this bucket.
func searchStrongHash(l map[string]*BlockHash, hashValue []byte) (bool, *BlockHash) {
	//for _, blockHash := range l {
	//	if string(blockHash.strongHash) == string(hashValue) {
	//		return true, &blockHash
	//	}
	//}
	if blockHash, ok := l[string(hashValue)]; ok {
		return true, blockHash
	}
	return false, nil
}

// Returns a strong hash for a given block of data
func strongHash(v []byte) []byte {
	h := md5.New()
	h.Write(v)
	return h.Sum(nil)
}

// Returns a weak hash for a given block of data.
func weakHash(v []byte) (uint32, uint32, uint32) {
	var a, b uint32
	for i := range v {
		a += uint32(v[i])
		b += (uint32(len(v)-1) - uint32(i) + 1) * uint32(v[i])
	}
	return (a % M) + (1 << 16 * (b % M)), a % M, b % M
}

// Returns the smaller of a or b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
