/** Complete workflow messages, including variable-derived notices. */
export const workflowMessages = {
  "voice.openEntityIn": "Open {title} in {context}",
  "voice.openEntity": "Open {title}",
  "voice.entityPosition": "{label} ({position} of {total})",
  "move.cannotContainSelf": "An asset cannot be moved into itself.",
  "move.saved": "Moved {title}.",
  "move.savedAtRoot": "Moved {title} to No parent.",
  "voice.disconnectedReview": "{reason} This review is disconnected. Your draft is kept here. Check the inventory before submitting the change again.",
  "color.degrees": "{value} degrees",
  "color.percent": "{value} percent",
  "onboarding.householdExample": "e.g. Maple Street household",
  "onboarding.inventoryExample": "e.g. Home Inventory",
  "assets.openNamed": "Open asset {title}",
  "customization.onlyIn": "Only in {inventory}",
  "settings.unavailableReason": "Settings unavailable. {reason}",
  "browse.currentInventory": "this inventory",
  "move.promotedPath": "{path} · Will become a container",
  "assets.savedWithPhotoFailures": {
    "one": "Saved {title}, but {count} photo upload failed.",
    "other": "Saved {title}, but {count} photo uploads failed."
  },
  "assets.savedNamed": "Saved {title}.",
  "assets.updatedNamed": "Updated {title}.",
  "customization.inheritedName": "{name} · Inherited",
  "photos.added": {
    "one": "{count} photo added.",
    "other": "{count} photos added."
  },
  "photos.failedReason": "Photos could not be uploaded: {reason}",
  "photos.partiallyAdded": "{added} of {total} photos added.",
  "workspace.inventoryCreated": "Created {title}.",
  "assets.archivedNamed": "Archived {title}.",
  "assets.restoredNamed": "Restored {title}.",
  "assets.deletedNamed": "Deleted {title}.",
  "assets.checkedOutNamed": "Checked out {title}.",
  "assets.returnedNamed": "Returned {title}.",
  "photos.archivedNamed": "Archived {title}.",
  "photos.deletedNamed": "Deleted {title}.",
  "photos.uploadedNamed": "Uploaded {title}.",
  "move.savedInto": "Moved {title} into {parent}.",
  "move.failedWithReason": "Move not saved. {title} stayed where it was. {reason}",
  "move.permissionDenied": "Move not saved. You do not have permission to move assets in this inventory.",
  "move.failed": "Move not saved. {title} stayed where it was.",
  "contents.insideNamed": "Inside {title}",
  "contents.spacesIn": "Spaces in {title}",
  "contents.itemsIn": "Items in {title}",
  "search.noSuggestions": "No suggestions for \"{query}\". Press Search to run a full search.",
  "move.currentDestination": "Current destination: {name}, {metadata}",
  "conversation.caseRevisionSaved": "Test case revision {revision} saved.",
  "conversation.workflowRevisionSaved": "Draft revision {revision} saved. Run test cases before activation.",
  "workspace.appliedRefreshNeeded": "{result} Reload to see the latest inventory.",
  "move.matchCount": {
    "one": "{count} match",
    "other": "{count} matches"
  },
  "move.destinationCount": {
    "one": "{count} possible destination",
    "other": "{count} possible destinations"
  },
  "move.suggestionCount": {
    "one": "Showing {count} suggested destination.",
    "other": "Showing {count} suggested destinations."
  },
  "move.truncatedMatches": "Showing the first {shown} of {total} matches."
} as const;
