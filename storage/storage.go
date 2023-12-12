package storage

import "github.com/test-network-function/collector/types"

type Storage interface {
	Get() *types.Claim
}
