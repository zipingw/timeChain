package bptreeS

import (
	"encoding/json"
	"math/rand"
	"testing"
	"strconv"
)

func TestBPT(t *testing.T) {
	bpt := NewBPTree(6)

	bpt.Set("10", 10)
	bpt.Set("23", 23)
	bpt.Set("33", 33)
	bpt.Set("35", 35)
	bpt.Set("15", 15)
	bpt.Set("16", 16)
	bpt.Set("17", 17)
	bpt.Set("19", 19)
	bpt.Set("20", 20)

	bpt.Remove("23")
	bpt.Remove("33")
	bpt.Remove("10")
	
	t.Log(bpt.Get("10"))
	t.Log(bpt.Get("15"))
	t.Log(bpt.Get("20"))

	data, _ := json.MarshalIndent(bpt.GetData(), "", "    ")
	t.Log(string(data))
}

func TestBPTRand(t *testing.T) {
	bpt := NewBPTree(4)

	for i:=0; i < 20; i++ {
		key := strconv.Itoa(rand.Int() % 20 +1)
		t.Log(key)
		bpt.Set(key, key)
	}

	data, _ := json.MarshalIndent(bpt.GetData(), "", "    ")
	t.Log(string(data))
}
func TestBPTRandTime(t *testing.T) {
	bpt := NewBPTree(10)
	bpt.Set("202312270000", 0)
	bpt.Set("202312270100", 1)
	bpt.Set("202312270800", 8)
	bpt.Set("202312271200", 12)
	bpt.Set("202312271500", 15)
	bpt.Set("202312272300", 23)
	bpt.Set("202312280200", 2)
	bpt.Set("202312290200", 2)

	data, _ := json.MarshalIndent(bpt.GetData(), "", "    ")
	t.Log(string(data))
	t.Log(bpt.Get("202312271200"))
}
