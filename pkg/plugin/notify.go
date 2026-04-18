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

package plugin

import "github.com/kymaroshq/kymaros/internal/notify"

// Notifier is the interface that notification plugins must implement.
type Notifier = notify.Notifier

// NotifierFactory creates a Notifier instance.
type NotifierFactory = notify.NotifierFactory

// Notification contains the data passed to a Notifier.
type Notification = notify.Notification

// RegisterNotifier registers a notifier factory under the given name.
// Call this from an init() function to make the notifier available at runtime.
func RegisterNotifier(name string, factory func() Notifier) {
	notify.RegisterNotifier(name, factory)
}
