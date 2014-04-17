package bloomfilter

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	bf := NewBloomFilter(3, 100)
	d1, d2 := []byte("Hello"), []byte("Jello")
	bf.Add(d1)

	if !bf.Check(d1) {
		t.Errorf("d1 should be present in the BloomFilter")
	}

	if bf.Check(d2) {
		t.Errorf("d2 should be absent from the BloomFilter")
	}
}

func TestStreamingBF(t *testing.T) {
	k, bfSize := 4, 1000
	bf := NewBloomFilter(k, bfSize)
	var errRate float64
	tol := 1 / 10.0
	var i uint64
	var count uint64 = 100
	var num uint64
	for i = 0; i < count; i++ {
		elt := make([]byte, 10)
		num = 1 << ((uint8(i) - 1) % 32)
		num += i
		binary.PutUvarint(elt, num)

		if !bf.Check(elt) {
			bf.Add(elt)
			fmt.Println(num)
		}
		errRate = bf.FalsePositiveRate()
		if errRate > tol {
			fmt.Println("False Positive Rate is too high: ", errRate)
		}
	}
}

func TestCountingBFBasic(t *testing.T) {
	cbf := NewCountingBloomFilter(3, 100)
	d1 := []byte("Hello")
	cbf.Add(d1)

	if !cbf.Check(d1) {
		t.Errorf("d1 should be present in the BloomFilter")
	}

	cbf.Remove(d1)

	if cbf.Check(d1) {
		t.Errorf("d1 should be absent from the BloomFilter after deletion")
	}
}

func TestScalableBFBasic(t *testing.T) {
	sbf := NewScalableBloomFilter(3, 20, 4, 10, 0.01)

	for i := 1; i < 1000; i++ {
		buf := make([]byte, 8)
		binary.PutVarint(buf, int64(i))
		sbf.Add(buf)
		if !sbf.Check(buf) {
			t.Errorf("%d should be present in the BloomFilter", i)
			return
		}
	}

	for i := 1; i < 1000; i++ {
		buf := make([]byte, 8)
		binary.PutVarint(buf, int64(i))
		if !sbf.Check(buf) {
			t.Errorf("%d should be present in the BloomFilter", i)
			return
		}
	}

	count := 0

	for i := 1000; i < 4000; i++ {
		buf := make([]byte, 8)
		binary.PutVarint(buf, int64(i))
		if sbf.Check(buf) {
			count++
		}
	}

	if sbf.FalsePositiveRate() > 0.04 {
		t.Errorf("False Positive Rate for this test should be < 0.04")
		return
	}

	sensitivity := 0.01 // TODO Make this configurable
	expectedFalsePositives :=
		(int)((4000 - 1000) * (sbf.FalsePositiveRate() + sensitivity))
	if count > expectedFalsePositives {
		t.Errorf("Actual false positives %d is greater than max expected false positives %d",
			count,
			expectedFalsePositives)
		return
	}
}
