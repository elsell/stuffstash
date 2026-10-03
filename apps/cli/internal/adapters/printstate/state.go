package printstate

// Store stores per-physical-device recovery evidence. Acquire is platform-specific
// because a nonblocking process lock is part of its safety contract.
type Store struct{ Directory string }
