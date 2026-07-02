package videos

import (
	"context"
	"file_share/internal/entity"
	"fmt"
)

func (s *Service) GetAll(ctx context.Context, query, rootFolderId, parentFolderId string, limit int, offset int) ([]entity.Video, error) {

	uLimit := uint64(limit)
	uOffset := uint64(offset)

	result, err := s.videoRepository.GetAll(ctx, query, rootFolderId, parentFolderId, uLimit, uOffset)
	if err != nil {
		return nil, fmt.Errorf("error getting videos")
	}

	return result, nil
}
