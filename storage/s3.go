package storage

import (
	"github.com/test-network-function/collector/types"
)

type S3Storage struct {
}

// constructor
func NewS3Storage() *S3Storage {
	return &S3Storage{}
}

func (s *S3Storage) Get() *types.Claim {
	return &types.Claim{}
}
