// Copyright 2012 Julian Gutierrez Oschmann (github.com/julian-gutierrez-o).
// All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Unit tests for core package
package rsync

import "testing"
import (
	"github.com/liangdas/mqant/log"
	"hash/crc32"
	"io/ioutil"
)

type filePair struct {
	original string
	modified string
}

func Test_SyncModifiedContent(t *testing.T) {
	//,filePair{"golang-original.bmp", "golang-modified.bmp"}
	files := []filePair{filePair{"text-original.txt", "text-modified.txt"}}

	for _, filePair := range files {
		original, _ := ioutil.ReadFile("test-data/" + filePair.original)
		modified, _ := ioutil.ReadFile("test-data/" + filePair.modified)
		rs := &LRsync{
			BlockSize: 16,
		}
		hashes := rs.CalculateBlockHashes(original)
		opsChannel := rs.CalculateDifferences(modified, hashes)
		log.Info("CalculateDifferences %v", filePair)
		//result := ApplyOps(original, opsChannel, len(modified))
		ieee := crc32.NewIEEE()
		ieee.Write(modified)
		s := ieee.Sum32()
		log.Info("IEEE modified = 0x%x", s)
		delta := rs.CreateDelta(opsChannel, len(modified), s)
		result, err := rs.Patch(original, delta)
		if err != nil {
			log.Info("rsync did not work as expected for %v  error %v", filePair, err)
		}
		log.TInfo(nil, "original %v opsChannel %v CreateDelta %v", len(original), len(opsChannel), len(delta))
		if string(result) != string(modified) {
			t.Errorf("rsync did not work as expected for %v", filePair)
		}
	}
}

func Test_WeakHash(t *testing.T) {
	content := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	expectedWeak := uint32(10813485)
	expectedA := uint32(45)
	expectedB := uint32(165)
	weak, a, b := weakHash(content)

	assertHash(t, "weak", content, expectedWeak, weak)
	assertHash(t, "a", content, expectedA, a)
	assertHash(t, "b", content, expectedB, b)
}

func assertHash(t *testing.T, name string, content []byte, expected uint32, found uint32) {
	if found != expected {
		t.Errorf("Incorrent "+name+" hash for %v - Expected %d - Found %d", content, expected, found)
	}
}
