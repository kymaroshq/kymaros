/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package plugin provides public types and registration functions for
// external backup adapter and notifier plugins.
package plugin

import (
	"github.com/kymaroshq/kymaros/internal/adapter"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BackupAdapter is the interface that backup provider plugins must implement.
type BackupAdapter = adapter.BackupAdapter

// AdapterFactory creates a BackupAdapter for the given controller-runtime client.
type AdapterFactory = adapter.AdapterFactory

// BackupInfo contains metadata about a backup.
type BackupInfo = adapter.BackupInfo

// RestoreOptions defines parameters for a restore operation.
type RestoreOptions = adapter.RestoreOptions

// RestoreResult captures the outcome of a restore operation.
type RestoreResult = adapter.RestoreResult

// ScheduleInfo contains metadata about a backup schedule.
type ScheduleInfo = adapter.ScheduleInfo

// RegisterAdapter registers a backup adapter factory under the given name.
// Call this from an init() function to make the adapter available at runtime.
//
// Example:
//
//	func init() {
//	    plugin.RegisterAdapter("mybackup", func(c client.Client) plugin.BackupAdapter {
//	        return &MyAdapter{client: c}
//	    })
//	}
func RegisterAdapter(name string, factory func(c client.Client) BackupAdapter) {
	adapter.Register(name, factory)
}
