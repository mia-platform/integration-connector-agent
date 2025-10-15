// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package confluence

import (
	"github.com/gofiber/fiber/v2"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

const (
	confluenceEventHeader = "X-Event-Key"

	// Page events
	pageCreatedEvent = "page_created"
	pageUpdatedEvent = "page_updated"
	pageMovedEvent   = "page_moved"
	pageRemovedEvent = "page_removed"

	// Blog post events
	blogCreatedEvent = "blog_created"
	blogUpdatedEvent = "blog_updated"
	blogRemovedEvent = "blog_removed"

	// Comment events
	commentCreatedEvent = "comment_created"
	commentUpdatedEvent = "comment_updated"
	commentRemovedEvent = "comment_removed"

	// Attachment events
	attachmentCreatedEvent = "attachment_created"
	attachmentUpdatedEvent = "attachment_updated"
	attachmentRemovedEvent = "attachment_removed"

	// Space events
	spaceCreatedEvent = "space_created"
	spaceUpdatedEvent = "space_updated"
	spaceRemovedEvent = "space_removed"

	// User events
	userCreatedEvent     = "user_created"
	userUpdatedEvent     = "user_updated"
	userRemovedEvent     = "user_removed"
	userDeactivatedEvent = "user_deactivated"

	// Label events
	labelCreatedEvent = "label_created"
	labelRemovedEvent = "label_removed"

	// Like events
	likeCreatedEvent = "like_created"
	likeRemovedEvent = "like_removed"

	// Template events
	templateCreatedEvent = "template_created"
	templateUpdatedEvent = "template_updated"
	templateRemovedEvent = "template_removed"

	// Group events
	groupCreatedEvent = "group_created"
	groupRemovedEvent = "group_removed"

	// Application link events
	applicationLinkCreatedEvent = "applicationlink_created"
	applicationLinkUpdatedEvent = "applicationlink_updated"
	applicationLinkRemovedEvent = "applicationlink_removed"

	// Connect addon events
	connectAddonInstalledEvent   = "connect_addon_installed"
	connectAddonUninstalledEvent = "connect_addon_uninstalled"
	connectAddonEnabledEvent     = "connect_addon_enabled"
	connectAddonDisabledEvent    = "connect_addon_disabled"
)

var SupportedEvents = &webhook.Events{
	Supported: map[string]webhook.Event{
		// Page events
		pageCreatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("page.id"),
		},
		pageUpdatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("page.id"),
		},
		pageMovedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("page.id"),
		},
		pageRemovedEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("page.id"),
		},

		// Blog post events
		blogCreatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("page.id"),
		},
		blogUpdatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("page.id"),
		},
		blogRemovedEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("page.id"),
		},

		// Comment events
		commentCreatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("comment.id"),
		},
		commentUpdatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("comment.id"),
		},
		commentRemovedEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("comment.id"),
		},

		// Attachment events
		attachmentCreatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("attachment.id"),
		},
		attachmentUpdatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("attachment.id"),
		},
		attachmentRemovedEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("attachment.id"),
		},

		// Space events
		spaceCreatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("space.id"),
		},
		spaceUpdatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("space.id"),
		},
		spaceRemovedEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("space.id"),
		},

		// User events
		userCreatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("user.userKey"),
		},
		userUpdatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("user.userKey"),
		},
		userRemovedEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("user.userKey"),
		},
		userDeactivatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("user.userKey"),
		},

		// Label events
		labelCreatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("label.id"),
		},
		labelRemovedEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("label.id"),
		},

		// Like events
		likeCreatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("like.id"),
		},
		likeRemovedEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("like.id"),
		},

		// Template events
		templateCreatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("template.id"),
		},
		templateUpdatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("template.id"),
		},
		templateRemovedEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("template.id"),
		},

		// Group events
		groupCreatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("group.name"),
		},
		groupRemovedEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("group.name"),
		},

		// Application link events
		applicationLinkCreatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("applicationLink.id"),
		},
		applicationLinkUpdatedEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("applicationLink.id"),
		},
		applicationLinkRemovedEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("applicationLink.id"),
		},

		// Connect addon events
		connectAddonInstalledEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("addon.key"),
		},
		connectAddonUninstalledEvent: {
			Operation:  entities.Delete,
			GetFieldID: webhook.GetPrimaryKeyByPath("addon.key"),
		},
		connectAddonEnabledEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("addon.key"),
		},
		connectAddonDisabledEvent: {
			Operation:  entities.Write,
			GetFieldID: webhook.GetPrimaryKeyByPath("addon.key"),
		},
	},
	GetEventType: func(data webhook.EventTypeParam) string {
		return data.Headers.Get(confluenceEventHeader)
	},
	PayloadKey: webhook.ContentTypeConfig{
		fiber.MIMEApplicationForm: "payload",
	},
}
