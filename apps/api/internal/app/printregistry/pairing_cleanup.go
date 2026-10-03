package printregistry

import "context"

func (s ConnectorService) CleanupExpiredPairings(ctx context.Context, limit int) (int, error) {
	return s.Repository.CleanupExpiredPrintPairings(ctx, s.Registry.Clock.Now(), limit)
}
