---
title: First Inventory
description: A short path from empty app to something useful.
---

The first good Stuff Stash session should end with one thing you can find again.

## 1. Sign In

Open the web app and sign in with SSO. In the local Compose stack, use the Dex
test account from [Run Stuff Stash](../self-hosting/).

## 2. Create An Inventory

Create an inventory for one real household area. Keep the scope small:

- `Household`
- `Garage`
- `Medicine`
- `Documents`
- `Tools`

Smaller inventories are easier to trust at the start.

## 3. Add One Location

Add the place where the item lives:

- `Garage`
- `Hall closet`
- `Office bins`
- `Medicine cabinet`

Locations can contain other locations, containers, and items. You do not need to
model the whole house before adding the first useful record.

## 4. Add One Item And Photo

Add the thing you are likely to look for later. Include a photo if it helps you
recognize it quickly.

Good first items:

- a tool in a storage bin,
- a medicine bottle,
- a document folder,
- a receipt,
- a cable or adapter you always lose.

## 5. Search For It

Search for the item by the word you would remember later. Stuff Stash should
match names, descriptions, custom fields, and useful attachment metadata.

## 6. Preview The Conversational Loop

The product center is conversational upkeep. Try the same task in that shape:

> Where is the fertilizer?

The useful answer is not a record ID. It is household language: garage shelf,
office bin, medicine cabinet, wire rack.

## 7. Know The Save Rule

For state-changing updates, the important rule is review before save:

> Move it from the garage shelf to the wire rack.

Stuff Stash should show the planned move before it changes the inventory. Approve
only when the source and destination look right.

## What You Should Have Afterward

- One inventory.
- One location.
- One item.
- A photo or useful detail.
- A search result that finds the item.
- A clear sense of how conversational review keeps the inventory current.

Next: learn the [core concepts](../concepts/).
