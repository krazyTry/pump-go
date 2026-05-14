package helpers

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"reflect"

	binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func minBigInt(a, b *big.Int) *big.Int {
	if a.Cmp(b) <= 0 {
		return new(big.Int).Set(a)
	}
	return new(big.Int).Set(b)
}

// Filter represents a filter for querying accounts by owner and offset
type Filter struct {
	Owner  solana.PublicKey // Account owner to filter by
	Offset uint64           // Offset for pagination
}

func discriminator(name string) []byte {
	hash := sha256.Sum256([]byte("account:" + name))
	var out [8]byte
	copy(out[:], hash[:8])
	return out[:]
}

// ComputeStructOffset gets the offset position of an object in a struct
func ComputeStructOffset(x any, o string) uint64 {
	t := reflect.TypeOf(x).Elem()
	fields := make([]reflect.StructField, 0)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name == o {
			break
		}
		fields = append(fields, f)
	}

	newType := reflect.StructOf(fields)
	newValue := reflect.New(newType).Elem()

	buf__ := new(bytes.Buffer)
	enc__ := binary.NewBorshEncoder(buf__)
	enc__.Encode(newValue.Interface())

	// instruction discriminators offset = 8
	return uint64(buf__.Len()) + 8
}

func CreateProgramAccountFilter(key string, filter *Filter) []rpc.RPCFilter {
	var filters []rpc.RPCFilter
	filters = append(filters, rpc.RPCFilter{
		Memcmp: &rpc.RPCFilterMemcmp{
			Offset: 0,
			Bytes:  discriminator(key),
		},
	})

	if filter != nil {
		filters = append(filters, rpc.RPCFilter{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: filter.Offset,
				Bytes:  filter.Owner[:],
			},
		})
	}

	return filters
}
