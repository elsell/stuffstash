---
title: Ask About Your Inventory
description: Find items and review changes through a typed inventory conversation.
---

In the web workspace, choose **Ask Stuff Stash** in the header. The panel shows
which inventory you are using. Type a question or describe a change, then choose
**Send**. Enter sends; Shift+Enter starts a new line.

Answers may include links to items. Requests that change inventory data produce
a review first. Read the proposed changes, then choose **Approve changes** or
**Cancel changes**. Closing the panel does not approve anything.

You can request a name or description change, set or clear an existing custom
field, and correct an expiration date. Review lists the actual field values,
including cleared values. Fields you did not mention stay unchanged. If a field
name is ambiguous, clarify which one you mean; unknown fields are not created
silently.

Use **Stop** to cancel a request. If the connection ends while an approved change
is running, its outcome may be unknown. Refresh the inventory before trying the
change again; the app does not automatically repeat an approval.

Conversations stay within the selected inventory. Changing inventories or signing
out clears the panel. Transcripts are not saved in browser storage. This web
surface accepts typed messages; it does not request microphone access.

Your administrator must configure and test the conversation providers. See
[Compatible Language Providers](../compatible-providers/) for hosted and local
language endpoint setup. Existing speech-provider configuration is still required
by the shared realtime session service, even for typed sessions.
