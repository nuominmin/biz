package topnheap_test

import (
	"math/rand"
	"testing"

	"github.com/nuominmin/biz/queues/topnheap"
)

const maxHeapSize = 3
const maxDataSize = 10000000

type UserExp struct {
	UserId int64
	Exp    int64
}

func (u UserExp) Less(other topnheap.Item) bool {
	o := other.(UserExp)
	if u.Exp == o.Exp {
		return u.UserId > o.UserId
	}

	return u.Exp > o.Exp
}

func TestLargeData(t *testing.T) {
	h := topnheap.New[UserExp](maxHeapSize)

	for i := 0; i < maxDataSize; i++ {
		h.Add(UserExp{
			UserId: int64(i),
			Exp:    rand.Int63(),
		})
	}

	result := h.SortedDesc()

	if len(result) != maxHeapSize {
		t.Errorf("Expected %d elements in the heap, but got %d", maxHeapSize, len(result))
	}

	t.Logf("Top %d Exp values:", maxHeapSize)
	for _, item := range result {
		t.Log(item.Exp)
	}
}

// 基准测试
func BenchmarkTopNHeap(b *testing.B) {
	rnd := rand.New(rand.NewSource(1))
	for i := 0; i < b.N; i++ {
		h := topnheap.New[UserExp](maxHeapSize)
		for j := 0; j < maxDataSize; j++ {
			h.Add(UserExp{
				UserId: int64(j),
				Exp:    rnd.Int63(),
			})
		}
	}
}
