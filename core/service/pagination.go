package service

import (
	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/proto"
)

func clampPageSize(size, fallback, max int32) int32 {
	switch {
	case size <= 0:
		return fallback
	case size > max:
		return max
	default:
		return size
	}
}

func pageOffset(req pagination.Request) (int, error) {
	token, err := pagination.ParsePageToken(req)
	if err != nil {
		return 0, err
	}
	return int(token.Offset), nil
}

func encodeNextPageToken(req pagination.Request, pageSize int32, hasNext bool) string {
	if !hasNext {
		return ""
	}
	cloned, ok := proto.Clone(req).(pagination.Request)
	if !ok {
		return ""
	}
	token, err := pagination.ParsePageToken(cloned)
	if err != nil {
		return ""
	}
	return token.Next(sizeRequest{Request: cloned, size: pageSize}).String()
}

type sizeRequest struct {
	pagination.Request
	size int32
}

func (r sizeRequest) GetPageSize() int32 { return r.size }
