package sync

import (
	"github.com/grigri201/prompt-vault/internal/interfaces"
)

// Re-export types from interfaces to maintain backward compatibility
type SyncStatus = interfaces.SyncStatus
type SyncProgress = interfaces.SyncProgress
type SyncDirection = interfaces.SyncDirection

// Re-export constants
const (
	SyncDirectionNone          = interfaces.SyncDirectionNone
	SyncDirectionDownload      = interfaces.SyncDirectionDownload
	SyncDirectionUpload        = interfaces.SyncDirectionUpload
	SyncDirectionBidirectional = interfaces.SyncDirectionBidirectional
)
