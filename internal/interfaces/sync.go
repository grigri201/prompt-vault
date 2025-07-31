package interfaces

import (
	"context"
	"fmt"
	"time"
)

// SyncProgress tracks synchronization progress
type SyncProgress struct {
	Total      int  `json:"total"`
	Completed  int  `json:"completed"`
	InProgress bool `json:"in_progress"`
}

// SyncDirection indicates the direction of synchronization
type SyncDirection int

const (
	SyncDirectionNone     SyncDirection = iota
	SyncDirectionDownload               // Remote to local
	SyncDirectionUpload                 // Local to remote
	SyncDirectionBidirectional
)

// String returns a string representation of the sync direction
func (d SyncDirection) String() string {
	switch d {
	case SyncDirectionDownload:
		return "下载"
	case SyncDirectionUpload:
		return "上传"
	case SyncDirectionBidirectional:
		return "双向"
	default:
		return "无需同步"
	}
}

// SyncStatus represents the current synchronization status
type SyncStatus struct {
	LocalTime  time.Time     `json:"local_time"`
	RemoteTime time.Time     `json:"remote_time"`
	NeedsSync  bool          `json:"needs_sync"`
	Direction  SyncDirection `json:"direction"`
	Progress   SyncProgress  `json:"progress"`
}

// DisplayString returns a user-friendly display string for the sync status
func (s SyncStatus) DisplayString() string {
	if !s.NeedsSync {
		return "数据已同步"
	}

	progress := ""
	if s.Progress.InProgress {
		progress = fmt.Sprintf(" (%d/%d)", s.Progress.Completed, s.Progress.Total)
	}

	return fmt.Sprintf("需要%s%s", s.Direction.String(), progress)
}

// SyncManager defines the interface for synchronization management
type SyncManager interface {
	Manager
	SynchronizeData(ctx context.Context) error
	GetSyncStatus() SyncStatus
}

// SyncMiddleware defines the interface for synchronization middleware
type SyncMiddleware interface {
	PreSync(ctx context.Context, cmd string) error
	PostSync(ctx context.Context, cmd string) error
}
