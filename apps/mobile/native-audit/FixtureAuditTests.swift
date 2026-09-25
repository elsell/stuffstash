import XCTest
import UIKit

final class FixtureAuditTests: XCTestCase {
  private let app = XCUIApplication(bundleIdentifier: "org.stuffstash.mobile")
  override func setUpWithError() throws {
    continueAfterFailure = false
    app.launch()
    let entry = (name.contains("testHomeCollectionsReplaceBrowseRefinementsAndRetainTabs")
      || name.contains("testHistoryJourneyClearsPersistentChrome")
      || name.contains("testVoiceAccessorySettledNavigationAppearance")
      || name.contains("testInventoryCollectionClearsPersistentChrome")
      || name.contains("testCustomizationCollectionClearsPersistentChrome")
      || name.contains("testNotificationJourneyRetainsTabsAndClearsFooter"))
      ? "View all recently changed assets" : "Audit Browse filters"
    XCTAssertTrue(app.buttons[entry].waitForExistence(timeout: 30))
    let providerOmitted = app.otherElements["audit-keyboard-provider-omitted"].exists
    let providerEvidence = XCTAttachment(string: "Keyboard provider omitted: \(providerOmitted)")
    providerEvidence.name = "keyboard-provider-configuration"
    providerEvidence.lifetime = .keepAlways
    add(providerEvidence)
  }
  override func tearDownWithError() throws {
    capture("final-state")
    app.terminate()
  }
  func testNativeMenuLocksOpenActionsAndRecovers() {
    let entry = app.buttons["Audit menu ownership"].firstMatch
    for _ in 0..<12 where !entry.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(entry.isHittable); entry.tap()
    let trigger = app.buttons["Menu ownership actions"].firstMatch
    XCTAssertTrue(trigger.waitForExistence(timeout: 5))
    app.buttons["Lock menu shortly"].tap()
    trigger.tap()
    let command = app.buttons["Run menu command"].firstMatch
    XCTAssertTrue(command.waitForExistence(timeout: 3))
    XCTAssertTrue(app.staticTexts["Menu locked"].waitForExistence(timeout: 10))
    let locked = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      !command.exists || !command.isEnabled
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [locked], timeout: 5), .completed)
    if command.exists {
      XCTAssertFalse(command.isEnabled)
      command.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5)).tap()
    }
    XCTAssertTrue(app.staticTexts["Menu activations: 0"].exists)
    capture("menu-owner-locked")
    // Tapping outside an open native menu dismisses it without choosing an item.
    if command.exists { app.navigationBars["Menu ownership"].tap() }
    XCTAssertTrue(command.waitForNonExistence(timeout: 5))
    let unlock = app.buttons["Unlock menu"].firstMatch
    XCTAssertTrue(unlock.isHittable); unlock.tap()
    XCTAssertTrue(app.staticTexts["Menu ready"].waitForExistence(timeout: 5))
    XCTAssertFalse(command.exists, "Unlock must not reopen a dismissed menu")
    trigger.tap()
    XCTAssertTrue(command.waitForExistence(timeout: 5)); XCTAssertTrue(command.isEnabled)
    command.tap()
    XCTAssertTrue(app.staticTexts["Menu activations: 1"].waitForExistence(timeout: 5))
    capture("menu-owner-recovered")
  }

  func testNotificationInboxReadStateAndNavigationReturn() {
    let entry = app.buttons["Audit Notifications"].firstMatch
    for _ in 0..<12 where !entry.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(entry.isHittable); entry.tap()
    let title = "Household medicine with a long descriptive label"
    let markRead = app.buttons["Mark \(title) read"].firstMatch
    XCTAssertTrue(markRead.waitForExistence(timeout: 10))
    XCTAssertTrue(markRead.isHittable)
    XCTAssertGreaterThan(markRead.frame.width, 0)
    XCTAssertGreaterThan(markRead.frame.height, 0)
    XCTAssertTrue(app.frame.contains(markRead.frame))
    capture("notification-inbox-long-row")
    markRead.tap()
    let markUnread = app.buttons["Mark \(title) unread"].firstMatch
    XCTAssertTrue(markUnread.waitForExistence(timeout: 5)); markUnread.tap()
    XCTAssertTrue(markRead.waitForExistence(timeout: 5))
    app.buttons["Open location Cold and cough supplies"].firstMatch.tap()
    XCTAssertTrue(app.staticTexts["Resolved target: box"].waitForExistence(timeout: 5))
    app.navigationBars["Notification destination"].buttons.firstMatch.tap()
    XCTAssertTrue(markRead.waitForExistence(timeout: 5))
    app.buttons["Open \(title)"].firstMatch.tap()
    XCTAssertTrue(app.staticTexts["Resolved target: medicine-item"].waitForExistence(timeout: 5))
    app.navigationBars["Notification destination"].buttons.firstMatch.tap()
    XCTAssertTrue(markUnread.waitForExistence(timeout: 5))
    app.buttons["Reminder settings"].firstMatch.tap()
    XCTAssertTrue(app.staticTexts["Resolved target: Reminder settings"].waitForExistence(timeout: 5))
    app.navigationBars["Notification destination"].buttons.firstMatch.tap()
    XCTAssertTrue(markUnread.waitForExistence(timeout: 5))
    app.buttons["Mark Emergency batteries unread"].firstMatch.tap()
    let markAll = app.buttons["Mark all read"].firstMatch
    XCTAssertTrue(app.buttons["Mark Emergency batteries read"].firstMatch.waitForExistence(timeout: 5))
    expectation(for: NSPredicate(format: "enabled == true"), evaluatedWith: markAll)
    waitForExpectations(timeout: 5); markAll.tap()
    XCTAssertTrue(app.buttons["Mark Emergency batteries unread"].firstMatch.waitForExistence(timeout: 5))
    let unreadFilter = app.buttons["Unread"].firstMatch
    expectation(for: NSPredicate(format: "enabled == true"), evaluatedWith: unreadFilter)
    waitForExpectations(timeout: 5); unreadFilter.tap()
    XCTAssertTrue(app.staticTexts["No unread notifications."].waitForExistence(timeout: 5))
    capture("notification-inbox-unread-empty")
  }
  func testNotificationReadControlDeliveredTouchRegion() {
    let entry = app.buttons["Audit Notifications"].firstMatch
    for _ in 0..<12 where !entry.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(entry.isHittable); entry.tap()
    let title = "Household medicine with a long descriptive label"
    let offsets: [(CGFloat, CGFloat)] = [(0, 0), (0, -21), (0, 21), (-21, 0), (21, 0),
                                          (-21, -21), (21, -21), (-21, 21), (21, 21)]
    var read = false
    for (index, offset) in offsets.enumerated() {
      let current = app.buttons["Mark \(title) \(read ? "unread" : "read")"].firstMatch
      XCTAssertTrue(current.waitForExistence(timeout: 5))
      let ready = XCTNSPredicateExpectation(predicate: NSPredicate(format: "enabled == true"), object: current)
      XCTAssertEqual(XCTWaiter.wait(for: [ready], timeout: 5), .completed)
      XCTAssertTrue(current.isHittable)
      let center = current.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
      center.withOffset(CGVector(dx: offset.0, dy: offset.1)).tap()
      let changed = app.buttons["Mark \(title) \(read ? "read" : "unread")"].firstMatch
      guard changed.waitForExistence(timeout: 5) else {
        capture("notification-read-hit-region-failed-\(index)")
        XCTFail("Read command did not toggle at probe \(index): \(offset)")
        return
      }
      XCTAssertTrue(current.waitForNonExistence(timeout: 5))
      XCTAssertTrue(app.navigationBars["Notifications"].exists)
      read.toggle()
    }
    capture("notification-read-delivered-hit-region")
  }
  func testInvitationAcceptanceRetainsAccessAfterOpeningFailure() {
    func openInvitation() {
      let entry = app.buttons["Audit invitation acceptance"].firstMatch
      for _ in 0..<12 where !entry.isHittable { app.scrollViews.firstMatch.swipeUp() }
      XCTAssertTrue(entry.isHittable)
      entry.tap()
      XCTAssertTrue(app.navigationBars["Invitation"].waitForExistence(timeout: 10))
      XCTAssertTrue(app.buttons["Join inventory"].waitForExistence(timeout: 5))
    }
    func visible(_ element: XCUIElement) -> Bool {
      guard element.exists else { return false }
      let viewport = app.scrollViews.firstMatch.frame.intersection(app.frame)
      let frame = element.frame
      return frame.width > 0 && frame.height > 0 &&
        frame.minY >= max(viewport.minY, app.navigationBars["Invitation"].frame.maxY) &&
        frame.maxY <= viewport.maxY && frame.minX >= viewport.minX && frame.maxX <= viewport.maxX
    }
    func reveal(_ element: XCUIElement, requiresHit: Bool = true) {
      for _ in 0..<8 where !visible(element) {
        if element.exists && element.frame.minY < app.navigationBars["Invitation"].frame.maxY {
          app.scrollViews.firstMatch.swipeDown()
        } else {
          app.scrollViews.firstMatch.swipeUp()
        }
      }
      XCTAssertTrue(visible(element), "Invitation content must fit below navigation and within the scroll viewport")
      if requiresHit { XCTAssertTrue(element.isHittable) }
    }
    openInvitation()
    let name = app.staticTexts["Family camping equipment and seasonal supplies shared with the household"].firstMatch
    reveal(name, requiresHit: false)
    reveal(app.staticTexts["Editor"].firstMatch, requiresHit: false)
    capture("invitation-review-normal")
    let later = app.buttons["Not now"].firstMatch
    reveal(later)
    later.tap()
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
    openInvitation()
    let join = app.buttons["Join inventory"].firstMatch
    reveal(join)
    join.tap()
    let open = app.buttons["Open inventory"].firstMatch
    XCTAssertTrue(open.waitForExistence(timeout: 5))
    reveal(open)
    open.tap()
    let recovery = app.staticTexts["The inventory could not be opened. Your access was still added."].firstMatch
    XCTAssertTrue(recovery.waitForExistence(timeout: 5))
    XCTAssertFalse(app.buttons["Join inventory"].exists)
    reveal(recovery, requiresHit: false)
    capture("invitation-open-recovery")
    reveal(open)
    open.tap()
    XCTAssertTrue(app.staticTexts["Opened invitation inventory; accepted once"].waitForExistence(timeout: 5))
  }
  private func verifyNoticePlacement(_ presentation: String) {
    let open = app.buttons["Audit Notice \(presentation)"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    let header = app.navigationBars["Notice placement"]
    XCTAssertTrue(header.waitForExistence(timeout: 10))
    app.buttons["Show placement notice"].tap()
    let dismiss = app.buttons["Audit notice. A retained action must leave navigation reachable. Dismiss message"]
    let action = app.buttons["Complete audit action"]
    XCTAssertTrue(dismiss.waitForExistence(timeout: 5))
    XCTAssertTrue(action.waitForExistence(timeout: 5))
    let notice = app.descendants(matching: .any).matching(identifier: "app-notice-container").firstMatch
    let content = app.descendants(matching: .any).matching(identifier: "notice-placement-content").firstMatch
    XCTAssertTrue(notice.exists); XCTAssertTrue(content.exists)
    var lastGeometry = "No geometry sample evaluated"
    func descendants(_ snapshot: any XCUIElementSnapshot) -> [any XCUIElementSnapshot] {
      [snapshot] + snapshot.children.flatMap { descendants($0) }
    }
    func belowNavigation(_ control: any XCUIElementSnapshot, bounds: CGRect, headerBottom: CGFloat) -> Bool {
      let rect = control.frame
      let contained = rect.width > 0 && rect.height > 0 && rect.minY >= max(bounds.minY, headerBottom) &&
        rect.maxY <= bounds.maxY && rect.minX >= bounds.minX && rect.maxX <= bounds.maxX
      lastGeometry += "\n\(control.identifier.isEmpty ? control.label : control.identifier): \(rect), contained=\(contained)"
      return contained
    }
    let settled = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      do {
        let snapshot = try self.app.snapshot()
        let elements = descendants(snapshot)
        guard let contentSnapshot = elements.first(where: { $0.identifier == "notice-placement-content" }),
              let headerSnapshot = elements.first(where: { $0.elementType == .navigationBar && $0.identifier == "Notice placement" }),
              let noticeSnapshot = elements.first(where: { $0.identifier == "app-notice-container" }),
              let dismissSnapshot = elements.first(where: { $0.elementType == .button && $0.label == "Audit notice. A retained action must leave navigation reachable. Dismiss message" }),
              let actionSnapshot = elements.first(where: { $0.elementType == .button && $0.label == "Complete audit action" }) else {
          lastGeometry = "Required notice geometry element missing from snapshot"
          return false
        }
        let bounds = contentSnapshot.frame.intersection(snapshot.frame)
        lastGeometry = "app=\(snapshot.frame), content=\(contentSnapshot.frame), header=\(headerSnapshot.frame)"
        guard !bounds.isEmpty, !bounds.isNull, !headerSnapshot.frame.isEmpty else { return false }
        let results = [noticeSnapshot, dismissSnapshot, actionSnapshot].map {
          belowNavigation($0, bounds: bounds, headerBottom: headerSnapshot.frame.maxY)
        }
        return results.allSatisfy { $0 }
      } catch {
        lastGeometry = "Could not capture notice geometry: \(error)"
        return false
      }
    }, object: nil)
    let placement = XCTWaiter.wait(for: [settled], timeout: 5)
    capture("notice-\(presentation)-placement")
    XCTAssertEqual(placement, .completed, "Full notice must fit in the active screen below navigation. Last sample: \(lastGeometry)")
    XCTAssertTrue(dismiss.isHittable); XCTAssertTrue(action.isHittable)
    let back = presentation == "sheet" ? header.buttons["Close notice fixture"] : header.buttons["BackButton"].firstMatch
    XCTAssertTrue(back.isHittable)
    action.tap()
    XCTAssertTrue(app.staticTexts["Notice actions completed: 1"].waitForExistence(timeout: 5))
    app.buttons["Show placement notice"].tap()
    XCTAssertTrue(dismiss.waitForExistence(timeout: 5)); dismiss.tap()
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: dismiss)], timeout: 5), .completed)
    XCTAssertTrue(app.staticTexts["Notice actions completed: 1"].exists)
    XCTAssertTrue(back.isHittable); back.tap()
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
  }

  func testNoticeKeepsPushedNavigationReachable() { verifyNoticePlacement("push") }
  func testNoticeKeepsSheetNavigationReachable() { verifyNoticePlacement("sheet") }

  private func verifyProviderEditor(_ kind: String, discard: Bool = false) {
    let open = app.buttons["Audit Provider \(kind)"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    let header = app.navigationBars["Provider editor"]
    XCTAssertTrue(header.waitForExistence(timeout: 10))
    let save = header.buttons[kind == "prompt" ? "Save Guidance" : "Save Credential"]
    XCTAssertTrue(save.waitForExistence(timeout: 5)); XCTAssertFalse(save.isEnabled)
    let field = kind == "prompt" ? app.textViews["New prompt guidance"] : app.secureTextFields["API key"]
    XCTAssertTrue(field.waitForExistence(timeout: 5)); XCTAssertTrue(field.isHittable)
    field.tap(); waitForKeyboard(); field.typeText("Synthetic replacement")
    XCTAssertTrue(save.isEnabled); XCTAssertTrue(save.isHittable)
    if kind == "prompt" { XCTAssertEqual(field.value as? String, "Synthetic replacement") }
    capture("provider-\(kind)-draft")
    let back = header.buttons["BackButton"].firstMatch
    XCTAssertTrue(back.isHittable); back.tap()
    let alert = app.alerts["Discard changes?"]
    XCTAssertTrue(alert.waitForExistence(timeout: 5))
    if discard {
      alert.buttons["Discard"].tap()
    } else {
      alert.buttons["Keep Editing"].tap()
      XCTAssertTrue(field.exists); XCTAssertTrue(save.isEnabled)
      save.tap()
      let error = app.staticTexts.matching(identifier: "Audit replacement unavailable. Try again.").firstMatch
      XCTAssertTrue(error.waitForExistence(timeout: 5))
      XCTAssertTrue(header.exists); XCTAssertTrue(save.isEnabled)
      if kind == "prompt" { XCTAssertEqual(field.value as? String, "Synthetic replacement") }
      if app.keyboards.firstMatch.exists {
        let dismiss = app.buttons["Dismiss keyboard"].firstMatch
        XCTAssertTrue(dismiss.isHittable); dismiss.tap()
        XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: app.keyboards.firstMatch)], timeout: 5), .completed)
      }
      let form = app.scrollViews.containing(.staticText, identifier: "Audit replacement unavailable. Try again.").firstMatch
      XCTAssertTrue(form.exists)
      func errorVisible() -> Bool {
        let bounds = form.frame.intersection(app.frame)
        let rect = error.frame
        return rect.height > 0 && rect.minY >= max(bounds.minY, header.frame.maxY) &&
          rect.maxY <= bounds.maxY && rect.minX >= bounds.minX && rect.maxX <= bounds.maxX
      }
      for _ in 0..<8 where !errorVisible() {
        if error.frame.minY < header.frame.maxY { form.swipeDown() } else { form.swipeUp() }
      }
      XCTAssertTrue(errorVisible()); XCTAssertTrue(save.isHittable); XCTAssertTrue(back.isHittable)
      capture("provider-\(kind)-failed-save")
      save.tap()
    }
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
    XCTAssertFalse(app.alerts["Discard changes?"].exists)
  }

  func testProviderCredentialNativeSaveRecovery() { verifyProviderEditor("credential") }
  func testProviderPromptNativeSaveRecovery() { verifyProviderEditor("prompt") }
  func testProviderPromptNativeDiscard() { verifyProviderEditor("prompt", discard: true) }

  func testSharingRecoveryKeepsHeaderAndCommandsReachable() {
    let open = app.buttons["Audit Sharing"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    let header = app.navigationBars["Sharing"]
    XCTAssertTrue(header.waitForExistence(timeout: 10))
    let form = app.scrollViews.firstMatch
    func inContent(_ element: XCUIElement) -> Bool {
      let bounds = form.frame.intersection(app.frame)
      let rect = element.frame
      return rect.height > 0 && rect.minY >= max(bounds.minY, header.frame.maxY) &&
        rect.maxY <= bounds.maxY && rect.minX >= bounds.minX && rect.maxX <= bounds.maxX
    }
    func reveal(_ element: XCUIElement, interactive: Bool = true) {
      XCTAssertTrue(element.waitForExistence(timeout: 5))
      for _ in 0..<18 {
        let bounds = form.frame.intersection(app.frame)
        let top = max(bounds.minY, header.frame.maxY)
        let rect = element.frame
        if inContent(element) && (!interactive || element.isHittable) { return }
        let above = rect.minY < top
        form.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.4 : 0.7))
          .press(forDuration: 0.05, thenDragTo: form.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.7 : 0.4)))
      }
      XCTFail("Sharing element must be fully within the content viewport below navigation")
    }
    func feedback(_ title: String, message: String? = nil, captureName: String) {
      let heading = app.staticTexts[title].firstMatch
      reveal(heading, interactive: false)
      if let message {
        let body = app.staticTexts[message].firstMatch
        reveal(body, interactive: false)
        XCTAssertTrue(inContent(heading) && inContent(body), "Normal-text feedback title and recovery must share the visible content viewport")
      }
      XCTAssertTrue(header.buttons["BackButton"].firstMatch.isHittable)
      capture(captureName)
    }
    let email = app.textFields["Invitee email"]
    reveal(email); email.tap(); waitForKeyboard()
    email.typeText("audit@example.invalid")
    XCTAssertEqual(email.value as? String, "audit@example.invalid")
    let dismiss = app.buttons["Dismiss keyboard"].firstMatch
    XCTAssertTrue(dismiss.waitForExistence(timeout: 5)); dismiss.tap()
    let create = app.buttons["Create Invitation"].firstMatch
    func assertCreationContained() {
      let creationForm = app.otherElements["invitation-creation-form"].firstMatch
      XCTAssertTrue(creationForm.exists)
      let bounds = creationForm.frame
      XCTAssertFalse(bounds.isEmpty)
      XCTAssertTrue(bounds.contains(create.frame), "The complete native action must fit its form")
      XCTAssertGreaterThanOrEqual(bounds.maxY - create.frame.maxY, 15, "Retain the form's bottom inset")
    }
    reveal(create); assertCreationContained(); create.tap()
    feedback("Invitation created, link unavailable", message: "If the invitation is still pending below, cancel it before retrying. If you already cancelled it, try again.", captureName: "sharing-unavailable-link")
    assertCreationContained()
    XCTAssertEqual(email.value as? String, "audit@example.invalid")
    XCTAssertFalse(app.staticTexts["Complete invitation link"].exists)

    let cancel = app.buttons["Cancel invitation"].firstMatch
    reveal(cancel)
    XCTAssertGreaterThanOrEqual(cancel.frame.width, 44)
    XCTAssertGreaterThanOrEqual(cancel.frame.height, 44)
    XCTAssertFalse(app.keyboards.firstMatch.exists, "Submitting an invitation must end keyboard editing")
    cancel.tap()
    let confirm = app.alerts.buttons["Cancel Invitation"]
    XCTAssertTrue(confirm.waitForExistence(timeout: 5)); confirm.tap()
    feedback("Could not cancel invitation", message: "Audit cancellation unavailable. Try again.", captureName: "sharing-cancel-recovery")
    reveal(cancel); cancel.tap()
    XCTAssertTrue(confirm.waitForExistence(timeout: 5)); confirm.tap()
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: cancel)], timeout: 5), .completed)
    XCTAssertEqual(email.value as? String, "audit@example.invalid")
    XCTAssertTrue(app.staticTexts["If the invitation is still pending below, cancel it before retrying. If you already cancelled it, try again."].exists)

    reveal(create); create.tap()
    let link = app.staticTexts["Complete invitation link"].firstMatch
    reveal(link, interactive: false)
    let copy = app.buttons["Copy link"].firstMatch
    reveal(copy); copy.tap()
    feedback("Could not copy invitation", message: "Audit copy unavailable. Try again.", captureName: "sharing-copy-recovery")
    reveal(copy); copy.tap()
    feedback("Invitation link copied", captureName: "sharing-copy-complete")
    let share = app.buttons["Share invitation"].firstMatch
    reveal(share); share.tap()
    feedback("Could not share invitation", message: "Audit sharing does not open external destinations.", captureName: "sharing-share-recovery")
    header.buttons["BackButton"].firstMatch.tap()
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
  }

  func testFooterAppearanceAndDisabledActions() {
    auditFooterAppearance(largeText: false)
  }

  func testFooterAppearanceAtAccessibilityTextSize() {
    auditFooterAppearance(largeText: true)
  }

  private func auditFooterAppearance(largeText: Bool) {
    if largeText {
      app.terminate()
      app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
      app.launch()
      XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
    }
    guard openFixtureURL("audit-footer-appearance") else { return }
    XCTAssertTrue(app.staticTexts["Footer appearance"].waitForExistence(timeout: 10))
    let form = app.scrollViews.containing(.staticText, identifier: "Footer appearance").firstMatch
    let root = app.otherElements["footer-appearance-actions"].firstMatch
    XCTAssertTrue(root.exists)
    func reveal(_ element: XCUIElement) {
      for _ in 0..<18 {
        let visible = form.frame.intersection(app.frame)
        let bounds = CGRect(x: visible.minX, y: visible.minY, width: visible.width,
          height: max(0, min(visible.maxY, root.frame.minY) - visible.minY))
        if element.isHittable && element.frame.minY >= bounds.minY && element.frame.maxY <= bounds.maxY { return }
        let above = element.frame.minY < bounds.minY
        form.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.4 : 0.7))
          .press(forDuration: 0.05, thenDragTo: form.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.7 : 0.4)))
      }
      XCTFail("Footer diagnostic control must be fully visible")
    }
    func footerVisible() {
      let sheetBounds = form.frame.intersection(app.frame)
      XCTAssertFalse(sheetBounds.isEmpty)
      XCTAssertGreaterThanOrEqual(root.frame.minY, sheetBounds.minY)
      XCTAssertLessThanOrEqual(root.frame.maxY, sheetBounds.maxY)
      XCTAssertGreaterThanOrEqual(root.frame.minX, sheetBounds.minX)
      XCTAssertLessThanOrEqual(root.frame.maxX, sheetBounds.maxX)
      let bounds = root.frame.intersection(sheetBounds)
      XCTAssertFalse(bounds.isEmpty)
      for label in ["Move", "Cancel"] {
        let button = app.buttons[label].firstMatch
        XCTAssertTrue(button.exists)
        XCTAssertGreaterThanOrEqual(button.frame.minY, bounds.minY)
        XCTAssertLessThanOrEqual(button.frame.maxY, bounds.maxY)
        XCTAssertGreaterThanOrEqual(button.frame.minX, bounds.minX)
        XCTAssertLessThanOrEqual(button.frame.maxX, bounds.maxX)
        XCTAssertGreaterThanOrEqual(button.frame.height, 44)
      }
      XCTAssertTrue(app.buttons["Cancel"].firstMatch.isHittable)
    }
    for appearance in ["light", "dark"] {
      let choose = app.buttons["Use \(appearance) appearance"]
      reveal(choose); choose.tap()
      XCTAssertTrue(app.staticTexts["Appearance: \(appearance)"].waitForExistence(timeout: 5))
      let move = app.buttons["Move"].firstMatch
      XCTAssertFalse(move.isEnabled)
      footerVisible()
      capture("footer-\(appearance)-disabled-\(largeText ? "accessibility" : "default")")
      let select = app.buttons["Select destination"]
      reveal(select); select.tap()
      XCTAssertTrue(move.isEnabled)
      footerVisible()
      capture("footer-\(appearance)-enabled-\(largeText ? "accessibility" : "default")")
      XCTAssertTrue(move.isHittable); move.tap()
      XCTAssertTrue(app.staticTexts["Move received"].waitForExistence(timeout: 5))
      let clear = app.buttons["Clear destination"]
      reveal(clear); clear.tap()
    }
    app.buttons["Cancel"].firstMatch.tap()
    XCTAssertTrue(app.staticTexts["Footer appearance"].waitForNonExistence(timeout: 5))
  }

  func testMoveDestinationCreationRetainsDraftAndRetries() {
    guard openFixtureURL("audit-move-destination") else { return }
    func reveal(_ element: XCUIElement) {
      for _ in 0..<8 where !element.isHittable {
        let scroll = app.scrollViews.firstMatch
        if element.frame.minY < app.navigationBars.firstMatch.frame.maxY { scroll.swipeDown() }
        else { scroll.swipeUp() }
      }
      XCTAssertTrue(element.isHittable)
    }
    let existing = app.descendants(matching: .any)["Choose destination Camping box"].firstMatch
    XCTAssertTrue(existing.waitForExistence(timeout: 10))
    existing.tap()
    XCTAssertTrue(app.buttons["Move"].firstMatch.isEnabled)
    XCTAssertFalse(app.textFields["Put in"].exists)
    let query = app.searchFields.firstMatch
    XCTAssertTrue(query.waitForExistence(timeout: 5)); XCTAssertTrue(query.isHittable)
    capture("move-selection-idle")
    query.tap(); waitForKeyboard(keyLabel: "t", timeout: 30)
    query.typeText("Audit")
    waitForExactEnteredText("Audit", in: query)

    let dismiss = app.buttons["Dismiss keyboard"].firstMatch
    XCTAssertTrue(dismiss.isHittable); dismiss.tap()
    let newDestination = app.buttons["New destination"].firstMatch
    XCTAssertTrue(newDestination.waitForExistence(timeout: 5))
    reveal(newDestination); newDestination.tap()
    let name = app.textFields["New destination name"].firstMatch
    XCTAssertTrue(name.waitForExistence(timeout: 5)); XCTAssertEqual(name.value as? String, "Audit")
    XCTAssertTrue(name.isHittable)
    XCTAssertGreaterThanOrEqual(name.frame.minY, app.navigationBars["New destination"].frame.maxY)
    let kind = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose destination kind")).firstMatch
    XCTAssertTrue(kind.isHittable)
    XCTAssertGreaterThanOrEqual(kind.frame.minY, name.frame.maxY)
    capture("move-destination-creation-entry")

    reveal(kind); kind.tap()
    let container = app.buttons["Container"].firstMatch
    XCTAssertTrue(container.waitForExistence(timeout: 5)); container.tap()
    func finishName() {
      reveal(name); name.tap(); waitForKeyboard(keyLabel: "c"); name.typeText(" crate")
      XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(
        predicate: NSPredicate(format: "value == %@", "Audit crate"), object: name
      )], timeout: 5), .completed)
      XCTAssertTrue(dismiss.isHittable); dismiss.tap()
      XCTAssertTrue(kind.exists, "Editing the name must retain Kind")
    }
    finishName()
    XCTAssertTrue(app.navigationBars["New destination"].exists)
    XCTAssertFalse(app.buttons["Move"].exists)
    capture("move-destination-creation")

    let cancelCreation = app.buttons["Cancel new destination"].firstMatch
    reveal(cancelCreation); cancelCreation.tap()
    XCTAssertTrue(name.waitForNonExistence(timeout: 5))
    reveal(newDestination); newDestination.tap()
    XCTAssertTrue(name.waitForExistence(timeout: 5)); XCTAssertEqual(name.value as? String, "Audit")
    finishName()
    let create = app.buttons["Create destination"].firstMatch

    reveal(create); create.tap()
    let failure = app.alerts["Could not create destination"]
    XCTAssertTrue(failure.waitForExistence(timeout: 5))
    XCTAssertTrue(failure.staticTexts["Audit destination temporarily unavailable"].exists)
    failure.buttons["OK"].tap()
    XCTAssertEqual(name.value as? String, "Audit crate")
    reveal(create); create.tap()
    XCTAssertTrue(create.waitForNonExistence(timeout: 5))
    let selected = app.descendants(matching: .any)["Choose destination Audit crate"].firstMatch
    XCTAssertTrue(selected.waitForExistence(timeout: 5))
    XCTAssertEqual(selected.value as? String, "Selected")

    XCTAssertTrue(app.descendants(matching: .any)["Choose destination Audit crate"].exists)
    let move = app.buttons["Move"].firstMatch
    reveal(move); move.tap()
    let rejected = app.alerts["Could not move asset"]
    XCTAssertTrue(rejected.waitForExistence(timeout: 5))
    XCTAssertTrue(rejected.staticTexts["Audit move temporarily unavailable"].exists)
    rejected.buttons["OK"].tap()
    XCTAssertTrue(selected.exists)
    capture("move-destination-rejected-selection-retained")
    reveal(move); move.tap()
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
  }

  func testMoveHereRejectedCommandRetainsSelectionAndRetryReturns() {
    let open = app.buttons["Audit Move here recovery"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let retry = app.buttons["Retry suggestions"].firstMatch
    XCTAssertTrue(retry.waitForExistence(timeout: 10))
    retry.tap()
    let candidate = app.descendants(matching: .any)["Choose item Audit tent"].firstMatch
    XCTAssertTrue(candidate.waitForExistence(timeout: 5))
    XCTAssertTrue(candidate.isEnabled)
    XCTAssertTrue(candidate.isHittable)
    candidate.tap()
    XCTAssertEqual(candidate.value as? String, "Selected")
    let query = app.searchFields.firstMatch
    XCTAssertTrue(query.waitForExistence(timeout: 5)); XCTAssertTrue(query.isHittable)
    capture("move-selection-idle")
    query.tap(); waitForKeyboard(keyLabel: "t")
    query.typeText("Tent")
    let entered = XCTNSPredicateExpectation(predicate: NSPredicate(format: "value == %@", "Tent"), object: query)
    XCTAssertEqual(XCTWaiter.wait(for: [entered], timeout: 5), .completed)
    XCTAssertEqual(candidate.value as? String, "Selected")
    let clear = query.buttons["Clear text"].firstMatch
    XCTAssertTrue(clear.isHittable); clear.tap()
    dismissNativeSearchKeyboardIfNeeded()
    XCTAssertTrue(candidate.waitForExistence(timeout: 5))
    XCTAssertEqual(candidate.value as? String, "Selected")

    let header = app.navigationBars["Move something here"]
    let closeSearch = header.buttons["Close"].firstMatch
    if closeSearch.exists && closeSearch.isHittable { closeSearch.tap() }
    let move = header.buttons["Move here"].firstMatch
    XCTAssertTrue(move.waitForExistence(timeout: 5))
    XCTAssertTrue(move.isHittable)
    XCTAssertTrue(move.isEnabled)
    move.tap()
    XCTAssertTrue(app.alerts["Could not move asset here"].waitForExistence(timeout: 5))
    app.alerts.buttons["OK"].tap()
    XCTAssertEqual(candidate.value as? String, "Selected")
    XCTAssertTrue(app.staticTexts["Camping box"].firstMatch.exists)

    XCTAssertTrue(move.isHittable)
    move.tap()
    XCTAssertTrue(move.waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
    XCTAssertTrue(open.isHittable)
    XCTAssertEqual(app.state, .runningForeground)
    capture("move-here-successful-retry")
  }

  func testMoveHereSuggestionsRecoverAtNormalTextSize() {
    verifyMoveHereSuggestionsRecovery(captureSuffix: "normal-size")
  }

  func testMoveHereRecoveryAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
    verifyMoveHereSuggestionsRecovery(captureSuffix: "accessibility-size")
  }

  private func dismissNativeSearchKeyboardIfNeeded() {
    let keyboard = app.keyboards.firstMatch
    let dismiss = app.buttons["Dismiss keyboard"].firstMatch
    // Native search may already have ended keyboard editing when this snapshot settles.
    let settled = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      !keyboard.exists || dismiss.isHittable
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [settled], timeout: 5), .completed)
    if keyboard.exists && dismiss.isHittable { dismiss.tap() }
    XCTAssertTrue(keyboard.waitForNonExistence(timeout: 5))
  }

  private func verifyMoveHereSuggestionsRecovery(captureSuffix: String) {
    let open = app.buttons["Audit Move here recovery"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let query = app.searchFields.firstMatch
    XCTAssertTrue(query.waitForExistence(timeout: 10)); XCTAssertTrue(query.isHittable)
    capture("move-here-idle")
    query.tap(); waitForKeyboard(keyLabel: "t")

    query.typeText("Tent")
    XCTAssertEqual(query.value as? String, "Tent")
    let dismiss = app.buttons["Dismiss keyboard"]
    XCTAssertTrue(dismiss.isHittable)
    dismiss.tap()
    XCTAssertTrue(app.keyboards.firstMatch.waitForNonExistence(timeout: 5))
    let retry = app.buttons["Retry suggestions"].firstMatch
    XCTAssertTrue(retry.waitForExistence(timeout: 10))
    XCTAssertFalse(app.staticTexts["No movable matches"].exists)
    XCTAssertTrue(retry.isHittable)
    capture("move-here-suggestions-error-\(captureSuffix)")
    retry.tap()
    XCTAssertTrue(app.descendants(matching: .any)["Choose item Audit tent"].firstMatch.waitForExistence(timeout: 5))
    XCTAssertEqual(query.value as? String, "Tent")
    capture("move-here-suggestions-recovered-\(captureSuffix)")
    let clear = query.buttons["Clear text"].firstMatch
    XCTAssertTrue(clear.isHittable); clear.tap()
    dismissNativeSearchKeyboardIfNeeded()

    let cancel = app.navigationBars["Move something here"].buttons["Cancel"].firstMatch
    XCTAssertTrue(cancel.waitForExistence(timeout: 5)); XCTAssertTrue(cancel.isHittable)
    cancel.tap()
    XCTAssertTrue(cancel.waitForNonExistence(timeout: 5))
    XCTAssertTrue(open.isHittable)
  }

  func testCommandHeightComparisonAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
    let open = app.buttons["Audit command height"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let retry = app.buttons["Retry asset types"].firstMatch
    XCTAssertTrue(retry.waitForExistence(timeout: 10))
    XCTAssertTrue(app.staticTexts["Baseline size"].exists)
    capture("command-height-baseline")
    let compare = app.buttons["Compare shipping sizing"].firstMatch
    XCTAssertTrue(compare.isHittable)
    compare.tap()
    XCTAssertTrue(app.staticTexts["Shipping size"].waitForExistence(timeout: 5))
    XCTAssertTrue(retry.isHittable)
    capture("command-height-shipping")
    retry.tap()
    XCTAssertTrue(app.staticTexts["Retry received"].waitForExistence(timeout: 5))
    let back = app.navigationBars.buttons["BackButton"].firstMatch
    XCTAssertTrue(back.isHittable)
    back.tap()
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
  }

  func testColdInventoryQueriesEnableDependentResource() {
    let open = app.buttons["Audit inventory query"]
    XCTAssertTrue(open.waitForExistence(timeout: 5))
    open.tap()
    XCTAssertTrue(app.staticTexts["First query ready"].waitForExistence(timeout: 10))
    XCTAssertTrue(app.staticTexts["Dependent query ready"].waitForExistence(timeout: 10))
    capture("cold-inventory-dependent-queries")
  }

  func testInventorySwitcherHouseholdRetryAndClose() {
    let open = app.buttons["Audit inventory switcher"]
    XCTAssertTrue(open.waitForExistence(timeout: 5))
    open.tap()
    let change = app.buttons["Switch household"]
    XCTAssertTrue(change.waitForExistence(timeout: 10))
    XCTAssertTrue(change.isHittable)
    capture("inventory-switcher-entry")
    change.tap()
    let household = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Workshop household")).firstMatch
    XCTAssertTrue(household.waitForExistence(timeout: 5))
    household.tap()
    let inventory = app.buttons["Switch to inventory Workshop tools"]
    XCTAssertTrue(inventory.waitForExistence(timeout: 5))
    inventory.tap()
    XCTAssertTrue(app.staticTexts["Could not switch inventories. Try again."].waitForExistence(timeout: 5))
    capture("inventory-switcher-selection-error")
    XCTAssertTrue(inventory.isEnabled)
    inventory.tap()
    XCTAssertTrue(open.waitForExistence(timeout: 5))
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: app.navigationBars["Inventories"])], timeout: 5), .completed)
    open.tap()
    let close = app.buttons["Close inventory switcher"]
    XCTAssertTrue(close.waitForExistence(timeout: 5))
    close.tap()
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: app.navigationBars["Inventories"])], timeout: 5), .completed)
    capture("inventory-switcher-dismissed")
  }

  private func waitForKeyboard(keyLabel: String? = nil, timeout: TimeInterval = 15) {
    let keyboard = app.keyboards.firstMatch
    XCTAssertTrue(keyboard.waitForExistence(timeout: 5))
    let ready = NSPredicate { _, _ in
      let keys = keyLabel.map { [keyboard.keys[$0]] } ?? keyboard.keys.allElementsBoundByIndex
      return keys.contains { key in
        guard key.exists else { return false }
        let bounds = key.frame
        guard !bounds.isEmpty, !bounds.isNull, !bounds.isInfinite,
              bounds.origin.x.isFinite, bounds.origin.y.isFinite,
              bounds.width.isFinite, bounds.height.isFinite else { return false }
        return key.isHittable
      }
    }
    let result = observePredicate("keyboard-readiness-timing", predicate: ready, object: nil, timeout: timeout)
    if result != .completed {
      recordHitTestState("keyboard-readiness", elements: [keyboard] + (keyLabel.map { [keyboard.keys[$0]] } ?? []))
    }
    XCTAssertEqual(result, .completed, "Typing requires an interactive keyboard")
  }

  private func observePredicate(_ name: String, predicate: NSPredicate, object: Any?, timeout: TimeInterval = 5, immediately: Bool = false) -> XCTWaiter.Result {
    let started = ProcessInfo.processInfo.systemUptime
    var observations: [String] = []
    func evaluate() -> Bool {
      let before = ProcessInfo.processInfo.systemUptime
      let matched = predicate.evaluate(with: object)
      let after = ProcessInfo.processInfo.systemUptime
      observations.append("start=\(before - started), duration=\(after - before), matched=\(matched)")
      return matched
    }
    let initialMatch = immediately && evaluate()
    let elapsed = ProcessInfo.processInfo.systemUptime - started
    let result: XCTWaiter.Result
    if immediately && elapsed >= timeout {
      result = .timedOut
    } else if initialMatch {
      result = .completed
    } else {
      let expectation = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in evaluate() }, object: nil)
      result = XCTWaiter.wait(for: [expectation], timeout: immediately ? timeout - elapsed : timeout)
    }
    let boundedResult: XCTWaiter.Result = immediately && ProcessInfo.processInfo.systemUptime - started > timeout ? .timedOut : result
    observations.append("wait duration=\(ProcessInfo.processInfo.systemUptime - started), result=\(boundedResult.rawValue)")
    let attachment = XCTAttachment(string: observations.joined(separator: "\n"))
    attachment.name = name
    attachment.lifetime = .keepAlways
    add(attachment)
    return boundedResult
  }

  private func recordHitTestState(_ name: String, elements: [XCUIElement]) {
    let lines = elements.map { element in
      guard element.exists else { return "Element no longer exists" }
      let bounds = element.frame
      guard !bounds.isEmpty, !bounds.isNull, !bounds.isInfinite,
            bounds.origin.x.isFinite, bounds.origin.y.isFinite,
            bounds.width.isFinite, bounds.height.isFinite else {
        return "\(element.label): frame=\(bounds), hittability skipped for invalid bounds"
      }
      return "\(element.label): frame=\(bounds), hittable=\(element.isHittable)"
    }
    let attachment = XCTAttachment(string: lines.joined(separator: "\n"))
    attachment.name = name
    attachment.lifetime = .keepAlways
    add(attachment)
  }

  private func capture(_ name: String) {
    let attachment = XCTAttachment(screenshot: app.screenshot())
    attachment.name = name
    attachment.lifetime = .keepAlways
    add(attachment)
    let hierarchy = XCTAttachment(string: app.debugDescription)
    hierarchy.name = "\(name)-hierarchy"
    hierarchy.lifetime = .keepAlways
    add(hierarchy)
    // debugDescription truncates AX values; retain the safe fixture diagnostic verbatim.
    let diagnostics = app.descendants(matching: .any).matching(identifier: "audit-query-readiness")
    for (index, element) in diagnostics.allElementsBoundByIndex.enumerated() {
      let value = [element.value as? String, element.label].compactMap { $0 }.first { candidate in
        guard let data = candidate.data(using: .utf8) else { return false }
        return (try? JSONSerialization.jsonObject(with: data)) is [String: Any]
      }
      guard let value else { continue }
      let readiness = XCTAttachment(string: value)
      readiness.name = "\(name)-query-readiness-\(index)"
      readiness.lifetime = .keepAlways
      add(readiness)
    }
  }
  private func verifyFullSheetLayout(_ variant: String) {
    let open = app.buttons["Audit \(variant) sheet"]
    for _ in 0..<5 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    XCTAssertTrue(app.navigationBars["Sheet diagnostic"].waitForExistence(timeout: 5))
    capture("sheet-\(variant)-layout")
    let row = app.buttons["Diagnostic Tags"]
    XCTAssertTrue(row.waitForExistence(timeout: 5))
    XCTAssertTrue(row.isHittable)
    if variant.contains("footer") {
      let finish = app.buttons["Finish diagnostic"]
      XCTAssertTrue(finish.isHittable)
      let geometry = XCTAttachment(string: "Finish frame: \(finish.frame); app frame: \(app.frame). Inspect against the sheet bounds in the retained screenshot/hierarchy.")
      geometry.name = "footer-placement-\(variant)"
      geometry.lifetime = .keepAlways
      add(geometry)
    }
  }

  func testDirectFullSheetLayout() { verifyFullSheetLayout("direct") }
  func testNestedFullSheetLayout() { verifyFullSheetLayout("nested") }
  func testFooterFullSheetLayout() { verifyFullSheetLayout("footer") }
  func testDirectFooterFullSheetLayout() { verifyFullSheetLayout("direct-footer") }
  func testScrollFooterFullSheetLayout() { verifyFullSheetLayout("scroll-footer") }

  func testCheckoutHistoryTextBoundsPaginationAndDismissal() {
    let open = app.buttons["Audit Checkout history"]
    for _ in 0..<7 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let bar = app.navigationBars["Checkout history"]
    XCTAssertTrue(bar.waitForExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["Audit checkout 1: borrowed for cleaning the gutters."].waitForExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["Asset name could not be loaded."].waitForExistence(timeout: 5))
    let historyScroll = app.scrollViews.containing(.staticText, identifier: "Audit checkout 1: borrowed for cleaning the gutters.").firstMatch
    XCTAssertTrue(historyScroll.exists)
    let note = historyScroll.staticTexts.matching(identifier: "Audit checkout 1: borrowed for cleaning the gutters.").firstMatch
    XCTAssertTrue(textFitsHistoryViewport(note, scroll: historyScroll, bar: bar))
    XCTAssertTrue(app.buttons["Close"].isHittable)
    capture("checkout-history-medium")
    let initialTop = bar.frame.minY
    bar.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
      .press(forDuration: 0.1, thenDragTo: app.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.12)))
    if UIDevice.current.userInterfaceIdiom == .phone {
      let expanded = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in bar.frame.minY < initialTop - 40 }, object: nil)
      XCTAssertEqual(XCTWaiter.wait(for: [expanded], timeout: 5), .completed)
    }
    XCTAssertTrue(textFitsHistoryViewport(note, scroll: historyScroll, bar: bar))
    capture("checkout-history-expanded")
    let retryName = app.buttons["Try loading asset name again"]
    for _ in 0..<6 where !retryName.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(retryName.isHittable)
    XCTAssertTrue(app.frame.contains(retryName.frame))
    capture("checkout-history-name-recovery")
    retryName.tap()
    let recoveredName = app.staticTexts["Audit ladder"]
    XCTAssertTrue(recoveredName.waitForExistence(timeout: 5))
    XCTAssertFalse(app.staticTexts["Asset name could not be loaded."].exists)
    let older = app.buttons["Load older checkouts"]
    for _ in 0..<6 where !older.isHittable {
      historyScroll.swipeUp()
    }
    XCTAssertTrue(older.isHittable)
    older.tap()
    let loaded = historyScroll.staticTexts.matching(identifier: "Older audit checkout").firstMatch
    XCTAssertTrue(loaded.waitForExistence(timeout: 5))
    for _ in 0..<4 where !textFitsHistoryViewport(loaded, scroll: historyScroll, bar: bar) { historyScroll.swipeUp() }
    XCTAssertTrue(textFitsHistoryViewport(loaded, scroll: historyScroll, bar: bar))
    capture("checkout-history-older-page")
    app.buttons["Close"].tap()
    XCTAssertTrue(open.waitForExistence(timeout: 5))
    XCTAssertFalse(bar.exists)
  }

  private func textFitsHistoryViewport(_ text: XCUIElement, scroll: XCUIElement, bar: XCUIElement) -> Bool {
    guard text.exists else { return false }
    let frame = text.frame
    let viewport = scroll.frame.intersection(app.frame)
    return !frame.isEmpty && !frame.isInfinite && !viewport.isNull
      && viewport.contains(frame) && frame.minY >= bar.frame.maxY
  }

  func testBrowseLastTagClearsActionFooterAndApplies() {
    app.buttons["Audit Browse filters"].tap()
    let tags = app.buttons["Choose tags"]
    XCTAssertTrue(tags.waitForExistence(timeout: 5)); tags.tap()
    let last = app.descendants(matching: .any).matching(identifier: "Filter by tag ZZ final tag").firstMatch
    XCTAssertTrue(last.waitForExistence(timeout: 5))
    let footer = app.otherElements["browse-filter-footer"].firstMatch
    XCTAssertTrue(footer.waitForExistence(timeout: 5))
    let scroll = app.scrollViews.containing(.any, identifier: "Filter by tag ZZ final tag").firstMatch
    XCTAssertTrue(scroll.exists)
    func fullyAboveActions() -> Bool {
      let bounds = scroll.frame.intersection(app.frame)
      let top = max(bounds.minY, app.navigationBars.firstMatch.frame.maxY)
      return last.isHittable && last.frame.height > 0 && last.frame.minY >= top && last.frame.maxY <= min(bounds.maxY, footer.frame.minY)
    }
    for _ in 0..<24 where !fullyAboveActions() {
      scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.65))
        .press(forDuration: 0.05, thenDragTo: scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.35)))
    }
    capture("browse-last-tag-above-actions")
    XCTAssertTrue(fullyAboveActions(), "The final tag must scroll completely above both fixed actions")
    let bounds = scroll.frame.intersection(app.frame)
    XCTAssertFalse(footer.frame.isEmpty)
    XCTAssertTrue(bounds.contains(footer.frame), "The footer must be fully inside the visible sheet")
    XCTAssertGreaterThanOrEqual(footer.frame.minY, app.navigationBars.firstMatch.frame.maxY)
    for label in ["Show results", "Back to filters"] {
      let action = app.buttons[label]
      XCTAssertTrue(action.isHittable)
      XCTAssertFalse(action.frame.isEmpty)
      XCTAssertTrue(footer.frame.intersection(bounds).contains(action.frame), "Every action must be fully visible")
    }
    last.tap()
    let selected = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      (last.value as? String) == "checkbox, checked"
    }, object: nil)
    let selectionResult = XCTWaiter.wait(for: [selected], timeout: 5)
    let selectionValue = XCTAttachment(string: "Last tag after tap: \(String(describing: last.value))")
    selectionValue.name = "browse-last-tag-selection"
    selectionValue.lifetime = .keepAlways
    add(selectionValue)
    capture("browse-last-tag-after-selection")
    XCTAssertEqual(selectionResult, .completed, "The last tag must be selected before returning to filters")
    app.buttons["Back to filters"].tap()
    XCTAssertTrue(app.buttons["Choose tags"].waitForExistence(timeout: 5))
    app.buttons["Show results"].tap()
    XCTAssertTrue(app.staticTexts["Browse selected tags: audit-last"].waitForExistence(timeout: 5))
  }

  private func assertFilterActionsClearKeyboard(_ apply: XCUIElement, _ back: XCUIElement) {
    let dismiss = app.buttons["Dismiss keyboard"]
    XCTAssertTrue(dismiss.waitForExistence(timeout: 5))
    let clearAccessory = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      let top = dismiss.frame.minY
      return dismiss.isHittable && [apply, back].allSatisfy { action in
        action.isHittable && !action.frame.isEmpty &&
          self.app.frame.contains(action.frame) && action.frame.maxY <= top
      }
    }, object: nil)
    let result = XCTWaiter.wait(for: [clearAccessory], timeout: 5)
    if result != .completed {
      recordHitTestState("filter-actions-keyboard-clearance", elements: [app, apply, back, dismiss])
    }
    XCTAssertEqual(result, .completed,
      "Both filter commands must be fully above the keyboard-dismiss accessory")
  }

  func testBrowseTagSearchKeepsActionsAboveKeyboardAccessory() {
    app.buttons["Audit Browse filters"].tap()
    let tags = app.buttons["Choose tags"]
    XCTAssertTrue(tags.waitForExistence(timeout: 5)); tags.tap()
    let searchButton = app.buttons["Search"].firstMatch
    XCTAssertTrue(searchButton.waitForExistence(timeout: 5)); searchButton.tap()
    let search = app.searchFields.firstMatch
    XCTAssertTrue(search.waitForExistence(timeout: 5))
    waitForKeyboard(keyLabel: "t")
    search.typeText("Tools")
    XCTAssertEqual(observePredicate("search-query-timing",
      predicate: NSPredicate(format: "value == %@", "Tools"), object: search),
      .completed, "Native search must retain the complete query")
    let tools = app.descendants(matching: .any).matching(identifier: "Filter by tag Tools").firstMatch
    XCTAssertTrue(tools.waitForExistence(timeout: 5))
    XCTAssertTrue(app.descendants(matching: .any).matching(identifier: "Filter by tag Holiday supplies").firstMatch.waitForNonExistence(timeout: 5))
    let apply = app.buttons["Show results"]
    let back = app.buttons["Back to filters"]
    capture("browse-search-keyboard")
    assertFilterActionsClearKeyboard(apply, back)
    tools.tap()
    XCTAssertTrue(app.keyboards.firstMatch.exists)
    assertFilterActionsClearKeyboard(apply, back)
    capture("browse-search-actions-clear-accessory")
    apply.tap()
    XCTAssertTrue(app.staticTexts["Browse selected tags: audit-tools"].waitForExistence(timeout: 5))
  }

  func testBrowseUsesInPlaceAvailabilityMenuAndReachableActions() throws {
    app.buttons["Audit Browse filters"].tap()
    let availability = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose availability")).firstMatch
    XCTAssertTrue(availability.waitForExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["Availability"].exists, "The value must retain its visible field label")
    capture("browse-filter-overview")
    availability.tap()
    let available = app.buttons["Available"]
    XCTAssertTrue(available.waitForExistence(timeout: 5))
    capture("browse-availability-menu")
    available.tap()
    XCTAssertTrue(app.navigationBars["Filters"].exists)
    let apply = app.buttons["Show results"]
    XCTAssertTrue(apply.isHittable)
    XCTAssertTrue(app.buttons["Cancel filters"].isHittable)
    apply.tap()
    XCTAssertTrue(app.staticTexts["Browse availability: available"].waitForExistence(timeout: 5))
    capture("browse-applied")
  }
  func testBrowseFiltersDetailAndBackPreserveContext() {
    guard openFixtureURL("search?query=Camping") else { return }
    let filters = app.buttons["Filters"].firstMatch
    XCTAssertTrue(filters.waitForExistence(timeout: 10)); filters.tap()
    let availability = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose availability")).firstMatch
    XCTAssertTrue(availability.waitForExistence(timeout: 5)); availability.tap()
    let available = app.buttons["Available"].firstMatch
    XCTAssertTrue(available.waitForExistence(timeout: 5)); available.tap()
    capture("connected-browse-filter-draft")
    app.buttons["Show results"].firstMatch.tap()
    let summary = app.staticTexts.matching(NSPredicate(format: "label CONTAINS %@ AND label CONTAINS %@", "18 shown", "Camping")).firstMatch
    XCTAssertTrue(summary.waitForExistence(timeout: 10))
    let list = app.scrollViews.firstMatch
    let cards = app.buttons.matching(NSPredicate(format: "label BEGINSWITH %@", "Open asset Camping item"))
    let initial = app.buttons[cards.firstMatch.label].firstMatch
    XCTAssertTrue(initial.isHittable)
    let initialY = initial.frame.minY
    list.swipeUp()
    XCTAssertTrue(!initial.exists || !initial.isHittable || abs(initial.frame.minY - initialY) > 20,
      "The list must actually move before testing scroll restoration")
    guard let card = cards.allElementsBoundByIndex.first(where: { $0.isHittable }) else {
      XCTFail("A filtered asset must be visible after scrolling"); return
    }
    let label = card.label
    let position = card.frame.minY
    capture("connected-browse-filter-results-scrolled")
    card.tap()
    XCTAssertTrue(app.navigationBars["Details"].waitForExistence(timeout: 10))
    capture("connected-browse-filter-detail")
    app.navigationBars.buttons["Back"].firstMatch.tap()
    let restored = app.buttons[label].firstMatch
    XCTAssertTrue(restored.waitForExistence(timeout: 10)); XCTAssertTrue(restored.isHittable)
    XCTAssertEqual(restored.frame.minY, position, accuracy: 3)
    XCTAssertTrue(summary.exists)
    capture("connected-browse-filter-detail-return")
  }

  func testBrowseExpirationDetailAndBackPreserveContext() {
    guard openFixtureURL("search?query=Camping") else { return }
    let filters = app.buttons["Filters"].firstMatch
    XCTAssertTrue(filters.waitForExistence(timeout: 10)); filters.tap()
    let availability = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose availability")).firstMatch
    XCTAssertTrue(availability.waitForExistence(timeout: 5)); availability.tap()
    let available = app.buttons["Available"].firstMatch
    XCTAssertTrue(available.waitForExistence(timeout: 5)); available.tap()
    let menu = app.buttons["Choose expiration review"].firstMatch
    let filterBody = app.scrollViews.containing(.button, identifier: "Choose expiration review").firstMatch
    for _ in 0..<5 {
      if menu.isHittable { break }
      // Start above the fixed footer so the gesture belongs to the filter list.
      let start = filterBody.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.45))
      let end = filterBody.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.15))
      start.press(forDuration: 0.05, thenDragTo: end)
    }
    XCTAssertTrue(menu.isHittable)
    capture("connected-expiration-filter-entry")
    menu.tap(); app.buttons["Expired"].firstMatch.tap()
    XCTAssertTrue(app.navigationBars["Expiration"].waitForExistence(timeout: 10))
    let expired = app.buttons.matching(NSPredicate(format: "label BEGINSWITH %@", "Open asset Camping item 03")).firstMatch
    XCTAssertTrue(expired.waitForExistence(timeout: 10)); XCTAssertTrue(expired.isHittable)
    XCTAssertFalse(app.buttons.matching(NSPredicate(format: "label BEGINSWITH %@", "Open asset Camping item 01")).firstMatch.exists)
    XCTAssertFalse(app.buttons.matching(NSPredicate(format: "label BEGINSWITH %@", "Open asset Kitchen item")).firstMatch.exists)
    // The native viewport must not retain the dismissed filter sheet's size.
    let viewport = app.scrollViews.firstMatch.frame
    XCTAssertGreaterThanOrEqual(viewport.width, app.frame.width - 40)
    XCTAssertGreaterThanOrEqual(viewport.height, app.frame.height - app.navigationBars["Expiration"].frame.maxY - 40)
    for label in ["All dates", "Expiring soon", "Expired"] {
      let mode = app.buttons[label].firstMatch
      XCTAssertTrue(mode.exists && mode.isHittable)
      XCTAssertTrue(viewport.contains(mode.frame), "Every mode must fit within the actual native viewport")
      mode.tap()
      XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "selected == true"), object: mode)], timeout: 5), .completed)
    }
    XCTAssertTrue(expired.waitForExistence(timeout: 10)); XCTAssertTrue(expired.isHittable)
    capture("connected-expiration-filter-results")
    expired.tap()
    XCTAssertTrue(app.navigationBars["Details"].waitForExistence(timeout: 10))
    app.navigationBars.buttons["Back"].firstMatch.tap()
    XCTAssertTrue(expired.waitForExistence(timeout: 10)); XCTAssertTrue(expired.isHittable)
    capture("connected-expiration-detail-return")
    app.navigationBars.buttons["Back"].firstMatch.tap()
    XCTAssertTrue(filters.waitForExistence(timeout: 10))
    let summary = app.staticTexts.matching(NSPredicate(format: "label CONTAINS %@ AND label CONTAINS %@", "24 shown", "Camping")).firstMatch
    XCTAssertTrue(summary.exists)
    capture("connected-expiration-browse-return")
  }

  func testBrowseExpirationReviewUsesOverviewMenu() {
    app.buttons["Audit Browse filters"].tap()
    let menu = app.buttons["Choose expiration review"].firstMatch
    XCTAssertTrue(menu.waitForExistence(timeout: 5))
    let sort = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose sort")).firstMatch
    XCTAssertTrue(sort.isHittable, "Ordinary filter choices must be reachable in the initial large sheet")
    XCTAssertTrue(app.buttons["Show results"].firstMatch.isHittable)
    XCTAssertTrue(app.buttons["Cancel filters"].firstMatch.isHittable)
    capture("browse-filter-overview")
    let scroll = app.scrollViews.containing(.button, identifier: "Choose expiration review").firstMatch
    for _ in 0..<6 {
      if menu.isHittable { break }
      scroll.swipeUp()
    }
    XCTAssertTrue(menu.isHittable)
    menu.tap()
    let expired = app.buttons["Expired"].firstMatch
    XCTAssertTrue(expired.waitForExistence(timeout: 5))
    XCTAssertTrue(expired.isHittable)
    XCTAssertTrue(app.navigationBars["Filters"].exists)
    XCTAssertFalse(app.navigationBars["Expiration"].exists)
    capture("browse-expiration-review-menu")
    expired.tap()
    XCTAssertTrue(app.staticTexts["Expiration mode: expired"].waitForExistence(timeout: 5))
  }

  func testExpirationSheetBodySurvivesExpansion() {
    app.buttons["Audit medium expiration filters"].tap()
    XCTAssertTrue(app.buttons["Choose tags"].waitForExistence(timeout: 5))
    capture("expiration-medium-body")
    let bar = app.navigationBars["Filters"]
    let initialTop = bar.frame.minY
    bar.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
      .press(forDuration: 0.1, thenDragTo: app.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.12)))
    if UIDevice.current.userInterfaceIdiom == .phone {
      let expanded = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
        bar.frame.minY < initialTop - 40
      }, object: nil)
      XCTAssertEqual(XCTWaiter.wait(for: [expanded], timeout: 5), .completed, "The sheet must actually expand")
    }
    XCTAssertTrue(app.buttons["Choose tags"].isHittable)
    capture("expiration-expanded-body")
    XCTAssertTrue(app.buttons["Apply expiration filters"].isHittable)
  }

  func testDetailContextVariantsKeepNativeCommandsAndStatus() {
    for (variant, title) in [("photo", "Camping tent"), ("checked-out", "Camping gear"),
                             ("read-only", "Spare camping gear"), ("place", "Garage")] {
      guard openFixtureURL("audit-detail-commands?variant=" + variant) else { return }
      let heading = app.staticTexts[title].firstMatch
      XCTAssertTrue(heading.waitForExistence(timeout: 10))
      let edit = app.navigationBars.buttons["Edit"].firstMatch
      if variant == "read-only" {
        XCTAssertTrue(edit.waitForNonExistence(timeout: 5))
        XCTAssertFalse(app.buttons["Move"].exists)
        XCTAssertFalse(app.buttons["Return"].exists)
      } else {
        XCTAssertTrue(edit.waitForExistence(timeout: 5)); XCTAssertTrue(edit.isHittable)
      }
      if variant == "photo" {
        let photo = app.buttons["Open photo 1 of 1"].firstMatch
        XCTAssertTrue(photo.waitForExistence(timeout: 5))
        XCTAssertGreaterThan(photo.frame.width, 0)
        XCTAssertLessThanOrEqual(photo.frame.width, 688 + 1)
        XCTAssertGreaterThanOrEqual(photo.frame.minX, 0)
        XCTAssertLessThanOrEqual(photo.frame.maxX, app.frame.maxX)
        XCTAssertEqual(photo.frame.midX, app.frame.midX, accuracy: 2)
        XCTAssertFalse(app.staticTexts["No photos"].exists)
      }
      if variant == "place" {
        XCTAssertTrue(app.buttons["Search"].firstMatch.waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Search"].firstMatch.isHittable)
        XCTAssertTrue(app.buttons["Move place"].firstMatch.isHittable)
        XCTAssertFalse(app.staticTexts["Availability"].exists)
      }
      capture("detail-context-" + variant + "-entry")
      if variant == "checked-out" || variant == "read-only" {
        let status = app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH %@", "Checked out ")).firstMatch
        let scroll = app.scrollViews.firstMatch
        for _ in 0..<6 where !status.isHittable { scroll.swipeUp() }
        XCTAssertTrue(status.isHittable)
        let headings = app.staticTexts.matching(identifier: "Availability").allElementsBoundByIndex
        for heading in headings {
          let frame = heading.frame
          XCTAssertFalse(frame.isEmpty || frame.isNull || frame.isInfinite)
          XCTAssertTrue([frame.minX, frame.minY, frame.width, frame.height].allSatisfy { $0.isFinite })
        }
        XCTAssertEqual(Set(headings.map { NSCoder.string(for: $0.frame) }).count, 1)
        if variant == "checked-out" { XCTAssertTrue(app.buttons["Return"].firstMatch.isHittable) }
        capture("detail-context-" + variant + "-availability")
      }
    }
  }

  func testDetailGalleryKeepsLaterPhotoInReadableColumn() {
    guard openFixtureURL("audit-detail-commands?variant=gallery") else { return }
    let first = app.buttons["Open photo 1 of 3"].firstMatch
    let second = app.buttons["Open photo 2 of 3"].firstMatch
    XCTAssertTrue(first.waitForExistence(timeout: 10))
    XCTAssertTrue(first.isHittable)
    first.swipeLeft(velocity: .slow)
    let secondCentered = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      second.isHittable && abs(second.frame.midX - self.app.frame.midX) < 3
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [secondCentered], timeout: 5), .completed)
    XCTAssertLessThanOrEqual(second.frame.width, 689)
    capture("detail-gallery-second-portrait")
    defer { XCUIDevice.shared.orientation = .portrait }
    if UIDevice.current.userInterfaceIdiom == .pad {
      XCUIDevice.shared.orientation = .landscapeLeft
      let landscape = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
        self.app.frame.width > self.app.frame.height && second.isHittable
          && abs(second.frame.midX - self.app.frame.midX) < 3
      }, object: nil)
      XCTAssertEqual(XCTWaiter.wait(for: [landscape], timeout: 10), .completed)
      XCTAssertLessThanOrEqual(second.frame.width, 689)
      capture("detail-gallery-second-landscape")
    }
    second.tap()
    XCTAssertTrue(app.buttons["Close photo viewer"].waitForExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["Fixture image 2.png"].waitForExistence(timeout: 5))
    app.buttons["Close photo viewer"].tap()
  }

  func testDetailCommandsRemainReachableAtNormalTextSize() {
    verifyDetailCommandReachability(captureSuffix: "normal-size")
  }

  func testDetailCommandsRemainReachableAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
    verifyDetailCommandReachability(captureSuffix: "accessibility-size")
  }

  private func verifyDetailCommandReachability(captureSuffix: String) {
    let open = app.buttons["Audit detail commands"]
    for _ in 0..<14 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let add = app.buttons["Add item here"].firstMatch
    XCTAssertTrue(add.waitForExistence(timeout: 10))
    for label in ["Add item here", "Move items here", "Check out", "Edit", "Move"] {
      let command = label == "Edit" ? app.navigationBars.buttons["Edit"].firstMatch : app.buttons[label].firstMatch
      XCTAssertTrue(command.exists)
      let scroll = app.scrollViews.firstMatch
      func fullyVisible() -> Bool {
        if label == "Edit" {
          return command.isHittable && app.navigationBars.firstMatch.frame.contains(command.frame)
        }
        let visible = scroll.frame.intersection(app.frame)
        let top = max(visible.minY, app.navigationBars.firstMatch.frame.maxY)
        return command.isHittable && command.frame.minY >= top && command.frame.maxY <= visible.maxY
      }
      for _ in 0..<12 where !fullyVisible() {
        let above = command.frame.minY < app.navigationBars.firstMatch.frame.maxY
        let start = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.4 : 0.7))
        let end = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.7 : 0.4))
        start.press(forDuration: 0.05, thenDragTo: end)
      }
      XCTAssertTrue(fullyVisible(), label)
      if label != "Edit" { XCTAssertGreaterThanOrEqual(command.frame.height, 44, label) }
      XCTAssertGreaterThanOrEqual(command.frame.minX, app.frame.minX, label)
      XCTAssertLessThanOrEqual(command.frame.maxX, app.frame.maxX, label)
      XCTAssertGreaterThanOrEqual(command.frame.width, 44, label)
      capture("detail-command-" + label.lowercased().replacingOccurrences(of: " ", with: "-") + "-" + captureSuffix)
    }
    let back = app.navigationBars.buttons.firstMatch
    XCTAssertTrue(back.isHittable)
    back.tap()
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
    XCTAssertTrue(add.waitForNonExistence(timeout: 5))
    XCTAssertTrue(open.isHittable)
  }

  func testPlaceContentsUseNativeSearchAndKeepNavigation() {
    verifyPlaceSearch(openLabel: "Audit place search")
  }

  func testPreconfiguredPlaceSearchKeepsProductionHandlersAndNavigation() {
    verifyPlaceSearch(openLabel: "Audit preconfigured place search")
  }

  private func verifyPlaceSearch(openLabel: String) {
    let open = app.buttons[openLabel]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let more = app.buttons["More actions for Audit place"]
    XCTAssertTrue(more.waitForExistence(timeout: 10))
    XCTAssertTrue(more.isHittable)
    let searchButton = app.buttons["Search"].firstMatch
    XCTAssertTrue(searchButton.waitForExistence(timeout: 10))
    XCTAssertTrue(searchButton.isHittable)
    let header = app.navigationBars.firstMatch
    XCTAssertGreaterThanOrEqual(searchButton.frame.minY, header.frame.minY)
    XCTAssertLessThanOrEqual(searchButton.frame.maxY, header.frame.maxY)
    let identity = app.staticTexts["Audit place"].firstMatch
    XCTAssertTrue(identity.waitForExistence(timeout: 5))
    XCTAssertGreaterThanOrEqual(identity.frame.minY, header.frame.maxY,
      "Place identity must clear native navigation chrome on initial entry")
    XCTAssertLessThanOrEqual(identity.frame.maxY, app.frame.maxY)
    capture("place-search-collapsed")
    searchButton.tap()
    let field = app.searchFields.firstMatch
    XCTAssertTrue(field.waitForExistence(timeout: 5))
    XCTAssertTrue(field.isHittable)
    XCTAssertEqual(field.placeholderValue, "Search this place")
    field.tap()
    field.typeText("19")
    XCTAssertEqual(field.value as? String, "19")
    XCTAssertTrue(app.buttons["Open asset Tool 19. Item"].firstMatch.waitForExistence(timeout: 5))
    XCTAssertFalse(app.buttons["Open asset Tool 0. Item"].exists)
    let dismissKeyboard = app.buttons["Dismiss keyboard"].firstMatch
    XCTAssertTrue(dismissKeyboard.isHittable)
    dismissKeyboard.tap()
    XCTAssertTrue(app.keyboards.firstMatch.waitForNonExistence(timeout: 5))
    let result = app.buttons["Open asset Tool 19. Item"].firstMatch
    let scroll = app.scrollViews.containing(.button, identifier: "Open asset Tool 19. Item").firstMatch
    XCTAssertTrue(scroll.exists)
    func resultVisible() -> Bool {
      let bounds = scroll.frame.intersection(app.frame)
      let top = max(bounds.minY, app.navigationBars.firstMatch.frame.maxY)
      return result.isHittable && result.frame.minY >= top && result.frame.maxY <= bounds.maxY
    }
    for _ in 0..<12 where !resultVisible() {
      let above = result.frame.minY < app.navigationBars.firstMatch.frame.maxY
      let start = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.4 : 0.7))
      let end = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.7 : 0.4))
      start.press(forDuration: 0.05, thenDragTo: end)
    }
    XCTAssertTrue(resultVisible())
    XCTAssertEqual(field.value as? String, "19")
    capture("place-search-filtered")
    let clear = field.buttons["Clear text"].firstMatch
    XCTAssertTrue(clear.isHittable)
    clear.tap()
    XCTAssertTrue(app.buttons["Open asset Tool 0. Item"].firstMatch.waitForExistence(timeout: 5))
    let cleared = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      (field.exists && field.isHittable) || (searchButton.exists && searchButton.isHittable)
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [cleared], timeout: 5), .completed)
    capture("place-search-cleared")
    if !field.isHittable {
      XCTAssertTrue(searchButton.isHittable)
      searchButton.tap()
      XCTAssertTrue(field.waitForExistence(timeout: 5))
    }
    XCTAssertTrue(field.isHittable)
    field.tap()
    field.typeText("19")
    XCTAssertEqual(field.value as? String, "19")
    XCTAssertTrue(app.buttons["Open asset Tool 19. Item"].firstMatch.waitForExistence(timeout: 5))
    XCTAssertFalse(app.buttons["Open asset Tool 0. Item"].exists)
    resetNativeSearch(field)
    XCTAssertTrue(app.buttons["Open asset Tool 0. Item"].firstMatch.waitForExistence(timeout: 5))
    XCTAssertTrue(searchButton.isHittable || (UIDevice.current.userInterfaceIdiom == .pad && field.isHittable))
    XCTAssertTrue(more.isHittable)
    capture("place-search-cancelled")
    let back = app.navigationBars.buttons.firstMatch
    XCTAssertTrue(back.isHittable)
    back.tap()
    XCTAssertTrue(open.waitForExistence(timeout: 5))
  }

  func testManagedSearchPlacementAfterEnableAndHeaderUpdate() {
    guard openFixtureURL("audit-managed-search") else { return }
    let enable = app.buttons["Enable managed search"]
    XCTAssertTrue(enable.waitForExistence(timeout: 10))
    enable.tap()
    XCTAssertTrue(app.staticTexts["Managed search enabled"].waitForExistence(timeout: 5))
    func headerSearchIsPresent(_ title: String) -> Bool {
      let header = app.navigationBars[title]
      let search = header.buttons["Search"].firstMatch
      let found = search.waitForExistence(timeout: 5)
      let field = app.searchFields["Managed search probe"].firstMatch
      let geometry = XCTAttachment(string: "header=\(header.frame); search=\(found ? String(describing: search.frame) : "absent"); field=\(field.exists ? String(describing: field.frame) : "absent")")
      geometry.name = title
      geometry.lifetime = .keepAlways
      add(geometry)
      return found && search.isHittable && header.frame.contains(search.frame)
    }
    let initialPlacement = headerSearchIsPresent("Managed search")
    capture("managed-search-after-enable")
    app.buttons["Reconfigure search header"].tap()
    XCTAssertTrue(app.navigationBars["Search reconfigured"].waitForExistence(timeout: 5))
    let updatedPlacement = headerSearchIsPresent("Search reconfigured")
    capture("managed-search-after-header-update")
    app.buttons["Add native header action"].tap()
    let action = app.navigationBars["Search reconfigured"].buttons["Probe action"].firstMatch
    XCTAssertTrue(action.waitForExistence(timeout: 5))
    let actionPlacement = headerSearchIsPresent("Search reconfigured")
    capture("managed-search-with-native-action")
    XCTAssertTrue(action.isHittable)
    if action.isHittable { action.tap() }
    XCTAssertTrue(app.staticTexts["Action activations: 1"].waitForExistence(timeout: 5))
    XCTAssertTrue(actionPlacement, "Native header actions must coexist with header search")
    XCTAssertTrue(initialPlacement, "Delayed native search must initially use the header")
    XCTAssertTrue(updatedPlacement, "Header updates must retain search placement")
  }

  func testStaticNativeSearchPlacementComparison() {
    let open = app.buttons["Audit static search placement"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let header = app.navigationBars["Search placement"]
    XCTAssertTrue(header.waitForExistence(timeout: 10))
    let search = header.buttons["Search"].firstMatch
    let ready = search.waitForExistence(timeout: 10)
    capture("static-search-placement-idle")
    XCTAssertTrue(ready, "Registered integrated-button search must appear in the navigation bar")
    guard ready else { return }
    XCTAssertTrue(search.isHittable)
    XCTAssertGreaterThanOrEqual(search.frame.minY, header.frame.minY)
    XCTAssertLessThanOrEqual(search.frame.maxY, header.frame.maxY)
    search.tap()
    let field = app.searchFields["Search placement probe"].firstMatch
    XCTAssertTrue(field.waitForExistence(timeout: 5))
    XCTAssertTrue(field.isHittable)
    capture("static-search-placement-expanded")
    waitForKeyboard()
    field.typeText("missing")
    XCTAssertEqual(field.value as? String, "missing")
    capture("static-search-before-focused-clear")
    let clear = field.buttons["Clear text"].firstMatch
    XCTAssertTrue(clear.isHittable)
    clear.tap()
    let available = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      (field.exists && field.isHittable) || (search.exists && search.isHittable)
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [available], timeout: 5), .completed)
    capture("static-search-after-focused-clear")
    if !field.isHittable { search.tap() }
    else if !app.keyboards.firstMatch.exists { field.tap() }
    XCTAssertTrue(field.waitForExistence(timeout: 5))
    XCTAssertTrue(field.isHittable)
    waitForKeyboard()
    field.typeText("Garage")
    XCTAssertEqual(field.value as? String, "Garage")
    capture("static-search-fresh-query")
  }

  func testAssetRegionsRecoverAtNormalTextSize() {
    verifyAssetRegionRecovery(captureSuffix: "normal-size")
  }

  func testAssetRegionRecoveryAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
    verifyAssetRegionRecovery(captureSuffix: "accessibility-size")
  }

  private func verifyAssetRegionRecovery(captureSuffix: String) {
    let open = app.buttons["Audit contents recovery"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let contents = app.buttons["Retry contents"]
    let photos = app.buttons["Retry photos"]
    XCTAssertTrue(contents.waitForExistence(timeout: 10))
    XCTAssertTrue(photos.waitForExistence(timeout: 10))
    XCTAssertFalse(app.staticTexts["Nothing here yet"].exists)
    XCTAssertFalse(app.staticTexts["No photos"].exists)
    let scroll = app.scrollViews.firstMatch
    XCTAssertTrue(scroll.exists)
    func reveal(_ element: XCUIElement) {
      func visible() -> Bool {
        let bounds = scroll.frame.intersection(app.frame)
        let top = max(bounds.minY, app.navigationBars.firstMatch.frame.maxY)
        return (element.elementType != .button || element.isHittable) && element.frame.minY >= top && element.frame.maxY <= bounds.maxY
      }
      for _ in 0..<12 where !visible() {
        let top = max(scroll.frame.minY, app.navigationBars.firstMatch.frame.maxY)
        let above = element.frame.minY < top
        let start = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.4 : 0.7))
        let end = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.7 : 0.4))
        start.press(forDuration: 0.05, thenDragTo: end)
      }
      XCTAssertTrue(visible())
    }
    reveal(photos)
    if captureSuffix == "normal-size" {
      XCTAssertGreaterThan(photos.frame.width, photos.frame.height * 1.5, "Short Retry label must not collapse into a narrow oval")
    }
    capture("asset-region-photo-error-\(captureSuffix)")
    photos.tap()
    XCTAssertTrue(photos.waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["No photos"].firstMatch.waitForExistence(timeout: 5))
    reveal(app.staticTexts["No photos"].firstMatch)
    capture("asset-region-photo-recovered-\(captureSuffix)")
    reveal(contents)
    if captureSuffix == "normal-size" {
      XCTAssertGreaterThan(contents.frame.width, contents.frame.height * 1.5, "Short Retry label must not collapse into a narrow oval")
    }
    capture("asset-region-contents-error-\(captureSuffix)")
    XCTAssertTrue(app.staticTexts["Could not load contents."].firstMatch.exists)
    contents.tap()
    let empty = app.staticTexts["Nothing inside yet"].firstMatch
    let recovered = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      !contents.exists && empty.exists
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [recovered], timeout: 15), .completed,
      "Contents retry must disappear and show recovered content after one tap")
    reveal(empty)
    capture("asset-region-recovered-\(captureSuffix)")
    let back = app.navigationBars.buttons.firstMatch
    XCTAssertTrue(back.isHittable)
    back.tap()
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
    XCTAssertTrue(empty.waitForNonExistence(timeout: 5))
    XCTAssertTrue(open.isHittable)
  }

  func testEditTagDisclosureRetainsNormalTextDraft() {
    verifyEditTagDisclosure(captureSuffix: "normal-size")
  }

  func testEditTagDisclosureAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
    verifyEditTagDisclosure(captureSuffix: "accessibility-size", directEntry: true)
  }

  private func verifyEditTagDisclosure(captureSuffix: String, directEntry: Bool = false) {
    let open = app.buttons["Audit Edit tags"]
    if directEntry {
      guard openFixtureURL("audit-edit-tags") else { return }
    } else {
      for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
      XCTAssertTrue(open.isHittable)
      open.tap()
    }
    let scroll = app.scrollViews.containing(.textField, identifier: "Asset name").firstMatch
    XCTAssertTrue(scroll.waitForExistence(timeout: 10))
    let cancel = app.buttons["Cancel"].firstMatch
    func reveal(_ element: XCUIElement, requiresHit: Bool = true) {
      func visible() -> Bool {
        let bounds = scroll.frame.intersection(app.frame)
        return (!requiresHit || element.isHittable) && element.frame.minY >= bounds.minY && element.frame.maxY <= bounds.maxY
      }
      for _ in 0..<18 where !visible() {
        let above = element.frame.minY < scroll.frame.minY
        if above { scroll.swipeDown() } else { scroll.swipeUp() }
      }
      XCTAssertTrue(visible())
      XCTAssertTrue(cancel.isHittable)
      XCTAssertGreaterThanOrEqual(cancel.frame.minY, app.frame.minY)
      XCTAssertLessThanOrEqual(cancel.frame.maxY, app.frame.maxY)
    }
    let chooseTags = app.buttons["Choose tags"].firstMatch
    XCTAssertTrue(chooseTags.waitForExistence(timeout: 5))
    XCTAssertFalse(app.buttons["Select tag Tag 13"].exists)
    reveal(chooseTags)
    if directEntry {
      XCTAssertGreaterThan(app.staticTexts["Tags"].firstMatch.frame.height, 30, "The direct-entry scenario must retain enlarged text")
    }
    XCTAssertFalse(app.textFields["New tag name"].exists)
    let newTag = app.buttons["New tag"].firstMatch
    reveal(newTag); newTag.tap()
    let entry = app.textFields["New tag name"].firstMatch
    XCTAssertTrue(entry.waitForExistence(timeout: 5))
    reveal(entry)
    entry.tap()
    waitForKeyboard(keyLabel: "C")
    entry.typeText("Camping")
    XCTAssertEqual(entry.value as? String, "Camping")
    let dismissKeyboard = app.buttons["Dismiss keyboard"].firstMatch
    XCTAssertTrue(dismissKeyboard.isHittable)
    dismissKeyboard.tap()
    XCTAssertTrue(app.keyboards.firstMatch.waitForNonExistence(timeout: 5))
    XCTAssertFalse(app.buttons["Save"].firstMatch.isEnabled)
    cancel.tap()
    let keep = app.alerts.buttons["Keep editing"]
    XCTAssertTrue(keep.waitForExistence(timeout: 5))
    keep.tap()
    XCTAssertEqual(entry.value as? String, "Camping")
    let explanation = app.staticTexts["Add this tag or clear its name and color before saving."].firstMatch
    XCTAssertTrue(explanation.exists)
    reveal(explanation, requiresHit: false)
    capture("edit-unstaged-tag-retained-\(captureSuffix)")
    let add = app.buttons["Add tag"].firstMatch
    reveal(add)
    add.tap()
    // A cleared SwiftUI field can expose no value while retaining its placeholder.
    let clearedEntry = app.textFields["New tag name"].firstMatch
    XCTAssertTrue(clearedEntry.exists)
    XCTAssertTrue(clearedEntry.value == nil || ["", "New tag"].contains(clearedEntry.value as? String ?? "unexpected"))
    XCTAssertTrue(app.buttons["Remove new tag Camping"].firstMatch.waitForExistence(timeout: 5))
    XCTAssertTrue(app.buttons["Save"].firstMatch.isEnabled)
    reveal(chooseTags)
    chooseTags.tap()
    let visibleDone = app.buttons["Done selecting tags"].firstMatch
    XCTAssertTrue(visibleDone.waitForExistence(timeout: 5))
    XCTAssertTrue(visibleDone.isHittable, "Tag selection must appear above its editor")
    let extra = app.descendants(matching: .any)["Select tag Tag 13"].firstMatch
    XCTAssertTrue(extra.waitForExistence(timeout: 5))
    let tagScroll = app.scrollViews.containing(.any, identifier: "Select tag Tag 13").firstMatch
    for _ in 0..<12 where !extra.isHittable { tagScroll.swipeUp() }
    XCTAssertTrue(extra.isHittable)
    extra.tap()
    let doneTags = app.buttons["Done selecting tags"].firstMatch
    XCTAssertTrue(doneTags.isHittable)
    doneTags.tap()
    XCTAssertTrue(doneTags.waitForNonExistence(timeout: 5))
    XCTAssertTrue(chooseTags.waitForExistence(timeout: 5))
    XCTAssertTrue(app.buttons["Remove new tag Camping"].exists)
    reveal(chooseTags)
    chooseTags.tap()
    XCTAssertTrue(extra.waitForExistence(timeout: 5))
    XCTAssertEqual(extra.value as? String, "checkbox, checked")
    let retained = app.descendants(matching: .any)["Select tag Tag 14"].firstMatch
    XCTAssertEqual(retained.value as? String, "checkbox, checked")
    capture("edit-tags-selection-retained-\(captureSuffix)")
    let cancelTags = app.buttons["Cancel selecting tags"].firstMatch
    cancelTags.tap()
    XCTAssertTrue(cancelTags.waitForNonExistence(timeout: 5))
    XCTAssertTrue(chooseTags.waitForExistence(timeout: 5))
    cancel.tap()
    let discard = app.alerts.buttons["Discard"]
    XCTAssertTrue(discard.waitForExistence(timeout: 5))
    discard.tap()
    XCTAssertTrue(app.textFields["Asset name"].firstMatch.waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
    if !directEntry {
      for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
      XCTAssertTrue(open.isHittable)
    }
    XCTAssertEqual(app.state, .runningForeground)
  }

  func testAssetMoveReturnsToUpdatedDetailAndReopensSelection() {
    let open = app.buttons["Audit asset Edit journey"].firstMatch
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    XCTAssertTrue(app.staticTexts["Camping tent"].firstMatch.waitForExistence(timeout: 10))
    let moveFromDetail = app.buttons["Move"].firstMatch
    for _ in 0..<4 where !moveFromDetail.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(moveFromDetail.isHittable); moveFromDetail.tap()
    let header = app.navigationBars["Move"]
    XCTAssertTrue(header.waitForExistence(timeout: 10))
    let commit = header.buttons["Move"].firstMatch
    let cancel = header.buttons["Cancel"].firstMatch
    XCTAssertTrue(commit.isHittable); XCTAssertTrue(cancel.isHittable)
    XCTAssertFalse(commit.isEnabled)
    capture("asset-move-journey-picker")
    let garage = app.descendants(matching: .any)["Choose destination Garage"].firstMatch
    XCTAssertTrue(garage.waitForExistence(timeout: 10)); XCTAssertTrue(garage.isHittable)
    garage.tap()
    XCTAssertEqual(garage.value as? String, "Selected"); XCTAssertTrue(commit.isEnabled)

    capture("asset-move-journey-selected")
    commit.tap()
    XCTAssertTrue(header.waitForNonExistence(timeout: 10))
    let location = app.buttons["Open location Garage"].firstMatch
    for _ in 0..<4 where !location.isHittable { app.scrollViews.firstMatch.swipeDown() }
    XCTAssertTrue(location.waitForExistence(timeout: 10)); XCTAssertTrue(location.isHittable)
    capture("asset-move-journey-updated-detail")
    for _ in 0..<4 where !moveFromDetail.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(moveFromDetail.isHittable); moveFromDetail.tap()
    XCTAssertTrue(header.waitForExistence(timeout: 10))
    XCTAssertTrue(garage.waitForExistence(timeout: 10)); XCTAssertEqual(garage.value as? String, "Selected")

    XCTAssertFalse(commit.isEnabled)
    capture("asset-move-journey-reopened")
    XCTAssertTrue(cancel.isHittable); cancel.tap()
    XCTAssertTrue(header.waitForNonExistence(timeout: 5))
    XCTAssertTrue(location.exists)
  }

  func testAssetEditSavesAndReturnsToUpdatedDetail() {
    let open = app.buttons["Audit asset Edit journey"].firstMatch
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    XCTAssertTrue(app.staticTexts["Camping tent"].firstMatch.waitForExistence(timeout: 10))
    let edit = app.buttons["Edit"].firstMatch
    XCTAssertTrue(edit.waitForExistence(timeout: 5)); XCTAssertTrue(edit.isHittable)
    capture("asset-edit-journey-detail-before")
    edit.tap()
    let name = app.textFields["Asset name"].firstMatch
    XCTAssertTrue(name.waitForExistence(timeout: 10))
    let header = app.navigationBars["Edit asset"]
    XCTAssertTrue(header.waitForExistence(timeout: 5))
    let save = header.buttons["Save"].firstMatch
    let cancel = header.buttons["Cancel"].firstMatch
    XCTAssertTrue(save.isHittable); XCTAssertTrue(cancel.isHittable)
    capture("asset-edit-journey-editor")
    name.tap(); waitForKeyboard(keyLabel: "space")
    name.typeText(" kit")
    XCTAssertEqual(name.value as? String, "Camping tent kit")
    XCTAssertTrue(save.isHittable); XCTAssertTrue(cancel.isHittable)
    capture("asset-edit-journey-keyboard")
    save.tap()
    XCTAssertTrue(name.waitForNonExistence(timeout: 5))
    let updated = app.staticTexts["Camping tent kit"].firstMatch
    XCTAssertTrue(updated.waitForExistence(timeout: 10))
    XCTAssertTrue(edit.isHittable)
    capture("asset-edit-journey-saved-detail")
    edit.tap()
    XCTAssertTrue(name.waitForExistence(timeout: 10))
    XCTAssertEqual(name.value as? String, "Camping tent kit")
    XCTAssertTrue(cancel.isHittable); cancel.tap()
    XCTAssertTrue(name.waitForNonExistence(timeout: 5))
    XCTAssertTrue(updated.isHittable)
    XCTAssertFalse(app.alerts["Discard changes?"].exists)
  }

  func testEditMetadataRecoveryRetainsNormalTextDraft() {
    let open = app.buttons["Audit Edit recovery"].firstMatch
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    let name = app.textFields["Asset name"].firstMatch
    XCTAssertTrue(name.waitForExistence(timeout: 10))
    let types = app.buttons["Retry asset types"].firstMatch
    let tags = app.buttons["Retry tags"].firstMatch
    XCTAssertTrue(types.waitForExistence(timeout: 10))
    XCTAssertTrue(tags.waitForExistence(timeout: 10))
    XCTAssertEqual(name.value as? String, "Audit tent")
    name.tap(); waitForKeyboard(keyLabel: "space")
    name.typeText(" camping kit")
    let expectedName = "Audit tent camping kit"
    XCTAssertEqual(name.value as? String, expectedName)
    let dismiss = app.buttons["Dismiss keyboard"].firstMatch
    XCTAssertTrue(dismiss.isHittable); dismiss.tap()
    XCTAssertTrue(app.keyboards.firstMatch.waitForNonExistence(timeout: 5))
    let scroll = app.scrollViews.containing(.textField, identifier: "Asset name").firstMatch
    func reveal(_ element: XCUIElement) {
      func visible() -> Bool {
        let bounds = scroll.frame.intersection(app.frame)
        return element.isHittable && element.frame.minY >= bounds.minY && element.frame.maxY <= bounds.maxY
      }
      for _ in 0..<12 where !visible() {
        if element.frame.minY < scroll.frame.minY { scroll.swipeDown() } else { scroll.swipeUp() }
      }
      XCTAssertTrue(visible())
    }
    reveal(tags); tags.tap()
    XCTAssertTrue(tags.waitForNonExistence(timeout: 5))
    reveal(types); types.tap()
    XCTAssertTrue(types.waitForNonExistence(timeout: 5))
    reveal(name)
    XCTAssertEqual(name.value as? String, expectedName)
    capture("edit-normal-metadata-recovered-draft")
    let save = app.buttons["Save"].firstMatch
    XCTAssertTrue(save.isEnabled); XCTAssertTrue(save.isHittable); save.tap()
    let failure = app.alerts["Could not save changes"]
    XCTAssertTrue(failure.waitForExistence(timeout: 5)); failure.buttons["OK"].tap()
    XCTAssertEqual(name.value as? String, expectedName)
    XCTAssertTrue(save.isEnabled)
    let cancel = app.buttons["Cancel"].firstMatch
    XCTAssertTrue(cancel.isHittable); cancel.tap()
    let keep = app.alerts.buttons["Keep editing"]
    XCTAssertTrue(keep.waitForExistence(timeout: 5)); keep.tap()
    XCTAssertEqual(name.value as? String, expectedName)
    capture("edit-normal-rejected-save-retained")
    cancel.tap()
    let discard = app.alerts.buttons["Discard"]
    XCTAssertTrue(discard.waitForExistence(timeout: 5)); discard.tap()
    XCTAssertTrue(name.waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    XCTAssertEqual(app.state, .runningForeground)
  }

  func testEditMetadataRecoveryAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
    guard openFixtureURL("audit-edit-recovery") else { return }
    let types = app.buttons["Retry asset types"]
    let tags = app.buttons["Retry tags"]
    XCTAssertTrue(types.waitForExistence(timeout: 10))
    XCTAssertTrue(tags.waitForExistence(timeout: 10))
    let cancel = app.buttons["Cancel"].firstMatch
    XCTAssertTrue(cancel.isHittable)
    let scroll = app.scrollViews.containing(.textField, identifier: "Asset name").firstMatch
    XCTAssertTrue(scroll.exists)
    func reveal(_ element: XCUIElement) {
      func visible() -> Bool {
        let bounds = scroll.frame.intersection(app.frame)
        return element.isHittable && element.frame.minY >= bounds.minY && element.frame.maxY <= bounds.maxY
      }
      for _ in 0..<12 where !visible() {
        let above = element.frame.minY < scroll.frame.minY
        let start = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.4 : 0.7))
        let end = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.7 : 0.4))
        start.press(forDuration: 0.05, thenDragTo: end)
      }
      XCTAssertTrue(visible())
      XCTAssertTrue(cancel.isHittable)
    }
    reveal(types)
    let message = app.staticTexts["Asset types could not be loaded."].firstMatch
    XCTAssertGreaterThan(message.frame.height, 30)
    capture("edit-metadata-types-accessibility-size")
    reveal(tags)
    capture("edit-metadata-tags-accessibility-size")
    tags.tap()
    XCTAssertTrue(tags.waitForNonExistence(timeout: 5))
    reveal(types)
    types.tap()
    XCTAssertTrue(types.waitForNonExistence(timeout: 5))
    let name = app.textFields["Asset name"]
    XCTAssertTrue(name.waitForExistence(timeout: 5))
    XCTAssertEqual(name.value as? String, "Audit tent")
    reveal(name)
    XCTAssertTrue(app.buttons["Cancel"].firstMatch.isHittable)
    capture("edit-metadata-recovered")
  }

  func testNativeChoiceLabelAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    let open = app.buttons["Audit Expiration filters"]
    XCTAssertTrue(open.waitForExistence(timeout: 30))
    for _ in 0..<5 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let label = app.staticTexts["Availability"]
    XCTAssertTrue(label.waitForExistence(timeout: 5))
    for _ in 0..<5 where !label.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(label.isHittable)
    XCTAssertGreaterThan(label.frame.height, 30, "The scenario must actually render enlarged text")
    capture("choice-label-accessibility-size")
    let choice = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose availability")).firstMatch
    for _ in 0..<5 where !choice.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(choice.isHittable)
    XCTAssertGreaterThanOrEqual(choice.frame.minY, label.frame.maxY, "Accessibility text places the menu below its label")
    choice.tap()
    XCTAssertTrue(app.buttons["Available"].waitForExistence(timeout: 5))
    capture("choice-menu-accessibility-size")
  }

  @available(iOS 17.0, *)
  private func auditAccessibility(_ types: XCUIAccessibilityAuditType) throws {
    try app.performAccessibilityAudit(for: types) { issue in
      let element = issue.element?.debugDescription ?? "XCTest did not identify an element"
      let details = XCTAttachment(string: "\(issue.compactDescription)\n\(issue.detailedDescription)\n\(element)")
      details.name = "accessibility-issue-element"
      details.lifetime = .keepAlways
      self.add(details)
      self.capture("accessibility-issue")
      return false
    }
  }

  func testExpirationOverviewAccessibility() throws {
    app.buttons["Audit Expiration filters"].tap()
    XCTAssertTrue(app.buttons["Choose tags"].waitForExistence(timeout: 5))
    XCTAssertTrue(app.buttons["Apply expiration filters"].isHittable)
    capture("expiration-overview-accessibility")
    if #available(iOS 17.0, *) {
      try auditAccessibility([.hitRegion, .sufficientElementDescription, .trait, .dynamicType, .textClipped])
    } else {
      throw XCTSkip("Accessibility auditing requires iOS 17 or later")
    }
  }

  func testExpirationDatePageKeepsBottomActionsReachable() throws {
    app.buttons["Audit Expiration filters"].tap()
    XCTAssertTrue(app.buttons["Choose date range"].waitForExistence(timeout: 5))
    app.buttons["Choose date range"].tap()
    XCTAssertTrue(app.navigationBars["Date range"].waitForExistence(timeout: 5))
    capture("expiration-date-range")
    let date = app.datePickers.firstMatch
    XCTAssertTrue(date.waitForExistence(timeout: 5))
    date.tap()
    capture("expiration-native-calendar")
    let dismiss = app.buttons["PopoverDismissRegion"]
    XCTAssertTrue(dismiss.waitForExistence(timeout: 5))
    let calendar = app.datePickers.containing(.button, identifier: "DatePicker.NextMonth").firstMatch
    XCTAssertTrue(calendar.waitForExistence(timeout: 5))
    let visible = dismiss.frame.intersection(app.frame)
    let popup = calendar.frame.intersection(visible)
    XCTAssertFalse(popup.isEmpty)
    let navigation = app.navigationBars["Date range"]
    XCTAssertTrue(navigation.exists)
    let navigationFrame = navigation.frame.intersection(visible)
    let outside = [
      CGRect(x: visible.minX, y: visible.minY, width: popup.minX - visible.minX, height: visible.height),
      CGRect(x: popup.maxX, y: visible.minY, width: visible.maxX - popup.maxX, height: visible.height),
      CGRect(x: visible.minX, y: visible.minY, width: visible.width, height: popup.minY - visible.minY),
      CGRect(x: visible.minX, y: popup.maxY, width: visible.width, height: visible.maxY - popup.maxY)
    ].map { $0.intersection(navigationFrame) }
      .filter { !$0.isEmpty && !$0.isInfinite && !$0.isNull }
    let commands = navigation.buttons.allElementsBoundByIndex.map { $0.frame }
    let titles = navigation.staticTexts.allElementsBoundByIndex.map { $0.frame }
    let candidates = outside.flatMap { region in
      [0.25, 0.5, 0.75].map { fraction in
        CGPoint(x: region.minX + region.width * fraction, y: region.midY)
      }
    }
    let selectedPoint = candidates.first { point in
      !commands.contains { $0.contains(point) } && !titles.contains { $0.contains(point) }
    }
    let geometry = XCTAttachment(string: "dismiss=\(visible); calendar=\(popup); navigation=\(navigationFrame); commands=\(commands); titles=\(titles); candidates=\(candidates); tap=\(String(describing: selectedPoint))")
    geometry.name = "calendar-dismissal-geometry"
    geometry.lifetime = .keepAlways
    add(geometry)
    let point = try XCTUnwrap(selectedPoint, "No non-command calendar dismissal target in the navigation bar")
    XCTAssertTrue(visible.contains(point))
    XCTAssertTrue(navigationFrame.contains(point))
    XCTAssertFalse(popup.contains(point))
    app.coordinate(withNormalizedOffset: .zero)
      .withOffset(CGVector(dx: point.x - app.frame.minX, dy: point.y - app.frame.minY)).tap()
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: dismiss)], timeout: 5), .completed)
    capture("expiration-calendar-dismissed")
    XCTAssertTrue(navigation.exists, "Dismissing the calendar must leave the Date range page open")
    XCTAssertTrue(app.buttons["Apply expiration filters"].isHittable)
    let back = app.buttons["Cancel or return to filters"]
    XCTAssertTrue(back.isHittable)
    back.tap()
    XCTAssertTrue(app.buttons["Choose date range"].waitForExistence(timeout: 5))
    back.tap()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 5))
  }
  func testExpirationSearchKeepsActionsReachableWithKeyboard() throws {
    app.buttons["Audit Expiration filters"].tap()
    let tags = app.buttons["Choose tags"]
    XCTAssertTrue(tags.waitForExistence(timeout: 5))
    tags.tap()
    let holiday = app.descendants(matching: .any).matching(identifier: "Holiday supplies").firstMatch
    XCTAssertTrue(holiday.waitForExistence(timeout: 5))
    let searchButton = app.buttons["Search"].firstMatch
    XCTAssertTrue(searchButton.waitForExistence(timeout: 5))
    XCTAssertTrue(searchButton.isHittable)
    capture("expiration-search-collapsed")
    searchButton.tap()
    let search = app.searchFields.firstMatch
    XCTAssertTrue(search.waitForExistence(timeout: 5))
    // Search activation must focus the field; a second tap may open its editing menu.
    waitForKeyboard(keyLabel: "t")
    search.typeText("Tools")
    XCTAssertEqual(observePredicate("search-query-timing",
      predicate: NSPredicate(format: "value == %@", "Tools"), object: search),
      .completed, "Native search must retain the complete query")
    XCTAssertTrue(holiday.waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.descendants(matching: .any).matching(identifier: "Tools").firstMatch.exists)
    XCTAssertTrue(app.keyboards.firstMatch.waitForExistence(timeout: 5))
    capture("expiration-search-keyboard")
    let back = app.buttons["Cancel or return to filters"]
    assertFilterActionsClearKeyboard(app.buttons["Apply expiration filters"], back)
    capture("expiration-search-actions-clear-accessory")
    back.tap()
    XCTAssertTrue(app.buttons["Choose tags"].waitForExistence(timeout: 5))
  }
  func testCustomFieldChoicesStayInPlaceAndRetainTargets() {
    let open = app.buttons["Audit field choices"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    let header = app.navigationBars["Native UI audit"]
    XCTAssertTrue(header.exists)
    let type = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose Type.")).firstMatch
    XCTAssertTrue(type.waitForExistence(timeout: 5)); XCTAssertTrue(type.isHittable)
    XCTAssertTrue(app.staticTexts["Type"].exists)
    type.tap()
    let enumChoice = app.buttons["Enum"]
    XCTAssertTrue(enumChoice.waitForExistence(timeout: 5)); enumChoice.tap()
    XCTAssertTrue(app.textFields["New enum option"].waitForExistence(timeout: 5))
    let applies = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose Applies to.")).firstMatch
    for _ in 0..<4 where !applies.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(applies.isHittable)
    XCTAssertTrue(app.staticTexts["Applies to"].exists)
    applies.tap()
    let selectedTypes = app.buttons["Selected asset types"]
    XCTAssertTrue(selectedTypes.waitForExistence(timeout: 5)); selectedTypes.tap()
    XCTAssertTrue(header.exists)
    func target(_ name: String) -> XCUIElement {
      app.descendants(matching: .any).matching(identifier: name).firstMatch
    }
    func assertTargets(_ value: String) {
      let result = app.staticTexts["Selected targets: \(value)"].firstMatch
      XCTAssertTrue(result.waitForExistence(timeout: 5))
      let scroll = app.scrollViews.firstMatch
      func visible() -> Bool {
        let viewport = scroll.frame.intersection(app.frame)
        let bounds = result.frame
        return bounds.width > 0 && bounds.height > 0 && viewport.contains(bounds)
      }
      for _ in 0..<8 where !visible() { scroll.swipeDown() }
      XCTAssertTrue(visible(), "The exact selected state must be readable; it is not a tap target")
    }
    let first = target("Audit type 01")
    for _ in 0..<6 where !first.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(first.isHittable); first.tap()
    assertTargets("type-1")
    let last = target("Audit type 12")
    for _ in 0..<8 where !last.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(last.isHittable); last.tap()
    assertTargets("type-1, type-12")
    for _ in 0..<8 where !first.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(first.isHittable); first.tap()
    assertTargets("type-12")
    XCTAssertTrue(app.staticTexts["Field type: enum"].exists)
    XCTAssertTrue(app.staticTexts["Applicability: custom_asset_types"].exists)
    XCTAssertTrue(header.exists)
    capture("custom-field-retained-target")
    let back = app.buttons["Back to audit menu"]
    for _ in 0..<4 where !back.isHittable { app.scrollViews.firstMatch.swipeDown() }
    XCTAssertTrue(back.isHittable); back.tap()
    XCTAssertTrue(open.waitForExistence(timeout: 5))
  }

  func testEnumOptionEntryRetainsDuplicateAndClearsAcceptedDraft() {
    let open = app.buttons["Audit field choices"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    let type = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose Type.")).firstMatch
    XCTAssertTrue(type.waitForExistence(timeout: 5)); type.tap()
    let enumChoice = app.buttons["Enum"]
    XCTAssertTrue(enumChoice.waitForExistence(timeout: 5)); enumChoice.tap()
    let field = app.textFields["New enum option"].firstMatch
    XCTAssertTrue(field.waitForExistence(timeout: 5))
    field.tap()
    XCTAssertTrue(app.keyboards.firstMatch.waitForExistence(timeout: 5))
    field.typeText("ready")
    XCTAssertEqual(field.value as? String, "ready")
    let add = app.buttons["Add option"].firstMatch
    func addOption() {
      for _ in 0..<4 where !add.isHittable { app.scrollViews.firstMatch.swipeUp() }
      XCTAssertTrue(add.isHittable)
      XCTAssertTrue(app.keyboards.firstMatch.exists, "This acceptance requires Add with the keyboard present")
      XCTAssertFalse(add.frame.intersects(app.keyboards.firstMatch.frame), "Add must clear the keyboard")
      add.tap()
    }
    addOption()
    XCTAssertTrue(app.staticTexts["This option already exists."].waitForExistence(timeout: 5))
    XCTAssertEqual(field.value as? String, "ready")
    capture("enum-duplicate-draft-retained")
    field.tap()
    field.typeText(String(repeating: XCUIKeyboardKey.delete.rawValue, count: 5))
    let clearedDraft = field.value as? String
    XCTAssertTrue(clearedDraft == "" || clearedDraft == "Add option", "Expected cleared draft, got \(String(describing: clearedDraft))")
    field.typeText("Camping kit")
    XCTAssertEqual(field.value as? String, "Camping kit")
    addOption()
    let added = app.buttons["Remove camping-kit"].firstMatch
    XCTAssertTrue(added.waitForExistence(timeout: 5))
    let acceptedDraft = field.value as? String
    XCTAssertTrue(acceptedDraft == "" || acceptedDraft == "Add option", "Expected accepted draft reset, got \(String(describing: acceptedDraft))")
    XCTAssertFalse(app.staticTexts["This option already exists."].exists)
    for _ in 0..<4 where !added.isHittable { app.scrollViews.firstMatch.swipeDown() }
    XCTAssertTrue(added.isHittable); added.tap()
    XCTAssertTrue(added.waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.buttons["Remove ready"].exists)
    capture("enum-new-option-removed-existing-retained")
    let back = app.buttons["Back to audit menu"]
    for _ in 0..<6 where !back.isHittable { app.scrollViews.firstMatch.swipeDown() }
    XCTAssertTrue(back.isHittable); back.tap()
    XCTAssertTrue(open.waitForExistence(timeout: 5))
  }

  func testDraftOptionRemovalPreservesSavedOptions() throws {
    app.buttons["Audit draft options"].tap()
    let remove = app.buttons["Remove draft"]
    XCTAssertTrue(remove.waitForExistence(timeout: 5))
    if !remove.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(remove.isHittable)
    capture("enum-draft-option")
    remove.tap()
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: remove)], timeout: 5), .completed)
    XCTAssertTrue(app.staticTexts["saved · Existing"].exists)
    XCTAssertFalse(app.buttons["Remove saved"].exists)
    capture("enum-draft-option-removed")
  }

  func testUnavailablePhotoExplainsFailureAndKeepsEscapeReachable() {
    let open = app.buttons["Audit unavailable photo"]
    for _ in 0..<8 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    XCTAssertTrue(app.staticTexts["Photo unavailable"].firstMatch.waitForExistence(timeout: 15))
    let retry = app.buttons["Retry photo"]
    XCTAssertTrue(retry.isHittable)
    XCTAssertTrue(app.buttons["Close photo viewer"].isHittable)
    capture("photo-unavailable-recovery")
    retry.tap()
    XCTAssertTrue(app.staticTexts["Photo unavailable"].firstMatch.waitForExistence(timeout: 15))
    XCTAssertTrue(app.buttons["Close photo viewer"].isHittable)
    app.buttons["Close photo viewer"].tap()
    XCTAssertTrue(app.buttons["Close photo viewer"].waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.buttons["Back to audit menu"].isHittable)
  }

  func testPhotoRemovalFailureAppearsAboveViewer() {
    let open = app.buttons["Audit photo removal recovery"]
    for _ in 0..<8 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let remove = app.buttons["Remove photo"]
    XCTAssertTrue(remove.waitForExistence(timeout: 5))
    for attempt in 1...2 {
      XCTAssertTrue(remove.isHittable)
      remove.tap()
      let confirmation = app.alerts["Remove photo?"]
      XCTAssertTrue(confirmation.waitForExistence(timeout: 5))
      confirmation.buttons["Remove"].tap()
      let failure = app.alerts["Could not remove photo"]
      XCTAssertTrue(failure.waitForExistence(timeout: 5))
      XCTAssertTrue(failure.buttons["OK"].isHittable)
      capture("photo-removal-failure-\(attempt)")
      failure.buttons["OK"].tap()
      XCTAssertTrue(failure.waitForNonExistence(timeout: 5))
      XCTAssertTrue(remove.isEnabled)
      XCTAssertTrue(app.buttons["Close photo viewer"].isHittable)
    }
    capture("photo-retained-after-retry")
    app.buttons["Close photo viewer"].tap()
    XCTAssertTrue(app.staticTexts["Removal attempts: 2"].waitForExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["Photos remaining: 1"].exists)
  }

  private func verifyAddressEntry(_ mode: String, withoutAccessory: Bool = false) {
    let open = app.buttons[withoutAccessory ? "Audit input without accessory" : "Audit \(mode) input"]
    for _ in 0..<8 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let input = app.textFields["Audit \(mode) address"]
    XCTAssertTrue(input.waitForExistence(timeout: 5))
    revealComparisonInput(input)
    XCTAssertTrue(input.isHittable)
    input.tap()
    waitForKeyboard()
    if withoutAccessory { XCTAssertFalse(app.buttons["Dismiss keyboard"].exists) }
    input.typeText("https://example.invalid")
    capture("\(mode)-address-entry\(withoutAccessory ? "-without-accessory" : "")")
    let enteredValue = input.value as? String
    let observedExpectedValue = app.staticTexts["Observed \(mode) input: https://example.invalid"].waitForExistence(timeout: 5)
    if mode != "system" { captureInputEvents(mode) }
    XCTAssertEqual(enteredValue, "https://example.invalid")
    XCTAssertTrue(observedExpectedValue)
  }

  private func revealComparisonInput(_ input: XCUIElement) {
    let scroll = app.scrollViews.firstMatch
    for _ in 0..<8 {
      let bounds = scroll.frame.intersection(app.frame)
      let top = max(bounds.minY, app.navigationBars.firstMatch.frame.maxY)
      if input.isHittable && input.frame.minY >= top && input.frame.maxY <= bounds.maxY { return }
      let above = input.frame.minY < top
      scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.4 : 0.7))
        .press(forDuration: 0.05, thenDragTo: scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.7 : 0.4)))
    }
    XCTFail("Comparison input must be fully visible before typing")
  }

  private func verifyOrdinaryTextEntry(_ mode: String, paced: Bool = false) {
    let open = app.buttons["Audit \(mode) input"]
    for _ in 0..<8 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let input = mode == "multiline" ? app.textViews["Audit \(mode) text"] : app.textFields["Audit \(mode) text"]
    XCTAssertTrue(input.waitForExistence(timeout: 5))
    revealComparisonInput(input)
    XCTAssertTrue(input.isHittable)
    input.tap()
    waitForKeyboard()
    if mode.hasSuffix("no-accessory") { XCTAssertFalse(app.buttons["Dismiss keyboard"].exists) }
    if paced {
      for character in "Native draft name" { input.typeText(String(character)) }
    } else {
      input.typeText("Native draft name")
    }
    capture("\(mode)-ordinary-text-entry\(paced ? "-paced" : "")")
    let enteredValue = input.value as? String
    let observedExpectedValue = app.staticTexts["Observed \(mode) input: Native draft name"].waitForExistence(timeout: 5)
    if mode != "native-default" { captureInputEvents(mode) }
    XCTAssertEqual(enteredValue, "Native draft name")
    XCTAssertTrue(observedExpectedValue)
  }

  private func captureInputEvents(_ mode: String) {
    let captureEvents = app.buttons["Capture input events"]
    guard captureEvents.isHittable else {
      capture("input-event-capture-unreachable-\(mode)")
      return
    }
    captureEvents.tap()
    let trace = app.staticTexts["audit-input-event-trace"]
    guard trace.waitForExistence(timeout: 5) else {
      XCTFail("The input trace command must publish its recorded events")
      return
    }
    let attachment = XCTAttachment(string: trace.label)
    attachment.name = "input-events-\(mode)"
    attachment.lifetime = .keepAlways
    add(attachment)
  }

  func testOrdinarySingleLineTextEntry() { verifyOrdinaryTextEntry("plain") }
  func testNativeDefaultAssistedTextEntryDiagnostic() { verifyOrdinaryTextEntry("native-default") }
  func testOrdinaryMultilineTextEntry() { verifyOrdinaryTextEntry("multiline") }
  func testOrdinaryControlledTextEntry() { verifyOrdinaryTextEntry("plain-controlled") }
  func testControlledTextEntryWithoutAssistance() { verifyOrdinaryTextEntry("plain-controlled-no-assistance") }
  func testControlledTextEntryWithoutAccessory() { verifyOrdinaryTextEntry("plain-controlled-no-accessory") }
  func testOrdinaryTextEntryWithoutAssistance() { verifyOrdinaryTextEntry("plain-no-assistance") }
  func testOrdinaryTextEntryWithoutAccessory() { verifyOrdinaryTextEntry("plain-no-accessory") }
  func testPacedControlledTextEntryDiagnostic() { verifyOrdinaryTextEntry("plain-controlled", paced: true) }
  func testPacedUncontrolledTextEntryDiagnostic() { verifyOrdinaryTextEntry("plain", paced: true) }

  func testSeededAddressWithoutKeyboardAccessory() { verifyAddressEntry("uncontrolled", withoutAccessory: true) }

  func testControlledAddressEntry() { verifyAddressEntry("controlled") }
  func testUncontrolledAddressEntry() { verifyAddressEntry("uncontrolled") }
  func testSystemAddressEntry() { verifyAddressEntry("system") }

  private func waitForExactEnteredText(_ text: String, in field: XCUIElement) {
    let started = Date()
    let entered = XCTNSPredicateExpectation(predicate: NSPredicate(format: "value == %@", text), object: field)
    let result = XCTWaiter.wait(for: [entered], timeout: 30)
    let timing = XCTAttachment(string: "Exact text observation elapsed: \(Date().timeIntervalSince(started)) seconds; result: \(result.rawValue)")
    timing.name = "selection-text-observation-timing"
    timing.lifetime = .keepAlways
    add(timing)
    XCTAssertEqual(result, .completed)
    XCTAssertEqual(field.value as? String, text)
  }


  func testAddDestinationSelectionPreservesDraftAndRecoversCreation() {
    guard openFixtureURL("audit-add-destination") else { return }
    let name = app.textFields["Asset name"].firstMatch
    XCTAssertTrue(name.waitForExistence(timeout: 10))
    let form = app.scrollViews.containing(.textField, identifier: "Asset name").firstMatch
    func reveal(_ element: XCUIElement, in scroll: XCUIElement) {
      for _ in 0..<12 where !element.isHittable { scroll.swipeUp() }
      XCTAssertTrue(element.isHittable)
    }
    func search(_ query: String) {
      let field = app.searchFields.firstMatch
      if !field.exists {
        let button = app.buttons["Search"].firstMatch
        XCTAssertTrue(button.waitForExistence(timeout: 5)); button.tap()
      }
      XCTAssertTrue(field.waitForExistence(timeout: 5)); XCTAssertTrue(field.isHittable)
      // Verify native input through the entered text and matching results below.
      // A separate keyboard-key snapshot can itself exhaust the wait budget.
      field.tap(); field.typeText(query)
      let entered = NSPredicate(format: "value == %@", query)
      XCTAssertEqual(observePredicate("add-search-exact-query", predicate: entered, object: field, immediately: true), .completed)
      dismissNativeSearchKeyboardIfNeeded()
      XCTAssertEqual(field.value as? String, query)
    }
    name.tap(); waitForKeyboard(keyLabel: "T"); name.typeText("Tent")
    waitForExactEnteredText("Tent", in: name)
    app.buttons["Dismiss keyboard"].firstMatch.tap()
    let choose = app.buttons["Choose destination"].firstMatch
    reveal(choose, in: form); choose.tap()
    let topLevel = app.descendants(matching: .any)["Choose inventory top level"].firstMatch
    XCTAssertTrue(topLevel.waitForExistence(timeout: 5))
    XCTAssertTrue(topLevel.isHittable)
    let pickerHeader = app.navigationBars["Put in"]
    XCTAssertGreaterThanOrEqual(topLevel.frame.minY, pickerHeader.frame.maxY)
    capture("add-destination-entry")
    search("Shelf 14")

    let shelf = app.descendants(matching: .any)["Choose destination Shelf 14"].firstMatch
    XCTAssertTrue(shelf.waitForExistence(timeout: 5))
    let closeSearch = app.navigationBars.buttons["Close"].firstMatch
    if closeSearch.exists && closeSearch.isHittable { closeSearch.tap() }
    let cancel = app.buttons["Cancel location selection"].firstMatch
    XCTAssertTrue(cancel.waitForExistence(timeout: 5))
    XCTAssertTrue(cancel.isHittable)
    cancel.tap(); XCTAssertTrue(cancel.waitForNonExistence(timeout: 5))
    XCTAssertEqual(choose.value as? String, "Garage")
    XCTAssertEqual(name.value as? String, "Tent")
    reveal(choose, in: form); choose.tap(); search("Shelf 14")
    XCTAssertTrue(shelf.waitForExistence(timeout: 5)); XCTAssertTrue(shelf.isHittable); shelf.tap()
    XCTAssertTrue(cancel.waitForNonExistence(timeout: 5))
    XCTAssertEqual(choose.value as? String, "Shelf 14")
    app.buttons["Save item"].firstMatch.tap()
    XCTAssertTrue(app.staticTexts["Could not save asset"].firstMatch.waitForExistence(timeout: 5))
    if app.alerts.buttons["OK"].firstMatch.exists { app.alerts.buttons["OK"].firstMatch.tap() }
    reveal(choose, in: form); choose.tap(); search("Audit shed")
    let destinations = app.scrollViews.containing(.any, identifier: "Choose inventory top level").firstMatch
    let newPlace = app.buttons["New place"].firstMatch
    let ready = XCTNSPredicateExpectation(predicate: NSPredicate(format: "enabled == true"), object: newPlace)
    XCTAssertEqual(XCTWaiter.wait(for: [ready], timeout: 5), .completed)
    reveal(newPlace, in: destinations); newPlace.tap()
    XCTAssertEqual(app.textFields["New place name"].firstMatch.value as? String, "Audit shed")
    XCTAssertTrue(app.navigationBars["New place"].exists)
    let cancelCreation = app.navigationBars["New place"].buttons["Cancel new place"].firstMatch
    XCTAssertTrue(cancelCreation.isHittable)
    let creationName = app.textFields["New place name"].firstMatch
    XCTAssertTrue(creationName.isHittable)
    XCTAssertGreaterThanOrEqual(creationName.frame.minY, app.navigationBars["New place"].frame.maxY)
    XCTAssertLessThanOrEqual(creationName.frame.minY - app.navigationBars["New place"].frame.maxY, 96, "Creation form must not inherit search/header spacing twice")
    capture("add-destination-creation-entry")
    cancelCreation.tap()
    XCTAssertTrue(app.navigationBars["Put in"].waitForExistence(timeout: 5))
    XCTAssertTrue(newPlace.isHittable); newPlace.tap()
    XCTAssertTrue(app.navigationBars["New place"].waitForExistence(timeout: 5))
    XCTAssertEqual(app.textFields["New place name"].firstMatch.value as? String, "Audit shed")
    let create = app.navigationBars["New place"].buttons["Create place"].firstMatch
    XCTAssertTrue(create.isHittable); create.tap()
    XCTAssertTrue(app.staticTexts["Place creation unavailable. Try again."].firstMatch.waitForExistence(timeout: 5))
    capture("add-destination-creation-retry")
    XCTAssertTrue(create.isHittable); create.tap()
    XCTAssertTrue(app.navigationBars["New place"].waitForNonExistence(timeout: 5))
    XCTAssertTrue(create.waitForNonExistence(timeout: 5))
    let returned = XCTNSPredicateExpectation(predicate: NSPredicate(format: "hittable == true"), object: choose)
    XCTAssertEqual(XCTWaiter.wait(for: [returned], timeout: 5), .completed)

    XCTAssertEqual(choose.value as? String, "Audit shed")
    XCTAssertEqual(name.value as? String, "Tent")
    capture("add-destination-created-and-returned")
  }

  func testAddTagSelectionRetainsDraftAcrossCancelAndSaveFailure() {
    guard openFixtureURL("audit-add-header") else { return }
    let name = app.textFields["Asset name"].firstMatch
    XCTAssertTrue(name.waitForExistence(timeout: 10))
    let form = app.scrollViews.containing(.textField, identifier: "Asset name").firstMatch
    func reveal(_ element: XCUIElement, in scroll: XCUIElement) {
      for _ in 0..<16 where !element.isHittable { scroll.swipeUp() }
      XCTAssertTrue(element.isHittable)
    }
    name.tap(); waitForKeyboard(keyLabel: "T"); name.typeText("Tent")
    let typed = XCTNSPredicateExpectation(predicate: NSPredicate(format: "value == %@", "Tent"), object: name)
    XCTAssertEqual(XCTWaiter.wait(for: [typed], timeout: 5), .completed)
    app.buttons["Dismiss keyboard"].firstMatch.tap()
    let details = app.buttons["More details"].firstMatch
    reveal(details, in: form); details.tap()
    let choose = app.buttons["Choose tags"].firstMatch
    reveal(choose, in: form); choose.tap()
    let choice = app.descendants(matching: .any)["Select tag Tag 13"].firstMatch
    XCTAssertTrue(choice.waitForExistence(timeout: 5))
    var choices = app.scrollViews.containing(.any, identifier: "Select tag Tag 13").firstMatch
    reveal(choice, in: choices); choice.tap()
    let selected = XCTNSPredicateExpectation(predicate: NSPredicate(format: "value == %@", "checkbox, checked"), object: choice)
    XCTAssertEqual(XCTWaiter.wait(for: [selected], timeout: 5), .completed)
    let cancel = app.buttons["Cancel selecting tags"].firstMatch
    cancel.tap(); XCTAssertTrue(cancel.waitForNonExistence(timeout: 5))
    reveal(choose, in: form); choose.tap()
    XCTAssertTrue(choice.waitForExistence(timeout: 5))
    choices = app.scrollViews.containing(.any, identifier: "Select tag Tag 13").firstMatch
    reveal(choice, in: choices)
    XCTAssertNotEqual(choice.value as? String, "checkbox, checked")
    choice.tap()
    let done = app.buttons["Done selecting tags"].firstMatch
    done.tap(); XCTAssertTrue(done.waitForNonExistence(timeout: 5))
    XCTAssertEqual(name.value as? String, "Tent")
    let save = app.buttons["Save item"].firstMatch
    XCTAssertTrue(save.isEnabled); save.tap()
    XCTAssertTrue(app.staticTexts["Could not save asset"].firstMatch.waitForExistence(timeout: 15))
    if app.alerts.buttons["OK"].firstMatch.exists { app.alerts.buttons["OK"].firstMatch.tap() }
    reveal(choose, in: form); choose.tap()
    XCTAssertTrue(choice.waitForExistence(timeout: 5))
    XCTAssertEqual(choice.value as? String, "checkbox, checked")
    capture("add-tag-selection-after-rejected-save")
    cancel.tap(); XCTAssertTrue(cancel.waitForNonExistence(timeout: 5))
    XCTAssertEqual(name.value as? String, "Tent")
  }

  func testAddRetainsUnfinishedTagAcrossDetailsDisclosure() {
    let open = app.buttons["Audit Add configured header"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let name = app.textFields["Asset name"].firstMatch
    XCTAssertTrue(name.waitForExistence(timeout: 10))
    let scroll = app.scrollViews.containing(.textField, identifier: "Asset name").firstMatch
    XCTAssertTrue(scroll.exists)
    func reveal(_ element: XCUIElement, requiresHit: Bool = true) {
      for attempt in 0...18 {
        let bounds = scroll.frame.intersection(app.frame)
        let top = max(bounds.minY, app.navigationBars["Add item"].frame.maxY)
        let frame = element.frame
        let contained = frame.minY >= top && frame.maxY <= bounds.maxY
        if contained && (!requiresHit || element.isHittable) { return }
        guard attempt < 18 else {
          XCTFail("Target must be visible below the header and inside the scroll viewport")
          return
        }
        let above = frame.minY < top
        let start = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.4 : 0.7))
        let end = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.7 : 0.4))
        start.press(forDuration: 0.05, thenDragTo: end)
      }
    }
    func dismissKeyboard() {
      let dismiss = app.buttons["Dismiss keyboard"].firstMatch
      XCTAssertTrue(dismiss.isHittable)
      dismiss.tap()
      XCTAssertTrue(app.keyboards.firstMatch.waitForNonExistence(timeout: 5))
    }
    reveal(name)
    name.tap()
    waitForKeyboard(keyLabel: "T")
    name.typeText("Tent")
    waitForExactEnteredText("Tent", in: name)
    dismissKeyboard()
    let save = app.buttons["Save item"].firstMatch
    XCTAssertTrue(save.isEnabled)
    let details = app.buttons["More details"].firstMatch
    reveal(details)
    details.tap()
    let entry = app.textFields["New tag name"].firstMatch
    XCTAssertFalse(entry.exists)
    let newTag = app.buttons["New tag"].firstMatch
    reveal(newTag); newTag.tap()
    XCTAssertTrue(entry.waitForExistence(timeout: 5))
    let cancelNewTag = app.buttons["Cancel new tag"].firstMatch
    reveal(cancelNewTag); cancelNewTag.tap()
    XCTAssertTrue(entry.waitForNonExistence(timeout: 5))
    XCTAssertTrue(save.isEnabled)
    reveal(newTag); newTag.tap()
    XCTAssertTrue(entry.waitForExistence(timeout: 5))
    reveal(entry)
    entry.tap()
    waitForKeyboard(keyLabel: "C")
    entry.typeText("Camping")
    waitForExactEnteredText("Camping", in: entry)
    dismissKeyboard()
    XCTAssertFalse(save.isEnabled)
    reveal(details)
    details.tap()
    XCTAssertTrue(entry.waitForNonExistence(timeout: 5))
    let guidance = app.staticTexts["Open More details to add or clear the unfinished tag before saving."].firstMatch
    XCTAssertTrue(guidance.waitForExistence(timeout: 5))
    reveal(guidance, requiresHit: false)
    capture("add-unstaged-tag-collapsed")
    reveal(details)
    details.tap()
    let retainedTag = XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == true AND value == %@", "Camping"), object: entry)
    XCTAssertEqual(XCTWaiter.wait(for: [retainedTag], timeout: 5), .completed)
    XCTAssertEqual(entry.value as? String, "Camping")
    let add = app.buttons["Add tag"].firstMatch
    reveal(add)
    add.tap()
    let staged = app.buttons["Remove new tag Camping"].firstMatch
    let stageComplete = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      guard staged.exists, save.exists else { return false }
      if !entry.exists { return save.isEnabled }
      let value = entry.value
      return (value == nil || value as? String == "" || value as? String == "New tag") && save.isEnabled
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [stageComplete], timeout: 5), .completed,
      "Staging must retain the tag, clear or close creation and enable Save")
    capture("add-tag-staged")
    let clear = app.buttons["Clear draft"].firstMatch
    reveal(clear)
    clear.tap()
    XCTAssertFalse(save.isEnabled)
    let close = app.buttons["Close Add"].firstMatch
    XCTAssertTrue(close.isHittable)
    close.tap()
    XCTAssertTrue(open.waitForExistence(timeout: 5))
  }

  func testAddDraftRetainsTextAndRecoversAfterRejectedSave() {
    verifyAddDraft(route: "audit-add")
  }

  func testAddDraftInNavigationStack() {
    verifyAddDraft(route: "audit-add-push")
  }

  func testAddDraftWithHeaderConfiguredBeforePresentation() {
    verifyAddDraft(route: "audit-add-header")
  }

  func testAddPhotoPreviewPagingAndDraftRemoval() {
    guard openFixtureURL("audit-add-header") else { return }
    let add = app.buttons["Add photos"].firstMatch
    XCTAssertTrue(add.waitForExistence(timeout: 10))
    XCTAssertTrue(add.isHittable)
    add.tap()
    let library = app.buttons["Choose from Library"].firstMatch
    XCTAssertTrue(library.waitForExistence(timeout: 5))
    library.tap()
    let secondThumbnail = app.buttons["Remove photo 2"].firstMatch
    XCTAssertTrue(secondThumbnail.waitForExistence(timeout: 5))
    func openFirstPhoto(count: Int) {
      let preview = app.descendants(matching: .any).matching(
        NSPredicate(format: "value == %@", "1 of \(count)")).firstMatch
      XCTAssertTrue(preview.waitForExistence(timeout: 5))
      XCTAssertTrue(preview.isHittable)
      preview.tap()
      XCTAssertTrue(app.buttons["Close photo viewer"].waitForExistence(timeout: 5))
      XCTAssertTrue(app.staticTexts["audit-draft-photo-1.png"].exists)
    }
    func confirmRemoval() {
      let remove = app.buttons["Remove photo"]
      XCTAssertTrue(remove.isHittable)
      remove.tap()
      XCTAssertTrue(app.alerts["Remove photo?"].waitForExistence(timeout: 5))
    }
    openFirstPhoto(count: 2)
    let next = app.buttons["Next photo"]
    XCTAssertTrue(next.isHittable)
    next.tap()
    XCTAssertTrue(app.staticTexts["audit-draft-photo-2.png"].waitForExistence(timeout: 5))
    confirmRemoval()
    app.alerts["Remove photo?"].buttons["Cancel"].tap()
    XCTAssertTrue(app.staticTexts["audit-draft-photo-2.png"].exists)
    XCTAssertTrue(app.staticTexts["2 of 2"].exists)
    confirmRemoval()
    app.alerts["Remove photo?"].buttons["Remove"].tap()
    XCTAssertTrue(app.staticTexts["audit-draft-photo-1.png"].waitForExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["1 of 1"].exists)
    XCTAssertFalse(next.exists)
    capture("add-photo-preview-surviving-draft")
    app.buttons["Close photo viewer"].tap()
    XCTAssertTrue(app.buttons["Remove photo 1"].waitForExistence(timeout: 5))
    XCTAssertFalse(secondThumbnail.exists)
    openFirstPhoto(count: 1)
    confirmRemoval()
    app.alerts["Remove photo?"].buttons["Remove"].tap()
    XCTAssertTrue(app.buttons["Close photo viewer"].waitForNonExistence(timeout: 5))
    XCTAssertTrue(add.isHittable)
    XCTAssertTrue(app.textFields["Asset name"].exists)
    XCTAssertFalse(app.buttons["Remove photo 1"].exists)
    capture("add-photo-last-removal-returns-to-draft")
    app.buttons["Close Add"].tap()
    XCTAssertTrue(app.textFields["Asset name"].waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
    let root = app.buttons["Audit Browse filters"]
    XCTAssertTrue(root.waitForExistence(timeout: 5))
    for _ in 0..<12 where !root.isHittable {
      let viewport = app.scrollViews.firstMatch
      if root.frame.maxY > viewport.frame.maxY { viewport.swipeUp() }
      else { viewport.swipeDown() }
    }
    XCTAssertTrue(root.isHittable)
    XCTAssertEqual(app.state, .runningForeground)
  }

  private func openFixtureURL(_ route: String) -> Bool {
    if #available(iOS 16.4, *) {
      app.open(URL(string: "stuffstash:///\(route)")!)
      return true
    }
    XCTFail("Native fixture URL entry requires iOS 16.4 or later")
    return false
  }

  private func verifyAddDraft(route: String) {
    // Synthetic menu scrolling previously delivered a tap to an unrelated fixture.
    // Production Home entry has its own tests; isolate this presentation comparison.
    guard openFixtureURL(route) else { return }
    let name = app.textFields["Asset name"]
    XCTAssertTrue(name.waitForExistence(timeout: 10))
    XCTAssertTrue(app.navigationBars["Add item"].exists)
    capture("add-native-header")
    name.tap()
    waitForKeyboard()
    name.typeText("Native draft name")
    let exactName = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      name.value as? String == "Native draft name"
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [exactName], timeout: 5), .completed,
      "Native input must retain the exact typed name before saving")
    let save = app.buttons["Save item"]
    let close = app.buttons["Close Add"]
    XCTAssertTrue(save.isHittable)
    save.tap()
    XCTAssertFalse(save.isEnabled)
    XCTAssertFalse(close.isEnabled)
    let rejected = app.staticTexts["Rejected draft: Native draft name"].firstMatch
    XCTAssertTrue(rejected.waitForExistence(timeout: 10))
    XCTAssertEqual(name.value as? String, "Native draft name")
    XCTAssertTrue(save.isEnabled)
    XCTAssertTrue(close.isEnabled)
    let errorHeading = app.staticTexts["Could not save asset"].firstMatch
    XCTAssertTrue(errorHeading.waitForExistence(timeout: 5))
    let navigationBar = app.navigationBars["Add item"]
    XCTAssertGreaterThanOrEqual(errorHeading.frame.minY, navigationBar.frame.maxY,
      "The entire error heading must remain below the native navigation bar")
    XCTAssertLessThanOrEqual(errorHeading.frame.maxY, rejected.frame.minY)
    capture("add-rejected-draft-retained")
    close.tap()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 5))
  }

  func testOnboardingSubmitsTheCompleteNativeAddress() {
    let open = app.buttons["Audit onboarding submission"]
    for _ in 0..<4 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let address = app.textFields["Server address"]
    XCTAssertTrue(address.waitForExistence(timeout: 5))
    address.tap()
    waitForKeyboard()
    address.typeText("https://example.invalid")
    XCTAssertEqual(address.value as? String, "https://example.invalid")
    let dismissKeyboard = app.buttons["Dismiss keyboard"]
    XCTAssertTrue(dismissKeyboard.isHittable)
    dismissKeyboard.tap()
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: app.keyboards.firstMatch)], timeout: 5), .completed)
    let connect = app.buttons["Connect and sign in"]
    for _ in 0..<3 where !connect.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(connect.isHittable)
    connect.tap()
    XCTAssertTrue(app.staticTexts["Submitted address: https://example.invalid"].waitForExistence(timeout: 5))
    capture("onboarding-complete-address-submission")
  }

  func testOnboardingConnectRemainsReachableWithKeyboardOpen() {
    let open = app.buttons["Audit onboarding submission"]
    for _ in 0..<4 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    let address = app.textFields["Server address"]
    XCTAssertTrue(address.waitForExistence(timeout: 5))
    XCTAssertFalse(app.navigationBars["Native UI audit"].isHittable, "Production setup has no navigation header")
    address.tap()
    waitForKeyboard()
    address.typeText("https://example.invalid")
    XCTAssertEqual(address.value as? String, "https://example.invalid")
    let keyboard = app.keyboards.firstMatch
    let connect = app.buttons["Connect and sign in"].firstMatch
    let scroll = app.scrollViews.containing(.textField, identifier: "Server address").firstMatch
    XCTAssertTrue(scroll.exists)
    func usableViewport() -> CGRect {
      let bounds = scroll.frame.intersection(app.frame)
      let header = app.navigationBars.firstMatch
      let top = header.exists && header.frame.intersects(bounds) ? max(bounds.minY, header.frame.maxY) : bounds.minY
      let bottom = min(bounds.maxY, keyboard.frame.minY)
      guard !bounds.isEmpty, bottom > top else { return .zero }
      return CGRect(x: bounds.minX, y: top, width: bounds.width, height: bottom - top)
    }
    func fullyVisible() -> Bool {
      guard keyboard.exists, connect.exists, connect.isEnabled, connect.isHittable else { return false }
      let visible = usableViewport()
      return !visible.isEmpty && connect.frame.minY >= visible.minY && connect.frame.maxY <= visible.maxY
        && connect.frame.minX >= visible.minX && connect.frame.maxX <= visible.maxX
    }
    for _ in 0..<3 where !fullyVisible() {
      XCTAssertTrue(keyboard.exists, "This check must keep the keyboard open")
      let origin = app.coordinate(withNormalizedOffset: .zero)
      let viewport = usableViewport()
      XCTAssertGreaterThan(viewport.height, 32, "A drag requires usable space between header and keyboard")
      let bottom = viewport.maxY - 16
      let x = viewport.midX - app.frame.minX
      let start = origin.withOffset(CGVector(dx: x, dy: bottom - app.frame.minY))
      let end = origin.withOffset(CGVector(dx: x, dy: max(viewport.minY + 16, bottom - 220) - app.frame.minY))
      start.press(forDuration: 0.1, thenDragTo: end)
    }
    XCTAssertTrue(fullyVisible(), "The entire Connect command must clear the open keyboard")
    capture("onboarding-connect-keyboard-clearance")
    connect.tap()
    XCTAssertTrue(app.staticTexts["Submitted address: https://example.invalid"].waitForExistence(timeout: 5))
    capture("onboarding-connect-keyboard-submitted")
  }

  func testOnboardingKeyboardGoSubmitsCompleteAddress() {
    let open = app.buttons["Audit onboarding submission"]
    for _ in 0..<4 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let address = app.textFields["Server address"]
    XCTAssertTrue(address.waitForExistence(timeout: 5))
    address.tap()
    waitForKeyboard()
    address.typeText("https://example.invalid")
    XCTAssertEqual(address.value as? String, "https://example.invalid")
    let go = app.keyboards.buttons["Go"]
    XCTAssertTrue(go.isHittable)
    go.tap()
    XCTAssertTrue(app.staticTexts["Submitted address: https://example.invalid"].waitForExistence(timeout: 5))
    capture("onboarding-keyboard-go-submission")
  }

  func testHomeViewerHeaderUsesSpaceWithoutOverlappingProfile() {
    guard openFixtureURL("audit-home-header?viewer=true") else { return }
    let profile = app.buttons["Open account and settings"]
    let selector = app.buttons["Current inventory Main inventory with a long household name, tenant Audit home. Switch inventory"]
    XCTAssertTrue(selector.waitForExistence(timeout: 10))
    XCTAssertTrue(selector.isHittable)
    XCTAssertTrue(profile.isHittable)
    XCTAssertFalse(app.buttons["Add an asset"].exists)
    XCTAssertFalse(app.buttons["Notifications, 2 unread"].exists)
    XCTAssertGreaterThan(selector.frame.width, 220)
    XCTAssertLessThanOrEqual(selector.frame.maxX, profile.frame.minX)
    XCTAssertTrue(app.frame.contains(selector.frame))
    capture("home-viewer-header-available-space")
    profile.tap()
    XCTAssertTrue(app.staticTexts["Header Profile destination"].waitForExistence(timeout: 5))
  }

  func testHomeHeaderKeepsAllActionsAboveScrollingContent() {
    let open = app.buttons["Audit Home header"]
    XCTAssertTrue(open.waitForExistence(timeout: 5))
    let menu = app.scrollViews.containing(.button, identifier: "Audit Home header").firstMatch
    for _ in 0..<6 where !open.isHittable { menu.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let add = app.buttons["Add an asset"]
    let notifications = app.buttons["Notifications, 2 unread"]
    let profile = app.buttons["Open account and settings"]
    let selector = app.buttons["Current inventory Main inventory with a long household name, tenant Audit home. Switch inventory"]
    XCTAssertTrue(add.waitForExistence(timeout: 10))
    let actions = [add, notifications, profile]
    func verifyActions() {
      for action in actions {
        XCTAssertTrue(action.isHittable)
        // AX bounds describe layout, not UIKit's delivered touch region.
        // The three independent action probes retain near-edge touch checks.
        XCTAssertGreaterThan(action.frame.width, 0)
        XCTAssertGreaterThan(action.frame.height, 0)
        XCTAssertTrue(app.frame.contains(action.frame))
      }
      XCTAssertTrue(selector.isHittable)
      XCTAssertLessThanOrEqual(selector.frame.maxX, add.frame.minX)
      XCTAssertLessThanOrEqual(add.frame.maxX, notifications.frame.minX)
      XCTAssertLessThanOrEqual(notifications.frame.maxX, profile.frame.minX)
    }
    verifyActions()
    let headerTop = add.frame.minY
    let recent = app.staticTexts["Recently changed"].firstMatch
    XCTAssertTrue(recent.exists)
    let contentTop = recent.frame.minY
    capture("home-header-before-scroll")
    app.scrollViews.containing(.staticText, identifier: "Recently changed").firstMatch.swipeUp()
    XCTAssertLessThan(recent.frame.minY, contentTop - 20)
    verifyActions()
    XCTAssertEqual(add.frame.minY, headerTop, accuracy: 2)
    capture("home-header-after-scroll")
  }

  func testMapRootContextAndAncestorReturn() {
    guard openFixtureURL("audit-browse-journey") else { return }
    let control = app.segmentedControls.firstMatch
    XCTAssertTrue(control.waitForExistence(timeout: 10))
    control.buttons["Map"].tap()
    let root = app.buttons["Open location Main Inventory"].firstMatch
    XCTAssertTrue(app.staticTexts["Main Inventory"].firstMatch.waitForExistence(timeout: 10))
    XCTAssertFalse(root.exists)
    let garage = app.buttons["Garage, Place, 10 inside"].firstMatch
    XCTAssertTrue(garage.waitForExistence(timeout: 5)); XCTAssertTrue(garage.isHittable)
    capture("map-root-context")
    garage.tap()
    let breadcrumb = app.buttons["Open location Garage"].firstMatch
    XCTAssertTrue(breadcrumb.waitForExistence(timeout: 5))
    XCTAssertTrue(root.isHittable)
    capture("map-garage-context")
    root.tap()
    XCTAssertTrue(breadcrumb.waitForNonExistence(timeout: 5))
    XCTAssertFalse(root.exists)
    XCTAssertTrue(garage.isHittable)
    XCTAssertTrue(app.staticTexts["Main Inventory"].firstMatch.exists)
    capture("map-root-return")
  }

  func testEmptyPhotoDetailPrioritizesIdentityAndActions() {
    guard openFixtureURL("audit-edit-journey") else { return }
    let title = app.staticTexts["Camping tent"].firstMatch
    let status = app.staticTexts["No photos"].firstMatch
    XCTAssertTrue(title.waitForExistence(timeout: 10))
    XCTAssertTrue(status.waitForExistence(timeout: 5))
    let edit = app.navigationBars["Details"].buttons["Edit"].firstMatch

    let move = app.buttons["Move"].firstMatch
    let add = app.buttons["Add photos"].firstMatch
    XCTAssertTrue(edit.isHittable); XCTAssertTrue(move.isHittable)
    XCTAssertTrue(add.isHittable)
    XCTAssertGreaterThan(title.frame.height, 0)
    XCTAssertGreaterThan(status.frame.height, 0)
    XCTAssertLessThanOrEqual(title.frame.maxY, status.frame.minY)
    XCTAssertLessThanOrEqual(edit.frame.maxY, status.frame.minY)
    XCTAssertLessThanOrEqual(move.frame.maxY, status.frame.minY)
    XCTAssertLessThanOrEqual(add.frame.maxY, app.frame.maxY)
    capture("detail-empty-photo-hierarchy")
  }

  func testBrowsePhotoFreeRowsKeepMixedMediaAligned() {
    guard openFixtureURL("audit-browse-journey") else { return }
    let garage = app.otherElements["asset-card-journey-0"].firstMatch
    XCTAssertTrue(garage.waitForExistence(timeout: 10))
    XCTAssertLessThan(garage.frame.height, garage.frame.width,
      "Confirmed photo-free grid cards must not reserve a square media panel")
    capture("browse-photo-free-compact")
    guard openFixtureURL("audit-browse-journey?photoMix=true") else { return }
    let mixed = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      garage.exists && garage.frame.height > garage.frame.width
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [mixed], timeout: 10), .completed)
    let garageTitle = app.buttons["Open asset Garage"].firstMatch
    let kitchenTitle = app.buttons["Open asset Kitchen"].firstMatch
    XCTAssertTrue(kitchenTitle.waitForExistence(timeout: 10))
    XCTAssertEqual(garageTitle.frame.minY, kitchenTitle.frame.minY, accuracy: 1,
      "Photo-free peers must align their titles with the photo card")
    capture("browse-mixed-photo-alignment")
  }

  func testBrowseGridFitsDeviceWidth() {
    guard openFixtureURL("audit-browse-journey") else { return }
    let garage = app.otherElements["asset-card-journey-0"].firstMatch
    let kitchen = app.otherElements["asset-card-journey-1"].firstMatch
    let tent = app.otherElements["asset-card-journey-2"].firstMatch
    XCTAssertTrue(garage.waitForExistence(timeout: 10))
    XCTAssertTrue(kitchen.waitForExistence(timeout: 10))
    XCTAssertTrue(tent.waitForExistence(timeout: 10))
    XCTAssertEqual(garage.frame.minY, kitchen.frame.minY, accuracy: 1)
    XCTAssertEqual(garage.frame.width, kitchen.frame.width, accuracy: 1)
    XCTAssertGreaterThan(kitchen.frame.minX, garage.frame.maxX)
    if UIDevice.current.userInterfaceIdiom == .pad {
      XCTAssertEqual(garage.frame.minY, tent.frame.minY, accuracy: 1)
      XCTAssertEqual(garage.frame.width, tent.frame.width, accuracy: 1)
      XCTAssertGreaterThanOrEqual(garage.frame.width, 220)
      XCTAssertLessThan(garage.frame.width, 300)
      XCTAssertGreaterThan(tent.frame.minX, kitchen.frame.maxX)
    } else {
      XCTAssertGreaterThan(tent.frame.minY, garage.frame.minY)
    }
    XCTAssertLessThanOrEqual(kitchen.frame.maxX, app.frame.maxX)
    XCTAssertLessThanOrEqual(tent.frame.maxX, app.frame.maxX)
    capture("browse-adaptive-grid")
  }

  func testBrowseViewSwitcherStaysAnchoredAcrossListMapAndScroll() {
    guard openFixtureURL("audit-browse-journey?dense=true") else { return }
    let control = app.segmentedControls.firstMatch
    XCTAssertTrue(control.waitForExistence(timeout: 10))
    let list = control.buttons["List"]
    let map = control.buttons["Map"]
    XCTAssertTrue(list.isSelected)
    let initialFrame = control.frame
    func verifyAnchor() {
      XCTAssertTrue(list.isHittable)
      XCTAssertTrue(map.isHittable)
      XCTAssertEqual(app.segmentedControls.count, 1)
      XCTAssertEqual(control.frame.minX, initialFrame.minX, accuracy: 1)
      XCTAssertEqual(control.frame.minY, initialFrame.minY, accuracy: 1)
      XCTAssertEqual(control.frame.width, initialFrame.width, accuracy: 1)
      XCTAssertTrue(app.buttons["Add an asset"].firstMatch.isHittable)
      XCTAssertTrue(app.buttons["Search"].firstMatch.isHittable)
    }
    let listItem = app.buttons["Open asset Camping tent"].firstMatch
    XCTAssertTrue(listItem.waitForExistence(timeout: 10))
    let listItemY = listItem.frame.minY
    verifyAnchor()
    capture("browse-journey-list-top")
    // Start in the card gutter so this scroll cannot activate a card command.
    let firstCard = app.otherElements["asset-card-journey-0"].firstMatch
    let secondCard = app.otherElements["asset-card-journey-1"].firstMatch
    let gutterX = (firstCard.frame.maxX + secondCard.frame.minX) / 2
    let origin = app.coordinate(withNormalizedOffset: .zero)
    let start = origin.withOffset(CGVector(dx: gutterX, dy: app.frame.height * 0.7))
    let end = origin.withOffset(CGVector(dx: gutterX, dy: app.frame.height * 0.3))
    start.press(forDuration: 0.05, thenDragTo: end)
    XCTAssertTrue(control.exists, "Scrolling must remain on Browse")
    XCTAssertTrue(!listItem.exists || listItem.frame.minY < listItemY - 40, "The list must actually scroll")
    verifyAnchor()
    capture("browse-journey-list-scrolled")
    map.tap()
    XCTAssertTrue(map.isSelected)
    let overview = app.staticTexts["36 active assets · 2 root items"].firstMatch
    XCTAssertTrue(overview.waitForExistence(timeout: 10))
    XCTAssertFalse(app.buttons["Filters"].exists)
    verifyAnchor()
    capture("browse-journey-map")
    list.tap()
    XCTAssertTrue(list.isSelected)
    XCTAssertTrue(app.buttons["Filters"].firstMatch.waitForExistence(timeout: 10))
    XCTAssertFalse(overview.exists)
    verifyAnchor()
    capture("browse-journey-list-return")
  }

  private func tabCandidates(_ name: String) -> XCUIElementQuery {
    if UIDevice.current.userInterfaceIdiom == .pad {
      // The fixture gives these labels only to native tabs. Ancestor indices can
      // become stale as UIKit replaces its iPad tab/navigation containers.
      return app.buttons.matching(identifier: name)
    }
    return app.tabBars.firstMatch.buttons.matching(identifier: name)
  }

  private func tab(_ name: String) -> XCUIElement {
    let candidates = tabCandidates(name)
    return candidates.allElementsBoundByIndex.first(where: { $0.isHittable }) ?? candidates.firstMatch
  }


  private func activateVisibleTabForTouchObservation(_ name: String) {
    if UIDevice.current.userInterfaceIdiom != .pad {
      tab(name).tap()
      return
    }
    let bounds = app.frame
    let candidate = tabCandidates(name).allElementsBoundByIndex.first { element in
      let frame = element.frame
      return !frame.isEmpty && !frame.isInfinite && !frame.isNull && bounds.contains(frame)
    }
    guard let candidate else {
      XCTFail("No on-screen native tab bounds for \(name)")
      return
    }
    candidate.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5)).tap()
  }


  func testSettingsOverviewNavigationAndFooterClearance() {
    guard openFixtureURL("audit-tabs/(home)/settings") else { return }
    let account = app.buttons["Open Account settings for household.member@example.invalid"]
    XCTAssertTrue(account.waitForExistence(timeout: 10))
    XCTAssertTrue(account.isHittable)
    // RN groups this row into one AX button; review visible label/subtitle in
    // the capture rather than asserting child text that AX does not expose.
    capture("settings-overview-root")
    account.tap()
    XCTAssertTrue(app.buttons["Sign Out"].waitForExistence(timeout: 10))
    XCTAssertTrue(app.staticTexts["household.member@example.invalid"].waitForExistence(timeout: 10))
    capture("settings-overview-account")
    app.navigationBars.buttons.firstMatch.tap()
    let inventory = app.buttons["Open inventory settings for Main Inventory, in Family household"]
    XCTAssertTrue(inventory.waitForExistence(timeout: 10))
    XCTAssertTrue(inventory.isHittable); inventory.tap()
    XCTAssertTrue(app.buttons["Open Notifications for Main Inventory"].waitForExistence(timeout: 10))
    capture("settings-overview-inventory")
    app.navigationBars.buttons.firstMatch.tap()
    let diagnostics = app.buttons["Open developer and connection Diagnostics"]
    XCTAssertTrue(diagnostics.waitForExistence(timeout: 10))
    for _ in 0..<5 where !diagnostics.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(diagnostics.isHittable); diagnostics.tap()
    // Selectable text exposes parent and child AX labels with identical frames.
    let version = app.staticTexts["overview-audit-1"].firstMatch
    XCTAssertTrue(version.waitForExistence(timeout: 10))
    verifyFooterClearsPersistentChrome(version)
    capture("settings-overview-diagnostics-clearance")
    tab("Browse").tap()
    XCTAssertTrue(app.buttons["Open Browse asset"].waitForExistence(timeout: 10))
    tab("Home").tap()
    XCTAssertTrue(version.waitForExistence(timeout: 10))
    capture("settings-overview-tab-return")
  }

  func testNotificationJourneyRetainsTabsAndClearsFooter() {
    guard openFixtureURL("(tabs)/(home)/notifications") else { return }
    let first = app.buttons["Open Camping item 01"].firstMatch
    XCTAssertTrue(first.waitForExistence(timeout: 10))
    XCTAssertTrue(tab("Home").isHittable); XCTAssertTrue(tab("Browse").isHittable)
    capture("notifications-tab-entry")
    first.tap()
    XCTAssertTrue(app.navigationBars["Details"].waitForExistence(timeout: 10))
    XCTAssertTrue(app.staticTexts["Camping item 01"].firstMatch.waitForExistence(timeout: 10))
    XCTAssertTrue(tab("Home").isHittable); XCTAssertTrue(tab("Browse").isHittable)
    capture("notification-asset-detail")
    app.navigationBars["Details"].buttons.firstMatch.tap()
    XCTAssertTrue(app.buttons["Mark Camping item 01 unread"].firstMatch.waitForExistence(timeout: 10))
    let more = app.buttons["Load more notifications"].firstMatch
    verifyFooterClearsPersistentChrome(more)
    capture("notification-pagination")
    more.tap()
    let finalContent = app.buttons["Open Camping item 24 with a long descriptive name"].firstMatch
    verifyFooterClearsPersistentChrome(finalContent)
    let final = app.buttons["Mark Camping item 24 with a long descriptive name read"].firstMatch
    verifyFooterClearsPersistentChrome(final)
    XCTAssertTrue(final.isHittable)
    capture("notifications-final-row")
  }

  func testCustomizationCollectionClearsPersistentChrome() {
    guard openFixtureURL("(tabs)/(home)/settings/inventory/tags") else { return }
    let first = app.buttons["Tag 01, No color"].firstMatch
    XCTAssertTrue(first.waitForExistence(timeout: 10))
    let header = app.navigationBars["Tags"]
    XCTAssertTrue(header.exists)
    let clearsHeader = first.frame.minY >= header.frame.maxY
    capture("customization-collection-entry")
    XCTAssertTrue(app.buttons["Add Tag"].firstMatch.isHittable)
    app.buttons["Search"].firstMatch.tap()
    let field = app.searchFields.firstMatch
    XCTAssertTrue(field.waitForExistence(timeout: 5))
    waitForKeyboard()
    field.typeText("Tag 01")
    XCTAssertEqual(field.value as? String, "Tag 01")
    XCTAssertTrue(app.buttons["Tag 02, No color"].waitForNonExistence(timeout: 5))
    XCTAssertTrue(first.waitForExistence(timeout: 5))
    XCTAssertTrue(first.isHittable, "Search retains its matching row")
    capture("customization-collection-search")
    resetNativeSearch(field)
    XCTAssertTrue(app.buttons["Tag 02, No color"].waitForExistence(timeout: 5))
    let final = app.buttons["Tag 40 with a long descriptive household storage name, No color"].firstMatch
    verifyFooterClearsPersistentChrome(final)
    capture("customization-collection-footer")
    XCTAssertTrue(final.isHittable)
    XCTAssertTrue(clearsHeader, "Initial collection row clears navigation chrome")
  }

  func testInventoryCollectionClearsPersistentChrome() {
    guard openFixtureURL("(tabs)/(home)/assets") else { return }
    XCTAssertTrue(app.buttons["Open asset Inventory item 1"].firstMatch.waitForExistence(timeout: 10))
    let heading = app.staticTexts["Recently changed"].firstMatch
    XCTAssertTrue(heading.exists)
    let header = app.navigationBars["Assets"]
    XCTAssertTrue(header.exists)
    let headingClearsHeader = heading.frame.minY >= header.frame.maxY
    capture("inventory-collection-entry")
    let footer = app.buttons["Search for tag Final inventory tag"].firstMatch
    verifyFooterClearsPersistentChrome(footer)
    capture("inventory-collection-footer")
    XCTAssertTrue(footer.isHittable)
    XCTAssertTrue(headingClearsHeader, "Inventory heading must clear the navigation header at entry")
  }

  func testVoiceAccessorySettledNavigationAppearance() {
    guard openFixtureURL("(tabs)/(home)/assets/history-item/history") else { return }
    let change = app.buttons.matching(NSPredicate(format: "label CONTAINS 'Name · Description · Location · Tags'")).firstMatch
    XCTAssertTrue(change.waitForExistence(timeout: 10))
    func captureSettledAccessory(_ label: String) {
      let command = app.buttons["Start voice interaction"].firstMatch
      XCTAssertTrue(command.waitForExistence(timeout: 10)); XCTAssertTrue(command.isHittable)
      // This is a bounded visual observation, not a longer functional timeout.
      // Its image must be reviewed; the accessible button name cannot prove a drawn glyph.
      RunLoop.current.run(until: Date().addingTimeInterval(2))
      capture(label)
      // Element screenshots can omit a glyph visible in the full-screen capture.
      // Review the full-screen composition above; the discrepancy's cause is unproven.
    }
    captureSettledAccessory("voice-navigation-list")
    XCTAssertTrue(change.isHittable, "History change must be interactive before opening its detail")
    change.tap()
    XCTAssertTrue(app.navigationBars["History detail"].waitForExistence(timeout: 10))
    XCTAssertTrue(app.staticTexts["What changed"].waitForExistence(timeout: 10))
    captureSettledAccessory("voice-navigation-detail-settled")
    tab("Browse").tap()
    XCTAssertTrue(app.segmentedControls.firstMatch.buttons["List"].waitForExistence(timeout: 10))
    tab("Home").tap()
    XCTAssertTrue(app.navigationBars["History detail"].waitForExistence(timeout: 10))
    captureSettledAccessory("voice-navigation-detail-tab-return")
  }

  func testHistoryJourneyClearsPersistentChrome() {
    guard openFixtureURL("(tabs)/(home)/assets/history-item/history") else { return }
    let title = app.staticTexts["Camping equipment"].firstMatch
    XCTAssertTrue(title.waitForExistence(timeout: 10))
    let mode = app.buttons.matching(NSPredicate(format: "label CONTAINS 'Show History'")).firstMatch
    XCTAssertTrue(mode.waitForExistence(timeout: 10))
    XCTAssertTrue(mode.isHittable); mode.tap()
    let allEvents = app.buttons["All events"].firstMatch
    XCTAssertTrue(allEvents.waitForExistence(timeout: 5)); allEvents.tap()
    XCTAssertTrue(app.buttons.matching(NSPredicate(format: "label CONTAINS 'Show History, All events'")).firstMatch.waitForExistence(timeout: 5))
    XCTAssertTrue(app.navigationBars["History"].exists)
    capture("history-all-events-choice")
    mode.tap()
    let changes = app.buttons["Changes"].firstMatch
    XCTAssertTrue(changes.waitForExistence(timeout: 5)); changes.tap()
    capture("history-list-entry")
    let headingBelowHeader = title.frame.minY >= app.navigationBars.firstMatch.frame.maxY
    let modeBelowHeader = mode.frame.minY >= app.navigationBars.firstMatch.frame.maxY
    let firstChange = app.buttons.matching(NSPredicate(format: "label CONTAINS 'Name · Description · Location · Tags'")).firstMatch
    XCTAssertTrue(firstChange.waitForExistence(timeout: 10))
    XCTAssertTrue(firstChange.isHittable); firstChange.tap()
    let disclosure = app.buttons.matching(NSPredicate(format: "label CONTAINS 'Technical details'")).firstMatch
    XCTAssertTrue(disclosure.waitForExistence(timeout: 10))
    capture("history-detail-entry")
    let scroll = app.scrollViews.firstMatch
    for _ in 0..<6 where !disclosure.isHittable { scroll.swipeUp() }
    XCTAssertTrue(disclosure.isHittable); disclosure.tap()
    let finalMetadata = app.staticTexts["history-request-0"].firstMatch
    XCTAssertTrue(finalMetadata.waitForExistence(timeout: 10))
    capture("history-detail-expanded")
    verifyFooterClearsPersistentChrome(finalMetadata)
    capture("history-detail-footer")
    app.navigationBars.firstMatch.buttons.firstMatch.tap()
    XCTAssertTrue(title.waitForExistence(timeout: 10))
    verifyFooterClearsPersistentChrome(app.buttons["Load older activity"].firstMatch)
    capture("history-pagination-footer")
    tab("Browse").tap()
    XCTAssertTrue(app.segmentedControls.firstMatch.buttons["List"].waitForExistence(timeout: 10))
    tab("Home").tap()
    XCTAssertTrue(title.waitForExistence(timeout: 10))
    capture("history-list-tab-return")
    XCTAssertTrue(headingBelowHeader, "History title must clear the native header")
    XCTAssertTrue(modeBelowHeader, "History mode must clear the native header")
  }

  func testDetailFooterClearsPersistentTabsAndVoiceAccessory() {
    guard openFixtureURL("audit-tabs/(home)/assets/footer-clearance") else { return }
    let footer = app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH 'Updated '")).firstMatch
    verifyFooterClearsPersistentChrome(footer)
    capture("detail-footer-above-native-tabs")
  }

  func testSharingFooterClearsPersistentTabsAndVoiceAccessory() {
    guard openFixtureURL("audit-tabs/(home)/settings/sharing?access=populated") else { return }
    let footer = app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH 'Invitation links are shown only'")).firstMatch
    XCTAssertTrue(footer.waitForExistence(timeout: 10))
    verifyFooterClearsPersistentChrome(footer)
    capture("sharing-footer-above-native-tabs")
  }

  private func verifyFooterClearsPersistentChrome(_ footer: XCUIElement) {
    let voice = app.buttons["Start voice interaction"].firstMatch
    let tabs = app.tabBars.firstMatch
    XCTAssertTrue(voice.waitForExistence(timeout: 10))
    XCTAssertTrue(tab("Home").isHittable); XCTAssertTrue(tab("Browse").isHittable)
    let scroll = app.scrollViews.firstMatch
    XCTAssertTrue(scroll.exists)
    func clearOfChrome() -> Bool {
      let tabBounds = tabs.exists ? tabs.frame : tab("Home").frame.union(tab("Browse").frame)
      // iPad can put its tab strip at the top; use the delivered chrome positions.
      let chrome = [tabBounds, voice.frame]
      let upper = chrome.filter { $0.midY < app.frame.midY }
        .reduce(app.navigationBars.firstMatch.frame.maxY) { max($0, $1.maxY) }
      let lower = chrome.filter { $0.midY >= app.frame.midY }
        .reduce(app.frame.maxY) { min($0, $1.minY) }
      return footer.exists && app.frame.contains(footer.frame) &&
        footer.frame.maxY <= lower && footer.frame.minY >= upper
    }
    for _ in 0..<8 where !clearOfChrome() { scroll.swipeUp() }
    XCTAssertTrue(clearOfChrome(), "Final content must clear persistent navigation surfaces")
  }

  func testHomeCollectionsReplaceBrowseRefinementsAndRetainTabs() {
    guard openFixtureURL("(tabs)/(search)/search?surface=map&query=Kitchen&checkoutState=available") else { return }
    let control = app.segmentedControls.firstMatch
    XCTAssertTrue(control.waitForExistence(timeout: 10))
    XCTAssertTrue(control.buttons["Map"].isSelected)
    tab("Home").tap()
    let recent = app.buttons["View all recently changed assets"].firstMatch
    XCTAssertTrue(recent.waitForExistence(timeout: 10)); recent.tap()
    let first = app.buttons.matching(NSPredicate(format: "label BEGINSWITH %@", "Open asset Camping item 01")).firstMatch
    let second = app.buttons.matching(NSPredicate(format: "label BEGINSWITH %@", "Open asset Camping item 02")).firstMatch
    XCTAssertTrue(first.waitForExistence(timeout: 10))
    XCTAssertTrue(second.exists)
    XCTAssertTrue(tab("Browse").isSelected)
    XCTAssertTrue(control.buttons["List"].isSelected)
    capture("home-recent-browse-list")
    first.tap()
    let title = app.staticTexts["Camping item 01"].firstMatch
    XCTAssertTrue(title.waitForExistence(timeout: 10))
    XCTAssertTrue(tab("Home").isHittable); XCTAssertTrue(tab("Browse").isHittable)
    app.navigationBars.buttons["Back"].firstMatch.tap()
    XCTAssertTrue(first.waitForExistence(timeout: 10))
    guard openFixtureURL("(tabs)/(search)/search?query=Kitchen&scope=containers&tagId=outdoors&lifecycleState=archived&sort=id_asc") else { return }
    XCTAssertTrue(app.staticTexts["No results for “Kitchen”"].firstMatch.waitForExistence(timeout: 10))
    XCTAssertFalse(first.exists)
    tab("Home").tap()
    let checked = app.buttons["View all checked-out assets"].firstMatch
    XCTAssertTrue(checked.waitForExistence(timeout: 10))
    if !checked.isHittable { app.swipeUp() }
    XCTAssertTrue(checked.isHittable); checked.tap()
    XCTAssertTrue(first.waitForExistence(timeout: 10))
    XCTAssertFalse(second.exists)
    XCTAssertTrue(tab("Browse").isSelected)
    capture("home-checked-out-browse-list")
    tab("Home").tap(); XCTAssertTrue(recent.waitForExistence(timeout: 10))
    tab("Browse").tap(); XCTAssertTrue(first.waitForExistence(timeout: 10))
    XCTAssertFalse(second.exists, "Ordinary tab return preserves the checked-out collection")
  }

  func testPersistentTabsRetainDestinationsDraftsAndModalReturn() {
    guard openFixtureURL("audit-tabs/assets/audit-edit-item") else { return }
    let move = app.buttons["Move"].firstMatch
    XCTAssertTrue(move.waitForExistence(timeout: 10))
    XCTAssertTrue(tab("Home").isHittable); XCTAssertTrue(tab("Browse").isHittable)
    capture("persistent-tabs-home-detail")
    move.tap()
    let cancel = app.buttons["Cancel"].firstMatch
    XCTAssertTrue(cancel.waitForExistence(timeout: 10)); cancel.tap()
    XCTAssertTrue(cancel.waitForNonExistence(timeout: 10))
    XCTAssertTrue(move.isHittable); XCTAssertTrue(tab("Browse").isHittable)
    tab("Browse").tap()
    let browse = app.staticTexts["Tab shell Browse placeholder"].firstMatch
    XCTAssertTrue(browse.waitForExistence(timeout: 10))
    let openBrowseAsset = app.buttons["Open Browse asset"].firstMatch
    XCTAssertTrue(openBrowseAsset.isHittable)
    XCTAssertGreaterThanOrEqual(openBrowseAsset.frame.minY, app.navigationBars.firstMatch.frame.maxY)
    openBrowseAsset.tap()
    XCTAssertTrue(move.waitForExistence(timeout: 10))
    tab("Home").tap(); XCTAssertTrue(move.waitForExistence(timeout: 10))
    tab("Browse").tap(); XCTAssertTrue(move.waitForExistence(timeout: 10))
    capture("persistent-tabs-browse-detail-return")
    app.navigationBars.buttons["Back"].firstMatch.tap()
    XCTAssertTrue(browse.waitForExistence(timeout: 10), "Browse must retain its own destination stack")
    guard openFixtureURL("audit-tabs/(home)/settings/inventory/tags/tools") else { return }
    let name = app.textFields["Name"].firstMatch
    XCTAssertTrue(name.waitForExistence(timeout: 10)); name.tap(); name.typeText(" emergency")
    XCTAssertEqual(name.value as? String, "Tools emergency")
    let dismiss = app.buttons["Dismiss keyboard"].firstMatch
    XCTAssertTrue(dismiss.isHittable); dismiss.tap()
    XCTAssertTrue(tab("Browse").isHittable); tab("Browse").tap()
    XCTAssertTrue(browse.waitForExistence(timeout: 10)); tab("Home").tap()
    XCTAssertTrue(name.waitForExistence(timeout: 10)); XCTAssertEqual(name.value as? String, "Tools emergency")
    capture("persistent-tabs-settings-draft-before-touch")
    activateVisibleTabForTouchObservation("Browse")
    XCTAssertTrue(browse.waitForExistence(timeout: 10), "Visible Browse tab must switch destinations")
    XCTAssertTrue(browse.isHittable)
    XCTAssertTrue(tab("Browse").isSelected)
    XCTAssertFalse(name.isHittable)
    capture("persistent-tabs-settings-browse-touch")
    activateVisibleTabForTouchObservation("Home")
    XCTAssertTrue(name.waitForExistence(timeout: 10))
    XCTAssertTrue(name.isHittable)
    XCTAssertTrue(tab("Home").isSelected)
    XCTAssertFalse(browse.isHittable)
    XCTAssertEqual(name.value as? String, "Tools emergency")
    capture("persistent-tabs-settings-draft-return")
    let tabsReturned = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      self.tab("Home").isHittable && self.tab("Browse").isHittable
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [tabsReturned], timeout: 10), .completed,
      "Both tabs must remain accessibility-hittable after returning to the draft editor")
  }

  func testExpirationFiltersReturnToTheirOwningTab() {
    guard openFixtureURL("audit-tabs") else { return }
    let homeExpiration = app.buttons["View all expiration dates"].firstMatch
    XCTAssertTrue(homeExpiration.waitForExistence(timeout: 10)); homeExpiration.tap()
    let kitchen = app.buttons.matching(NSPredicate(format: "label BEGINSWITH %@", "Open asset Kitchen item")).firstMatch
    XCTAssertTrue(kitchen.waitForExistence(timeout: 10))
    XCTAssertTrue(tab("Browse").isHittable); tab("Browse").tap()
    let browseExpiration = app.buttons["Open Browse expiration"].firstMatch
    XCTAssertTrue(browseExpiration.waitForExistence(timeout: 10)); browseExpiration.tap()
    let camping = app.buttons.matching(NSPredicate(format: "label BEGINSWITH %@", "Open asset Camping item 01")).firstMatch
    XCTAssertTrue(camping.waitForExistence(timeout: 10))
    tab("Home").tap()
    XCTAssertTrue(kitchen.waitForExistence(timeout: 10)); XCTAssertFalse(camping.exists,
      "Tab return must retain Home's query before Filters is opened")
    tab("Browse").tap(); XCTAssertTrue(camping.waitForExistence(timeout: 10))
    let filters = app.buttons["Filter expiration items"].firstMatch
    XCTAssertTrue(filters.isHittable); filters.tap()
    let apply = app.buttons["Apply expiration filters"].firstMatch
    XCTAssertTrue(apply.waitForExistence(timeout: 10)); XCTAssertTrue(apply.isHittable); apply.tap()
    XCTAssertTrue(apply.waitForNonExistence(timeout: 10))
    XCTAssertTrue(camping.waitForExistence(timeout: 10))
    XCTAssertTrue(tab("Home").isHittable); tab("Home").tap()
    XCTAssertTrue(kitchen.waitForExistence(timeout: 10)); XCTAssertFalse(camping.exists)
    tab("Browse").tap()
    XCTAssertTrue(camping.waitForExistence(timeout: 10)); XCTAssertFalse(kitchen.exists)
    capture("expiration-filters-browse-origin-retained")
  }

  func testHomeTabShellPreservesActionsAndAccessoryAfterTabReturn() {
    let entry = app.buttons["Audit Home in tabs"].firstMatch
    for _ in 0..<12 where !entry.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(entry.isHittable)
    entry.tap()
    let add = app.buttons["Add an asset"].firstMatch
    let notifications = app.buttons["Notifications, 2 unread"].firstMatch
    let profile = app.buttons["Open account and settings"].firstMatch
    let voice = app.buttons["Start voice interaction"].firstMatch
    XCTAssertTrue(add.waitForExistence(timeout: 10))
    XCTAssertTrue(voice.waitForExistence(timeout: 10))
    let tabs: XCUIElement
    if UIDevice.current.userInterfaceIdiom == .pad {
      // iPad's top strip is an Other containing nested Home/Browse buttons.
      let containers = app.otherElements.containing(.button, identifier: "Home")
        .containing(.button, identifier: "Browse").allElementsBoundByIndex
      guard let strip = containers.filter({ $0.frame.height > 0 })
        .min(by: { $0.frame.height < $1.frame.height }) else {
        XCTFail("Native Home/Browse tab strip is missing")
        return
      }
      tabs = strip
    } else {
      tabs = app.tabBars.firstMatch
    }
    XCTAssertTrue(tabs.exists)
    for label in ["Home", "Browse"] {
      let tab = tabs.buttons[label].firstMatch
      XCTAssertTrue(tab.isHittable)
      XCTAssertTrue(app.frame.contains(tab.frame))
      XCTAssertTrue(tabs.frame.contains(tab.frame))
    }
    func verifyActions() {
      for action in [add, notifications, profile, voice] {
        XCTAssertTrue(action.isHittable)
        XCTAssertTrue(app.frame.contains(action.frame))
      }
      XCTAssertLessThanOrEqual(add.frame.maxX, notifications.frame.minX)
      XCTAssertLessThanOrEqual(notifications.frame.maxX, profile.frame.minX)
    }
    verifyActions()
    capture("home-tab-shell-entry")
    let headerTop = add.frame.minY
    let lastReturn = app.buttons["Return Audit garden tools"].firstMatch
    let scroll = app.scrollViews.containing(.staticText, identifier: "Recently changed").firstMatch
    func clearOfChrome() -> Bool {
      lastReturn.exists && lastReturn.isHittable && app.frame.contains(lastReturn.frame) &&
        !lastReturn.frame.intersects(voice.frame) && !lastReturn.frame.intersects(tabs.frame)
    }
    for _ in 0..<8 where !clearOfChrome() { scroll.swipeUp() }
    XCTAssertTrue(clearOfChrome(), "Last Home action must remain reachable clear of native tabs and voice entry")
    verifyActions()
    XCTAssertEqual(add.frame.minY, headerTop, accuracy: 2)
    capture("home-tab-shell-scrolled")
    let browseTab = tabs.buttons["Browse"].firstMatch
    XCTAssertTrue(browseTab.isHittable)
    browseTab.tap()
    XCTAssertTrue(app.staticTexts["Tab shell Browse placeholder"].waitForExistence(timeout: 5))
    let homeTab = tabs.buttons["Home"].firstMatch
    XCTAssertTrue(homeTab.isHittable)
    homeTab.tap()
    XCTAssertTrue(add.waitForExistence(timeout: 10))
    XCTAssertTrue(app.activityIndicators.firstMatch.waitForNonExistence(timeout: 5), "Tab return must not leave a pull indicator active")
    verifyActions()
    capture("home-tab-shell-return")
  }

  func testHomeNotificationHitRegionBeyondAccessibilityFrame() {
    let open = app.buttons["Audit Home header"]
    XCTAssertTrue(open.waitForExistence(timeout: 5))
    let menu = app.scrollViews.containing(.button, identifier: "Audit Home header").firstMatch
    for _ in 0..<6 where !open.isHittable { menu.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let action = app.buttons["Notifications, 2 unread"]
    XCTAssertTrue(action.waitForExistence(timeout: 10))
    XCTAssertTrue(action.isHittable)
    // Probe inside each edge of a centered 44-point square, independently of
    // the AX frame. Each tap must reach this action exactly once.
    let offsets: [(CGFloat, CGFloat)] = [(0, 0), (0, -21), (0, 21), (-21, 0), (21, 0),
                                          (-21, -21), (21, -21), (-21, 21), (21, 21)]
    for (index, offset) in offsets.enumerated() {
      let center = action.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
      center.withOffset(CGVector(dx: offset.0, dy: offset.1)).tap()
      XCTAssertTrue(app.staticTexts["Header notification activations: \(index + 1)"].waitForExistence(timeout: 5),
                    "Notification did not receive edge probe \(index): \(offset)")
    }
    capture("home-notification-hit-region")
  }

  func testHomeAddHitRegionBeyondAccessibilityFrame() {
    verifyHomeNavigationHitRegion(label: "Add an asset", destination: "Header Add destination", captureName: "home-add-hit-region")
  }

  func testHomeProfileHitRegionBeyondAccessibilityFrame() {
    verifyHomeNavigationHitRegion(label: "Open account and settings", destination: "Header Profile destination", captureName: "home-profile-hit-region")
  }

  private func verifyHomeNavigationHitRegion(label: String, destination: String, captureName: String) {
    let open = app.buttons["Audit Home header"]
    XCTAssertTrue(open.waitForExistence(timeout: 5))
    let menu = app.scrollViews.containing(.button, identifier: "Audit Home header").firstMatch
    for _ in 0..<6 where !open.isHittable { menu.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let action = app.buttons[label]
    let offsets: [(CGFloat, CGFloat)] = [(0, 0), (0, -21), (0, 21), (-21, 0), (21, 0),
                                          (-21, -21), (21, -21), (-21, 21), (21, 21)]
    for (index, offset) in offsets.enumerated() {
      XCTAssertTrue(action.waitForExistence(timeout: 10))
      XCTAssertTrue(action.isHittable)
      let center = action.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
      center.withOffset(CGVector(dx: offset.0, dy: offset.1)).tap()
      XCTAssertTrue(app.staticTexts[destination].waitForExistence(timeout: 5),
                    "\(label) did not reach its destination for edge probe \(index): \(offset)")
      let back = app.navigationBars.buttons["BackButton"].firstMatch
      XCTAssertTrue(back.isHittable)
      back.tap()
    }
    XCTAssertTrue(action.waitForExistence(timeout: 10))
    capture(captureName)
  }

  private func openHomeReturn() {
    let open = app.buttons["Audit Home Return"]
    XCTAssertTrue(open.waitForExistence(timeout: 5))
    open.tap()
    let action = app.buttons["Return Audit drill"]
    XCTAssertTrue(action.waitForExistence(timeout: 10))
    for _ in 0..<3 where !action.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(action.isHittable)
    action.tap()
    XCTAssertTrue(app.navigationBars["Return details"].waitForExistence(timeout: 5))
  }

  func testHomeReturnCancelRestoresCheckout() {
    openHomeReturn()
    let cancel = app.buttons["Cancel return"]
    XCTAssertTrue(cancel.waitForExistence(timeout: 5))
    XCTAssertTrue(cancel.isHittable)
    capture("home-return-native-sheet")
    cancel.tap()
    XCTAssertTrue(app.buttons["Return Audit drill"].waitForExistence(timeout: 5))
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: app.navigationBars["Return details"])], timeout: 5), .completed)
    capture("home-return-undo-restored")
  }

  func testHomeReturnDetailsRecoverInsideSheet() {
    openHomeReturn()
    let details = app.textViews["Optional return details"]
    XCTAssertTrue(details.waitForExistence(timeout: 5))
    details.tap()
    waitForKeyboard()
    details.typeText("Returned clean")
    let completeText = XCTNSPredicateExpectation(predicate: NSPredicate(format: "value == %@", "Returned clean"), object: details)
    XCTAssertEqual(XCTWaiter.wait(for: [completeText], timeout: 5), .completed)
    XCTAssertEqual(details.value as? String, "Returned clean")
    let dismiss = app.buttons["Dismiss keyboard"]
    XCTAssertTrue(dismiss.isHittable)
    dismiss.tap()
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: app.keyboards.firstMatch)], timeout: 5), .completed)
    let save = app.buttons["Save"]
    for _ in 0..<3 where !save.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(save.isHittable)
    save.tap()
    XCTAssertTrue(app.staticTexts["Return details error"].waitForExistence(timeout: 5))
    XCTAssertEqual(details.value as? String, "Returned clean")
    // UIKit exposes this nested alert text as both a container and its child.
    // Select the containing text without weakening the complete-frame check.
    let error = app.staticTexts.matching(identifier: "Return details error").firstMatch
    let errorVisible = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      let rect = error.frame
      let header = self.app.navigationBars["Return details"].frame
      return error.exists && rect.height > 0 && rect.minY >= header.maxY &&
        rect.maxY <= self.app.frame.maxY && rect.minX >= self.app.frame.minX && rect.maxX <= self.app.frame.maxX
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [errorVisible], timeout: 5), .completed,
      "The complete failure message must be visible below the native header")
    capture("home-return-save-error-retained")
    XCTAssertTrue(save.isHittable)
    save.tap()
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: app.navigationBars["Return details"])], timeout: 5), .completed)
    XCTAssertEqual(app.state, .runningForeground, "Completing return details must not exit or crash the app")
    XCTAssertTrue(app.staticTexts["Recently changed"].waitForExistence(timeout: 5),
      "Saving must return to Home, not merely remove the details screen")
    capture("home-return-save-complete")
  }

  private func openDraftPhotos() {
    let open = app.buttons["Audit draft photos"]
    for _ in 0..<6 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    XCTAssertTrue(app.buttons["Add photos"].waitForExistence(timeout: 5))
  }

  func testDraftPhotosRemoveTheChosenAttachmentAndRetainReadOnlyPreviews() {
    openDraftPhotos()
    let rail = app.otherElements["voice-plan-photo-previews"].scrollViews.firstMatch
    XCTAssertTrue(rail.exists)
    rail.swipeLeft()
    let last = app.buttons["Remove photo 4"]
    XCTAssertTrue(last.isHittable)
    XCTAssertGreaterThanOrEqual(last.frame.height, 44)
    last.tap()
    XCTAssertTrue(app.staticTexts["Removed photo: photo-4"].waitForExistence(timeout: 5))
    rail.swipeRight()
    app.buttons["Remove photo 2"].tap()
    XCTAssertTrue(app.staticTexts["Removed photo: photo-2"].waitForExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["Photos remaining: 2"].exists)
    app.buttons["Add photos"].tap()
    XCTAssertTrue(app.staticTexts["Photo add requests: 1"].waitForExistence(timeout: 5))
    app.buttons["Make photos read only"].tap()
    let readOnly = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      !self.app.buttons["Add photos"].exists && !self.app.buttons["Remove photo 1"].exists
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [readOnly], timeout: 5), .completed)
    XCTAssertFalse(app.buttons["Add photos"].exists)
    XCTAssertFalse(app.buttons["Remove photo 1"].exists)
    XCTAssertTrue(rail.exists)
    XCTAssertEqual(rail.images.count, 2)
    XCTAssertTrue(app.staticTexts["Photos remaining: 2"].exists)
    capture("draft-photo-native-commands")
  }

  func testDraftPhotoControlsAccessibility() throws {
    openDraftPhotos()
    capture("draft-photo-accessibility-before-audit")
    if #available(iOS 17.0, *) {
      try auditAccessibility([.hitRegion, .sufficientElementDescription, .trait, .contrast, .dynamicType, .textClipped])
    } else {
      throw XCTSkip("XCTest accessibility audit requires iOS 17")
    }
  }

  private func openSettingsControls(scrollEnabled: Bool = true) {
    let button = app.buttons[scrollEnabled ? "Audit settings controls" : "Audit settings controls without scrolling"]
    for _ in 0..<4 where !button.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(button.isHittable)
    button.tap()
    XCTAssertTrue(app.buttons["Back to audit menu"].waitForExistence(timeout: 5))
  }

  func testAppearanceUsesMenuWithoutNavigation() {
    openSettingsControls()
    let choice = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose appearance")).firstMatch
    XCTAssertTrue(choice.waitForExistence(timeout: 5))
    choice.tap()
    app.buttons["Dark"].tap()
    XCTAssertTrue(app.staticTexts["Appearance value: dark"].waitForExistence(timeout: 5))
    XCTAssertTrue(app.buttons["Back to audit menu"].isHittable)
    capture("appearance-in-place-dark")
  }

  func testSettingsCommandsRecoverReminderDraft() {
    guard openFixtureURL("audit-settings-commands") else { return }
    let choice = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose reminder mode")).firstMatch
    XCTAssertTrue(choice.waitForExistence(timeout: 5)); choice.tap()
    app.buttons["Off"].tap()
    let retry = app.buttons["Retry saving reminders"].firstMatch
    let discard = app.buttons["Discard reminder changes"].firstMatch
    XCTAssertTrue(retry.waitForExistence(timeout: 5))
    XCTAssertTrue(retry.isHittable); XCTAssertTrue(discard.isHittable)
    XCTAssertFalse(retry.frame.intersects(discard.frame))
    XCTAssertTrue(app.staticTexts["Saved reminders: defaults"].exists)
    capture("settings-reminder-recovery-pair")
    retry.tap()
    XCTAssertTrue(app.staticTexts["Saved reminders: off"].waitForExistence(timeout: 5))
    app.buttons["Fail next save"].tap(); choice.tap(); app.buttons["Custom"].tap()
    XCTAssertTrue(discard.waitForExistence(timeout: 5)); discard.tap()
    XCTAssertTrue(discard.waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["Saved reminders: off"].exists)
    XCTAssertFalse(app.buttons["Before expiration"].exists)
    let device = app.buttons["Open device settings"].firstMatch
    XCTAssertTrue(device.isHittable); device.tap()
    XCTAssertTrue(app.staticTexts["Device settings activations: 1"].waitForExistence(timeout: 5))
    capture("settings-command-long-label")
  }

  func testReminderModeUsesMenuWithoutNavigation() {
    openSettingsControls()
    let choice = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose reminder mode")).firstMatch
    for _ in 0..<4 where !choice.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(choice.isHittable)
    choice.tap()
    app.buttons["Custom"].tap()
    XCTAssertTrue(app.staticTexts["Reminder mode: custom"].waitForExistence(timeout: 5))
    XCTAssertTrue(app.buttons["Before expiration"].exists)
    choice.tap()
    app.buttons["Use defaults"].tap()
    XCTAssertTrue(app.staticTexts["Reminder mode: defaults"].waitForExistence(timeout: 5))
    XCTAssertFalse(app.buttons["Before expiration"].exists)
    capture("reminder-mode-in-place")
  }

  func testColorPickerOpensDirectlyAndClearPreservesParentDraft() {
    openSettingsControls()
    XCTAssertTrue(app.staticTexts["Color value: none"].exists)
    XCTAssertFalse(app.buttons["Choose a custom tag color"].exists)
    let picker = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose any color")).firstMatch
    XCTAssertTrue(picker.waitForExistence(timeout: 5))
    let target = XCTAttachment(string: "Color target before ordinary tap: \(picker.frame); hittable=\(picker.isHittable)")
    target.name = "color-ordinary-tap-target"
    target.lifetime = .keepAlways
    add(target)
    capture("color-before-ordinary-tap")
    picker.tap()
    let sliders = app.buttons["Sliders"]
    let openedDirectly = sliders.waitForExistence(timeout: 5)
    if !openedDirectly {
      capture("color-not-open-after-five-seconds")
      let appearedLater = sliders.waitForExistence(timeout: 15)
      let timing = XCTAttachment(string: "Opened within five seconds: false; appeared during further observation: \(appearedLater)")
      timing.name = "color-late-presentation"
      timing.lifetime = .keepAlways
      add(timing)
      capture("color-after-late-presentation-observation")
    }
    XCTAssertTrue(openedDirectly, "The system color picker should open directly")
    capture("native-color-picker")
    if UIDevice.current.userInterfaceIdiom == .pad {
      // The retained iPad hierarchy exposes the system popover dismiss region;
      // its Close element exists but is not a visible, hittable button.
      let dismiss = app.otherElements["PopoverDismissRegion"]
      XCTAssertTrue(dismiss.exists)
      dismiss.coordinate(withNormalizedOffset: CGVector(dx: 0.1, dy: 0.9)).tap()
    } else {
      app.buttons["close"].tap()
    }
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: sliders)], timeout: 5), .completed)
    XCTAssertTrue(app.staticTexts["Color value: none"].exists, "Opening and closing must not invent a color")
    app.buttons["Choose Green tag color"].tap()
    XCTAssertTrue(app.staticTexts["Color value: #2E7D32"].exists)
    app.buttons["No tag color"].tap()
    XCTAssertTrue(app.staticTexts["Color value: none"].exists)
  }

  func testColorFirstTapWithPreTapCapture() {
    assertColorFirstTap(captureBeforeTap: true)
  }

  func testColorFirstTapWithoutPreTapCapture() {
    assertColorFirstTap(captureBeforeTap: false)
  }

  func testColorFirstTapWithoutScrolling() {
    assertColorFirstTap(captureBeforeTap: false, scrollEnabled: false)
  }

  private func assertColorFirstTap(captureBeforeTap: Bool, scrollEnabled: Bool = true) {
    // Each test has an independent setUp launch; never retry a missed first tap.
    openSettingsControls(scrollEnabled: scrollEnabled)
    XCTAssertTrue(app.staticTexts["Color value: none"].exists)
    XCTAssertFalse(app.buttons["Choose a custom tag color"].exists)
    let picker = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose any color")).firstMatch
    XCTAssertTrue(picker.waitForExistence(timeout: 5))
    XCTAssertTrue(picker.isEnabled)
    XCTAssertTrue(picker.isHittable)
    let before = picker.frame
    if captureBeforeTap { capture("color-comparison-before-tap") }
    picker.tap()
    let opened = app.buttons["Sliders"].waitForExistence(timeout: 5)
    let evidence = XCTAttachment(string: "Scrolling: \(scrollEnabled); pre-tap capture: \(captureBeforeTap); target before tap: \(before); opened after one tap: \(opened)")
    evidence.name = "color-first-tap-comparison"
    evidence.lifetime = .keepAlways
    add(evidence)
    capture("color-comparison-after-first-tap")
    XCTAssertTrue(opened, "One ordinary tap must present the system picker")
  }

  func testNativeColorLockDisablesTheWellAndPreservesDraft() {
    openSettingsControls()
    app.buttons["Choose Green tag color"].tap()
    XCTAssertTrue(app.staticTexts["Color value: #2E7D32"].waitForExistence(timeout: 5))
    app.buttons["Lock color editing"].tap()
    let picker = app.buttons["Choose any color"].firstMatch
    XCTAssertTrue(picker.waitForExistence(timeout: 5))
    let locked = XCTNSPredicateExpectation(predicate: NSPredicate(format: "enabled == false"), object: picker)
    XCTAssertEqual(XCTWaiter.wait(for: [locked], timeout: 5), .completed)
    capture("native-color-locked")
    picker.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5)).tap()
    let opening = XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == true"), object: app.buttons["Sliders"])
    opening.isInverted = true
    XCTAssertEqual(XCTWaiter.wait(for: [opening], timeout: 2), .completed, "A disabled native well must not present the picker")
    XCTAssertTrue(app.staticTexts["Color value: #2E7D32"].exists)
    app.buttons["Unlock color editing"].tap()
    let unlocked = XCTNSPredicateExpectation(predicate: NSPredicate(format: "enabled == true"), object: picker)
    XCTAssertEqual(XCTWaiter.wait(for: [unlocked], timeout: 5), .completed)
    XCTAssertTrue(picker.isHittable)
    picker.tap()
    XCTAssertTrue(app.buttons["Sliders"].waitForExistence(timeout: 5))
    capture("native-color-unlocked")
  }

  func testNativeColorRedEditsPreserveOtherChannelsInDraft() {
    openSettingsControls()
    app.buttons["Choose Green tag color"].tap()
    XCTAssertTrue(app.staticTexts["Color value: #2E7D32"].waitForExistence(timeout: 5))
    var previous = "#2E7D32"
    for (index, position) in [0.75, 0.25, 0.85].enumerated() {
      let picker = app.buttons["Choose any color"].firstMatch
      XCTAssertTrue(picker.waitForExistence(timeout: 5))
      XCTAssertTrue(picker.isHittable)
      picker.tap()
      let sliders = app.buttons["Sliders"]
      XCTAssertTrue(sliders.waitForExistence(timeout: 5), "The native picker must open before RGB editing")
      sliders.tap()
      capture("color-rgb-before-edit-\(index)")
      let red = app.sliders["Red"].firstMatch
      XCTAssertTrue(red.waitForExistence(timeout: 5), "Require the labeled Red slider, not an assumed control order")
      XCTAssertTrue(red.isHittable)
      red.adjust(toNormalizedSliderPosition: CGFloat(position))
      capture("color-rgb-after-edit-\(index)")
      if UIDevice.current.userInterfaceIdiom == .pad {
        let dismiss = app.otherElements["PopoverDismissRegion"]
        XCTAssertTrue(dismiss.exists)
        dismiss.coordinate(withNormalizedOffset: CGVector(dx: 0.1, dy: 0.9)).tap()
      } else {
        app.buttons["close"].tap()
      }
      XCTAssertTrue(sliders.waitForNonExistence(timeout: 5))
      let parent = app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH %@", "Color value: ")).firstMatch
      XCTAssertTrue(parent.waitForExistence(timeout: 5))
      let changed = XCTNSPredicateExpectation(
        predicate: NSPredicate(format: "label != %@", "Color value: \(previous)"), object: parent)
      XCTAssertEqual(XCTWaiter.wait(for: [changed], timeout: 5), .completed)
      let value = String(parent.label.dropFirst("Color value: ".count))
      XCTAssertNotNil(value.range(of: "^#[0-9A-F]{6}$", options: .regularExpression))
      XCTAssertEqual(String(value.suffix(4)), "7D32", "Editing Red must preserve the original Green and Blue bytes")
      XCTAssertNotEqual(value, previous, "The Red edit must reach the parent draft")
      previous = value
      capture("color-rgb-parent-draft-\(index)")
    }
  }

  func testSettingsSaveReadsBackFromTheSameCollection() {
    guard openFixtureURL("audit-settings-readback") else { return }
    let original = app.buttons["Tools, No color"].firstMatch
    XCTAssertTrue(original.waitForExistence(timeout: 10)); XCTAssertTrue(original.isHittable)
    XCTAssertGreaterThanOrEqual(original.frame.minY, app.navigationBars["Tags"].frame.maxY)
    XCTAssertLessThanOrEqual(original.frame.minY, app.navigationBars["Tags"].frame.maxY + 48)
    capture("settings-connected-collection-entry")
    original.tap()
    let name = app.textFields["Name"].firstMatch
    XCTAssertTrue(name.waitForExistence(timeout: 10)); XCTAssertEqual(name.value as? String, "Tools")
    name.tap(); name.typeText(" emergency supplies")
    XCTAssertEqual(name.value as? String, "Tools emergency supplies")
    app.buttons["Dismiss keyboard"].firstMatch.tap()
    let save = app.buttons["Save"].firstMatch
    XCTAssertTrue(save.isHittable); save.tap()
    XCTAssertTrue(app.staticTexts["Could not save"].firstMatch.waitForExistence(timeout: 10))
    XCTAssertEqual(name.value as? String, "Tools emergency supplies")
    capture("settings-connected-save-rejected")
    XCTAssertTrue(save.isEnabled); save.tap()
    let updated = app.buttons["Tools emergency supplies, No color"].firstMatch
    XCTAssertTrue(updated.waitForExistence(timeout: 10)); XCTAssertFalse(original.exists)
    XCTAssertTrue(updated.isHittable)
    XCTAssertGreaterThanOrEqual(updated.frame.minY, app.navigationBars["Tags"].frame.maxY)
    XCTAssertLessThanOrEqual(updated.frame.minY, app.navigationBars["Tags"].frame.maxY + 48)
    capture("settings-connected-collection-readback")
    updated.tap()
    XCTAssertTrue(name.waitForExistence(timeout: 10)); XCTAssertEqual(name.value as? String, "Tools emergency supplies")
    capture("settings-connected-reopened-editor")
    app.buttons["Back to settings collection"].firstMatch.tap()
    XCTAssertTrue(updated.waitForExistence(timeout: 10))
    app.buttons["Add Tag"].firstMatch.tap()
    XCTAssertTrue(name.waitForExistence(timeout: 10)); name.tap(); name.typeText("Camping")
    let committedName = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      name.value as? String == "Camping"
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [committedName], timeout: 15), .completed,
      "The complete typed name must settle without retyping before Save")
    app.buttons["Dismiss keyboard"].firstMatch.tap(); save.tap()
    let created = app.buttons["Camping, No color"].firstMatch
    XCTAssertTrue(created.waitForExistence(timeout: 10)); created.tap()
    XCTAssertTrue(name.waitForExistence(timeout: 10)); XCTAssertEqual(name.value as? String, "Camping")
    let archive = app.buttons["Archive"].firstMatch
    for _ in 0..<6 where !archive.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(archive.isHittable); archive.tap()
    let confirm = app.alerts["Archive Camping?"].buttons["Archive"].firstMatch
    XCTAssertTrue(confirm.waitForExistence(timeout: 5)); confirm.tap()
    XCTAssertTrue(updated.waitForExistence(timeout: 10)); XCTAssertFalse(created.exists)
    capture("settings-connected-created-tag-archived")
  }

  func testSettingsCollectionUsesNativeSearchAndAdd() {
    let open = app.buttons["Audit settings collection"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    XCTAssertTrue(app.navigationBars["Tags"].waitForExistence(timeout: 5), "Match the production route title before evaluating native header placement")
    let add = app.buttons["Add Tag"]
    XCTAssertTrue(add.waitForExistence(timeout: 10))
    XCTAssertTrue(add.isHittable)
    add.tap()
    XCTAssertTrue(app.staticTexts["Add tag requested"].waitForExistence(timeout: 5))
    XCTAssertEqual(app.textFields.count, 0)
    let search = app.buttons["Search"].firstMatch
    XCTAssertTrue(search.isHittable)
    search.tap()
    let field = app.searchFields.firstMatch
    XCTAssertTrue(field.waitForExistence(timeout: 5))
    waitForKeyboard()
    field.typeText("Tools")
    XCTAssertEqual(field.value as? String, "Tools")
    XCTAssertTrue(app.buttons["Tools, No color"].waitForExistence(timeout: 5))
    XCTAssertTrue(app.buttons["Garden, No color"].waitForNonExistence(timeout: 5))
    capture("settings-collection-native-search")
    resetNativeSearch(field)
    XCTAssertTrue(app.buttons["Garden, No color"].waitForExistence(timeout: 5))
    XCTAssertTrue(add.isHittable)
    capture("settings-collection-native-header")
  }

  private func resetNativeSearch(_ field: XCUIElement) {
    if UIDevice.current.userInterfaceIdiom == .pad {
      let clear = field.buttons["Clear text"]
      XCTAssertTrue(clear.isHittable)
      clear.tap()
      if !app.keyboards.firstMatch.waitForNonExistence(timeout: 2) {
        let dismiss = app.buttons["Dismiss keyboard"]
        XCTAssertTrue(dismiss.isHittable)
        dismiss.tap()
      }
    } else {
      let cancel = app.buttons.matching(NSPredicate(format: "label IN %@", ["Cancel", "Close search", "Close"])).firstMatch
      XCTAssertTrue(cancel.isHittable)
      cancel.tap()
      XCTAssertTrue(field.waitForNonExistence(timeout: 5))
    }
    XCTAssertTrue(app.keyboards.firstMatch.waitForNonExistence(timeout: 5))
  }

  private func openCustomizationEditor() {
    let open = app.buttons["Audit settings editor"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    XCTAssertTrue(app.textFields["Name"].waitForExistence(timeout: 10))
    XCTAssertEqual(app.textFields["Name"].value as? String, "Tools")
  }

  func testVoiceProposalLocationSearchRetryAndReturn() {
    let open = app.buttons["Audit voice proposal"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    let original = app.buttons["Change containing location, currently Inventory root"]
    XCTAssertTrue(original.waitForExistence(timeout: 15))
    let context = app.staticTexts["Audit inventory · Audit home"].firstMatch
    XCTAssertTrue(context.exists)
    let conversationHeader = app.navigationBars["Conversation"]
    XCTAssertGreaterThanOrEqual(context.frame.minY, conversationHeader.frame.maxY,
      "The inventory context must be fully below native navigation chrome")
    XCTAssertLessThanOrEqual(context.frame.maxY, app.frame.maxY)
    capture("voice-proposal-entry-context")
    revealVoiceProposalLocation(original)
    original.tap()
    let header = app.navigationBars["Containing location"]
    XCTAssertTrue(header.waitForExistence(timeout: 10))
    let retry = app.buttons["Retry locations"]
    XCTAssertTrue(retry.waitForExistence(timeout: 10))
    XCTAssertTrue(app.staticTexts["Could not load locations."].exists)
    XCTAssertTrue(retry.isHittable)
    capture("voice-location-retry"); retry.tap()
    let bin = app.descendants(matching: .any).matching(identifier: "Select Garage bin, Garage / Garage bin").firstMatch
    XCTAssertTrue(bin.waitForExistence(timeout: 10))
    let search = app.buttons["Search"].firstMatch
    XCTAssertTrue(search.isHittable); search.tap()
    let field = app.searchFields.firstMatch
    XCTAssertTrue(field.waitForExistence(timeout: 5))
    waitForKeyboard(); field.typeText("missing")
    XCTAssertEqual(field.value as? String, "missing")
    XCTAssertTrue(app.staticTexts["No matching locations"].waitForExistence(timeout: 10))
    XCTAssertTrue(bin.waitForNonExistence(timeout: 5))
    capture("voice-location-empty-search")
    let clear = field.buttons["Clear text"]
    XCTAssertTrue(clear.isHittable); clear.tap()
    let clearedField = app.searchFields.firstMatch
    let idleSearch = header.buttons["Search"].firstMatch
    let available = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      (clearedField.exists && clearedField.isHittable) || (idleSearch.exists && idleSearch.isHittable)
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [available], timeout: 5), .completed,
      "Clearing must leave a usable field or collapsed Search control")
    XCTAssertTrue(bin.waitForExistence(timeout: 10), "Clearing must restore unfiltered locations")
    capture("voice-location-after-focused-clear")
    if !clearedField.isHittable {
      XCTAssertGreaterThanOrEqual(idleSearch.frame.minX, header.frame.minX)
      XCTAssertLessThanOrEqual(idleSearch.frame.maxX, header.frame.maxX)
      XCTAssertGreaterThanOrEqual(idleSearch.frame.minY, header.frame.minY)
      XCTAssertLessThanOrEqual(idleSearch.frame.maxY, header.frame.maxY)
      idleSearch.tap()
    } else if !app.keyboards.firstMatch.exists {
      clearedField.tap()
    }
    XCTAssertTrue(clearedField.waitForExistence(timeout: 5), "Cleared search must accept a fresh query")
    XCTAssertTrue(clearedField.isHittable); waitForKeyboard(); clearedField.typeText("Garage")
    XCTAssertEqual(clearedField.value as? String, "Garage")
    XCTAssertTrue(bin.waitForExistence(timeout: 10))
    XCTAssertTrue(bin.isHittable); bin.tap()
    let changed = app.buttons["Change containing location, currently Garage / Garage bin"]
    XCTAssertTrue(changed.waitForExistence(timeout: 10))
    revealVoiceProposalLocation(changed)
    capture("voice-proposal-selected-location"); changed.tap()
    XCTAssertTrue(header.waitForExistence(timeout: 10))
    XCTAssertTrue(app.staticTexts["Garage / Garage bin"].firstMatch.waitForExistence(timeout: 5))
    let back = header.buttons["Back to conversation"].firstMatch
    XCTAssertTrue(back.isHittable); back.tap()
    XCTAssertTrue(changed.waitForExistence(timeout: 10))
    XCTAssertFalse(original.exists)
    revealVoiceProposalLocation(changed)
    capture("voice-proposal-location-back")
  }

  private func revealVoiceProposalLocation(_ control: XCUIElement) {
    let scroll = app.scrollViews.containing(.button, identifier: control.label).firstMatch
    XCTAssertTrue(scroll.exists, "The proposal location must belong to the conversation scroll view")
    let header = app.navigationBars["Conversation"]
    func visibleViewport() -> CGRect {
      let bounds = scroll.frame.intersection(app.frame)
      let top = max(bounds.minY, header.frame.maxY)
      return CGRect(x: bounds.minX, y: top, width: bounds.width, height: max(0, bounds.maxY - top))
    }
    func visible() -> Bool {
      let viewport = visibleViewport()
      return control.isHittable && control.frame.minY >= viewport.minY &&
        control.frame.maxY <= viewport.maxY && control.frame.minX >= viewport.minX &&
        control.frame.maxX <= viewport.maxX
    }
    for _ in 0..<18 where !visible() {
      let viewport = visibleViewport()
      guard viewport.height > 0 else { break }
      let above = control.frame.minY < viewport.minY
      let origin = app.coordinate(withNormalizedOffset: .zero)
      func point(_ fraction: CGFloat) -> XCUICoordinate {
        origin.withOffset(CGVector(dx: viewport.midX - app.frame.minX,
          dy: viewport.minY + viewport.height * fraction - app.frame.minY))
      }
      let start = point(above ? 0.25 : 0.75)
      let end = point(above ? 0.75 : 0.25)
      start.press(forDuration: 0.05, thenDragTo: end)
    }
    XCTAssertTrue(visible(), "The complete proposal location control must be reachable within its sheet")
  }

  func testVoiceNativeHeaderKeepsProposalOnCloseAndCancelledReset() {
    let open = app.buttons["Audit voice proposal"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    let header = app.navigationBars["Conversation"]
    XCTAssertTrue(header.waitForExistence(timeout: 10))
    let proposal = app.buttons["Change containing location, currently Inventory root"]
    XCTAssertTrue(proposal.waitForExistence(timeout: 15))
    let newConversation = header.buttons["New conversation"]
    let close = header.buttons["Close voice session"]
    XCTAssertTrue(newConversation.isHittable); XCTAssertTrue(close.isHittable)
    newConversation.tap()
    let confirmation = app.alerts["Start a new conversation?"]
    XCTAssertTrue(confirmation.waitForExistence(timeout: 5))
    confirmation.buttons["Keep conversation"].tap()
    XCTAssertTrue(proposal.exists)
    capture("voice-native-header-protected-proposal")
    XCTAssertTrue(close.isHittable); close.tap()
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
    XCTAssertTrue(open.isHittable); open.tap()
    XCTAssertTrue(header.waitForExistence(timeout: 10))
    XCTAssertTrue(proposal.waitForExistence(timeout: 10))
    capture("voice-native-header-returned-proposal")
  }

  func testSettingsEditorNativeBackProtectsDirtyDraft() {
    openCustomizationEditor()
    let name = app.textFields["Name"]
    name.tap(); waitForKeyboard(); name.typeText("x")
    XCTAssertEqual(name.value as? String, "Toolsx")
    let back = app.buttons["Back to settings collection"]
    XCTAssertTrue(back.isHittable)
    back.tap()
    let alert = app.alerts["Discard changes?"]
    XCTAssertTrue(alert.waitForExistence(timeout: 5))
    alert.buttons["Keep Editing"].tap()
    XCTAssertEqual(name.value as? String, "Toolsx")
    back.tap()
    XCTAssertTrue(alert.waitForExistence(timeout: 5))
    capture("settings-editor-native-discard")
    alert.buttons["Discard"].tap()
    XCTAssertTrue(app.buttons["Add Tag"].waitForExistence(timeout: 10))
    XCTAssertFalse(name.exists)
  }

  func testSettingsFullNameSurvivesRejectedSaveAndRetry() {
    let open = app.buttons["Audit settings save recovery"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    let name = app.textFields["Name"].firstMatch
    XCTAssertTrue(name.waitForExistence(timeout: 10))
    XCTAssertEqual(name.value as? String, "Tools")
    name.tap()
    XCTAssertTrue(app.keyboards.firstMatch.waitForExistence(timeout: 5))
    name.typeText(" emergency supplies")
    XCTAssertEqual(observePredicate("settings-complete-name",
      predicate: NSPredicate(format: "value == %@", "Tools emergency supplies"), object: name), .completed)
    let dismiss = app.buttons["Dismiss keyboard"].firstMatch
    XCTAssertTrue(dismiss.isHittable); dismiss.tap()
    let save = app.buttons["Save"].firstMatch
    XCTAssertTrue(save.isHittable); XCTAssertTrue(save.isEnabled); save.tap()
    XCTAssertTrue(app.staticTexts["Could not save"].waitForExistence(timeout: 10))
    XCTAssertEqual(name.value as? String, "Tools emergency supplies")
    XCTAssertTrue(name.isEnabled)
    XCTAssertEqual(observePredicate("settings-save-recovered",
      predicate: NSPredicate(format: "enabled == true"), object: save), .completed)
    capture("settings-full-name-rejected-save-retained")
    XCTAssertTrue(save.isHittable); save.tap()
    assertCustomizationNotice("Tag saved")
    XCTAssertTrue(app.buttons["Add Tag"].waitForExistence(timeout: 10))
    XCTAssertFalse(name.exists)
  }

  func testSettingsEditorNativeSaveReturnsToCollection() {
    openCustomizationEditor()
    let name = app.textFields["Name"]
    name.tap(); waitForKeyboard(); name.typeText("x")
    XCTAssertEqual(name.value as? String, "Toolsx")
    let dismiss = app.buttons["Dismiss keyboard"].firstMatch
    XCTAssertTrue(dismiss.isHittable)
    dismiss.tap()
    let save = app.buttons["Save"].firstMatch
    for _ in 0..<6 where !save.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(save.isHittable)
    XCTAssertTrue(save.isEnabled)
    capture("settings-editor-native-save")
    save.tap()
    XCTAssertTrue(app.buttons["Add Tag"].waitForExistence(timeout: 10))
    assertCustomizationNotice("Tag saved")
  }

  func testSettingsEditorNativeArchiveConfirmsBeforeReturning() {
    openCustomizationEditor()
    let archive = app.buttons["Archive"].firstMatch
    for _ in 0..<6 where !archive.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(archive.isHittable)
    archive.tap()
    let alert = app.alerts["Archive Tools?"]
    XCTAssertTrue(alert.waitForExistence(timeout: 5))
    alert.buttons["Cancel"].tap()
    XCTAssertTrue(archive.isEnabled)
    archive.tap()
    XCTAssertTrue(alert.waitForExistence(timeout: 5))
    capture("settings-editor-native-archive")
    alert.buttons["Archive"].tap()
    assertCustomizationNotice("Tag archived")
    XCTAssertTrue(app.buttons["Add Tag"].waitForExistence(timeout: 10))
  }

  private func assertCustomizationNotice(_ title: String) {
    let notice = app.descendants(matching: .any).matching(identifier: "app-notice-container").firstMatch
    let expected = XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == true AND label CONTAINS %@", title), object: notice)
    XCTAssertEqual(XCTWaiter.wait(for: [expected], timeout: 5), .completed)
  }

  func testColorWellTargetOpensSystemPicker() {
    openSettingsControls()
    let picker = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose any color")).firstMatch
    XCTAssertTrue(picker.waitForExistence(timeout: 5))
    XCTAssertTrue(picker.isHittable)
    // The system well's AX frame is smaller than its delivered touch region.
    // testColorWellDeliveredTouchRegion independently probes the surrounding area.
    XCTAssertGreaterThan(picker.frame.width, 0)
    XCTAssertGreaterThan(picker.frame.height, 0)
    XCTAssertTrue(app.frame.contains(picker.frame), "The complete visible well must remain onscreen")
    XCTAssertLessThanOrEqual(picker.frame.width, picker.frame.height + 1,
      "The accessible target must be the well, not a wide inactive label row")
    picker.tap()
    XCTAssertTrue(app.buttons["Sliders"].waitForExistence(timeout: 5))
    capture("color-visible-well-target")
  }

  func testColorWellHasSingleAccessibleName() {
    openSettingsControls()
    let picker = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose any color")).firstMatch
    XCTAssertTrue(picker.waitForExistence(timeout: 5))
    XCTAssertEqual(picker.label, "Choose any color", "The native color well must not repeat its name")
    capture("color-well-accessible-name")
  }

  func testColorWellDeliveredTouchRegion() {
    openSettingsControls()
    let picker = app.buttons.matching(NSPredicate(format: "label == %@", "Choose any color")).firstMatch
    let offsets: [(CGFloat, CGFloat)] = [(0, 0), (0, -21), (0, 21), (-21, 0), (21, 0),
                                          (-21, -21), (21, -21), (-21, 21), (21, 21)]
    for (index, offset) in offsets.enumerated() {
      XCTAssertTrue(picker.waitForExistence(timeout: 5))
      XCTAssertTrue(picker.isHittable)
      let center = picker.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
      center.withOffset(CGVector(dx: offset.0, dy: offset.1)).tap()
      let sliders = app.buttons["Sliders"]
      guard sliders.waitForExistence(timeout: 5) else {
        capture("color-hit-region-failed-probe-\(index)")
        XCTFail("Color well did not open for probe \(index): \(offset)")
        return
      }
      if UIDevice.current.userInterfaceIdiom == .pad {
        let dismiss = app.otherElements["PopoverDismissRegion"]
        XCTAssertTrue(dismiss.exists)
        dismiss.coordinate(withNormalizedOffset: CGVector(dx: 0.1, dy: 0.9)).tap()
      } else {
        app.buttons["close"].tap()
      }
      guard sliders.waitForNonExistence(timeout: 5) else {
        capture("color-hit-region-dismissal-failed-\(index)")
        XCTFail("System color picker did not dismiss after probe \(index)")
        return
      }
      XCTAssertTrue(app.staticTexts["Color value: none"].exists,
                    "Opening the picker must preserve the unset parent value")
    }
    capture("color-delivered-hit-region")
  }

  func testExactExpirationUsesCompactNativePicker() {
    openSettingsControls()
    app.buttons["Expiration"].tap()
    XCTAssertTrue(app.staticTexts["Expiration value: No expiration"].exists)
    app.buttons["Add expiration date"].tap()
    let picker = app.datePickers.firstMatch
    XCTAssertTrue(picker.waitForExistence(timeout: 5))
    XCTAssertFalse(app.buttons["Use date"].exists)
    picker.tap()
    capture("compact-asset-expiration-calendar")
    app.navigationBars.firstMatch.tap()
    let clear = app.buttons["Clear expiration"]
    if !clear.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(clear.isHittable)
    clear.tap()
    XCTAssertTrue(app.staticTexts["Expiration value: No expiration"].exists)
  }

  func testActionableFeedbackStaysAvailable() throws {
    app.buttons["Audit feedback"].tap()
    let retry = app.buttons["Retry audit action"]
    XCTAssertTrue(retry.waitForExistence(timeout: 5))
    let vanished = XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: retry)
    XCTAssertEqual(XCTWaiter.wait(for: [vanished], timeout: 8), .timedOut)
    capture("persistent-actionable-feedback")
    XCTAssertTrue(retry.isHittable)
    retry.tap()
    XCTAssertTrue(app.staticTexts["Audit retry completed"].waitForExistence(timeout: 5))
  }

  func testAccountNativeCommandCancelAndRecovery() { verifySessionSettings("Account", command: "Sign Out") }
  func testConnectionNativeCommandCancelAndRecovery() { verifySessionSettings("Connection", command: "Change Server") }

  private func verifySessionSettings(_ kind: String, command: String) {
    let open = app.buttons["Audit \(kind)"]
    for _ in 0..<15 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
    let header = app.navigationBars[kind]
    XCTAssertTrue(header.waitForExistence(timeout: 10))
    let action = app.buttons[command].firstMatch
    XCTAssertTrue(action.waitForExistence(timeout: 10))
    XCTAssertTrue(action.isHittable)
    XCTAssertGreaterThanOrEqual(action.frame.height, 44)
    XCTAssertGreaterThanOrEqual(action.frame.minY, header.frame.maxY)
    XCTAssertLessThanOrEqual(action.frame.maxY, app.frame.maxY)
    capture("session-\(kind)-command")
    action.tap()
    let alert = app.alerts.firstMatch
    XCTAssertTrue(alert.waitForExistence(timeout: 5))
    alert.buttons["Cancel"].tap()
    XCTAssertTrue(alert.waitForNonExistence(timeout: 5))
    XCTAssertTrue(action.isEnabled)
    action.tap(); XCTAssertTrue(alert.waitForExistence(timeout: 5))
    alert.buttons[command].tap()
    let notice = app.descendants(matching: .any).matching(identifier: "app-notice-container").firstMatch
    XCTAssertTrue(notice.waitForExistence(timeout: 5))
    XCTAssertTrue(notice.label.contains("Audit session action unavailable. Try again."))
    XCTAssertTrue(action.isEnabled); XCTAssertTrue(action.isHittable)
    XCTAssertTrue(header.buttons.element(boundBy: 0).isHittable)
    let placement = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      let rect = notice.frame
      return rect.height > 0 && rect.minY >= header.frame.maxY && rect.maxY <= self.app.frame.maxY &&
        rect.minX >= self.app.frame.minX && rect.maxX <= self.app.frame.maxX
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [placement], timeout: 5), .completed, "Entire failure notice must remain visible below navigation")
    capture("session-\(kind)-recovery")
    action.tap(); XCTAssertTrue(alert.waitForExistence(timeout: 5))
    alert.buttons[command].tap()
    XCTAssertTrue(header.waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.navigationBars["Native UI audit"].exists)
  }

}
