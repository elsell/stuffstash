package memory

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
	"time"
)

func (s *Store) BackfillThumbnailJobs(ctx context.Context, limit int, now time.Time) (ports.ThumbnailBackfillProgress, error) {
	if limit < 1 || limit > 1000 || now.IsZero() {
		return ports.ThumbnailBackfillProgress{}, errors.New("invalid thumbnail backfill batch")
	}
	if err := ctx.Err(); err != nil {
		return ports.ThumbnailBackfillProgress{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	progress := ports.ThumbnailBackfillProgress{Cursor: s.thumbnailBackfill.Cursor, Complete: s.thumbnailBackfill.Complete}
	if progress.Complete {
		return progress, nil
	}
	ids := []media.ID{}
	for id := range s.attachments {
		if id.String() > progress.Cursor.String() {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	progress.Complete = len(ids) < limit
	if len(ids) > limit {
		ids = ids[:limit]
	}
	jobs := []media.ThumbnailJob{}
	for _, id := range ids {
		attachment := s.attachments[id]
		if attachment.ContentType.IsImage() {
			job, err := media.NewThumbnailJob(attachment, media.ThumbnailJobBackfill, now)
			if err != nil {
				return ports.ThumbnailBackfillProgress{}, err
			}
			if _, exists := s.thumbnailJobs[thumbnailKey(job)]; !exists {
				jobs = append(jobs, job)
			}
		}
		progress.Scanned++
		progress.Cursor = id
	}
	if err := ctx.Err(); err != nil {
		return ports.ThumbnailBackfillProgress{}, err
	}
	for _, job := range jobs {
		s.enqueueThumbnailJob(&job)
		progress.Enqueued++
	}
	s.thumbnailBackfill = ports.ThumbnailBackfillProgress{Cursor: progress.Cursor, Complete: progress.Complete}
	return progress, nil
}
