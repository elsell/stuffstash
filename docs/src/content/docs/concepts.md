---
title: Concepts
description: The small set of ideas behind Stuff Stash.
---

Stuff Stash uses a few domain terms. They exist to make household inventory
flexible without making the UI feel like warehouse software.

## Inventory

An inventory is a collection of stuff you manage together.

Examples:

- Household
- Garage
- Medicine
- Documents
- Tools

Inventories are the main unit for browsing, sharing, searching, and configuring
fields.

## Tenant

A tenant is the top-level security boundary. Most people can think of it as the
household or account space that owns one or more inventories.

Tenant language appears where security or setup needs it. Day to day, inventory
names should matter more.

## Item

An item is a thing you want to find, move, document, or share.

Examples:

- Fertilizer
- Passport folder
- Drill battery
- Aspirin

## Location

A location is a place-like record: garage, shelf, closet, cabinet, room, or bin
area.

Locations are backed by the same containment model as other assets, so they can
hold items, containers, and other locations.

## Container

A container is a movable thing that can hold other things.

Examples:

- Storage bin
- Toolbox
- Folder
- Parts organizer

This matters because a toolbox can move from the garage to the car without
losing the items inside it.

## Photos And Files

Photos help you recognize the thing. Files preserve receipts, documents, labels,
and other evidence.

Stuff Stash treats media as part of inventory, not decoration.

## Custom Types And Fields

Different kinds of stuff need different details. Medicine may need an expiration
date. Tools may need a battery type. Documents may need an issuer or renewal
date.

Custom types and fields let an inventory grow without forcing every item into
the same shape.

## Sharing

Inventories can be shared with scoped access. A viewer can see an inventory. An
editor can help maintain it. Sharing one inventory should not expose everything
else.

## Import And Export

Your inventory should be portable. Stuff Stash is designed around JSON and CSV
import/export boundaries so migration and backups can grow without trapping data
inside the app.
