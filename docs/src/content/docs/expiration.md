---
title: Expiration Dates
description: Track package dates and choose your own reminders.
---

Track expiration dates for medicine, food, supplies, or any custom asset type.
Each item has one optional date. If two bottles expire at different times, keep
them as separate items; their tags and storage location can stay the same.

## Enable Tracking

Turn on **Expiration tracking** when creating or editing a custom asset type.
Choose that type when adding an item, then enter its expiration date. For an
existing item without a type, assign an enabled type in the item editor.

Use an exact date when the label includes a day. Choose **Month and year** for a
label such as February 2028. Stuff Stash preserves that precision: a month-only
date is tracked through the end of that month. Exact dates are tracked through
the end of the recorded day, using your reminder timezone.

You can edit or clear the date later. Turning off a type's tracking keeps its
recorded dates but stops reminders for that type. A missing date means no date
has been recorded.

## Choose Your Reminders

Open **Notifications** in inventory settings. Settings belong to you and the
current inventory; other members make their own choices.

- **Inventory defaults** control reminders for types that inherit them. Choose
  whether to receive upcoming and expired alerts, and how many days ahead to warn.
- **Asset type reminders** let you override those defaults for an individual
  type. Choose **Use inventory defaults** to remove an override.
- **Calendar timezone** determines when dates end and reminder windows begin.

The initial warning window is 30 days. Turning off the inventory default does
not turn off a type that has its own enabled override.

## Read Alerts

The notification bell opens your inbox on web and mobile. Its unread indicator
and read status stay in sync. Open an alert to see the item; its location trail
also helps you find where it belongs. Marking an alert read does not change the
item or its date.

Each recorded date can produce one upcoming alert and one expired alert.
Changing or clearing the date removes obsolete alerts. Already-expired items
produce an expired alert without sending an old upcoming alert as well.

Enable **Mobile push alerts** in the mobile app and allow notifications when
your device asks. Push also requires your server's delivery provider to be
[configured](../configuration/). Your inbox remains available if device alerts
are off. The web app does not request browser notification permission.

## Ask In Conversation

You can speak or type requests such as:

- “What medicine expires soon?”
- “Add aspirin to Bin 8, using the Medicine type, expiring February 2028.”
- “Change my Tylenol expiration to February 12, 2028.”
- “Remove the expiration date from my Tylenol.”

Review the item, destination, and date before approving a change. Give the month
name and year when possible. Location answers can also include an approaching
or passed expiration date. Typed requests receive written replies without speech.
