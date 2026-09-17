import XCTest
import UIKit

final class FixtureAuditTests: XCTestCase {
  private let app = XCUIApplication(bundleIdentifier: "org.stuffstash.mobile")
  override func setUpWithError() throws {
    continueAfterFailure = false
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
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
    reveal(create); create.tap()
    feedback("Invitation created, link unavailable", message: "If the invitation is still pending below, cancel it before retrying. If you already cancelled it, try again.", captureName: "sharing-unavailable-link")
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
    let open = app.buttons["Audit footer appearance"]
    for _ in 0..<16 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
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

  func testMoveHereRejectedCommandRetainsSelectionAndRetryReturns() {
    let open = app.buttons["Audit Move here recovery"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let retry = app.buttons["Retry suggestions"].firstMatch
    XCTAssertTrue(retry.waitForExistence(timeout: 10))
    retry.tap()
    // Native accessibility groups title, kind and location into the candidate button.
    let candidate = app.buttons["Audit tent, Item, Garage"].firstMatch
    XCTAssertTrue(candidate.waitForExistence(timeout: 5))
    XCTAssertTrue(candidate.isEnabled)
    XCTAssertTrue(candidate.isHittable)
    candidate.tap()
    let move = app.buttons["Move here"].firstMatch
    XCTAssertTrue(move.isEnabled)
    move.tap()
    XCTAssertTrue(app.alerts["Could not move asset here"].waitForExistence(timeout: 5))
    app.alerts.buttons["OK"].tap()
    XCTAssertTrue(app.staticTexts["Audit tent -> Camping box"].firstMatch.exists)
    XCTAssertTrue(move.isHittable)
    move.tap()
    XCTAssertTrue(app.textFields["Find item, box, or place"].waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.navigationBars["Native UI audit"].waitForExistence(timeout: 5))
    XCTAssertTrue(open.isHittable)
    XCTAssertEqual(app.state, .runningForeground)
    capture("move-here-successful-retry")
  }

  func testMoveHereRecoveryAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
    let open = app.buttons["Audit Move here recovery"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let query = app.textFields["Find item, box, or place"]
    XCTAssertTrue(query.waitForExistence(timeout: 10))
    let form = app.scrollViews.containing(.textField, identifier: "Find item, box, or place").firstMatch
    XCTAssertTrue(form.exists)
    func queryVisible() -> Bool {
      let bounds = form.frame.intersection(app.frame)
      return query.isHittable && query.frame.minY >= bounds.minY && query.frame.maxY <= bounds.maxY
    }
    for _ in 0..<12 where !queryVisible() {
      let above = query.frame.minY < form.frame.minY
      form.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.4 : 0.7))
        .press(forDuration: 0.05, thenDragTo: form.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.7 : 0.4)))
    }
    XCTAssertTrue(queryVisible())
    query.tap()
    waitForKeyboard()
    query.typeText("Tent")
    XCTAssertEqual(query.value as? String, "Tent")
    let dismiss = app.buttons["Dismiss keyboard"]
    XCTAssertTrue(dismiss.isHittable)
    dismiss.tap()
    XCTAssertTrue(app.keyboards.firstMatch.waitForNonExistence(timeout: 5))
    let retry = app.buttons["Retry suggestions"].firstMatch
    XCTAssertTrue(retry.waitForExistence(timeout: 10))
    XCTAssertFalse(app.staticTexts["No movable matches"].exists)
    let results = app.scrollViews.containing(.button, identifier: "Retry suggestions").firstMatch
    XCTAssertTrue(results.exists)
    for _ in 0..<8 where !retry.isHittable { results.swipeUp() }
    XCTAssertTrue(retry.isHittable)
    XCTAssertTrue(app.buttons["Cancel"].firstMatch.isHittable)
    capture("move-here-suggestions-error")
    retry.tap()
    XCTAssertTrue(app.buttons["Audit tent, Item, Garage"].firstMatch.waitForExistence(timeout: 5))
    XCTAssertEqual(query.value as? String, "Tent")
    capture("move-here-suggestions-recovered")
    app.buttons["Cancel"].firstMatch.tap()
    XCTAssertTrue(query.waitForNonExistence(timeout: 5))
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

  private func waitForKeyboard() {
    let keyboard = app.keyboards.firstMatch
    XCTAssertTrue(keyboard.waitForExistence(timeout: 5))
    let ready = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      keyboard.keys.allElementsBoundByIndex.contains { key in
        guard key.exists else { return false }
        let bounds = key.frame
        guard !bounds.isEmpty, !bounds.isNull, !bounds.isInfinite,
              bounds.origin.x.isFinite, bounds.origin.y.isFinite,
              bounds.width.isFinite, bounds.height.isFinite else { return false }
        return key.isHittable
      }
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [ready], timeout: 5), .completed, "Typing requires an interactive keyboard")
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
    XCTAssertEqual(XCTWaiter.wait(for: [clearAccessory], timeout: 5), .completed,
      "Both filter commands must be fully above the keyboard-dismiss accessory")
  }

  func testBrowseTagSearchKeepsActionsAboveKeyboardAccessory() {
    app.buttons["Audit Browse filters"].tap()
    let tags = app.buttons["Choose tags"]
    XCTAssertTrue(tags.waitForExistence(timeout: 5)); tags.tap()
    let searchButton = app.buttons["Search"].firstMatch
    XCTAssertTrue(searchButton.waitForExistence(timeout: 5)); searchButton.tap()
    let search = app.searchFields.firstMatch
    XCTAssertTrue(search.waitForExistence(timeout: 5)); search.tap()
    waitForKeyboard()
    search.typeText("Tools")
    let completeQuery = XCTNSPredicateExpectation(predicate: NSPredicate(format: "value == %@", "Tools"), object: search)
    XCTAssertEqual(XCTWaiter.wait(for: [completeQuery], timeout: 5), .completed, "Native search must retain the complete query")
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

  func testDetailCommandsRemainReachableAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
    let open = app.buttons["Audit detail commands"]
    for _ in 0..<14 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let add = app.buttons["Add item here"].firstMatch
    XCTAssertTrue(add.waitForExistence(timeout: 10))
    for label in ["Add item here", "Move items here", "Check out", "Edit", "Move"] {
      let command = app.buttons[label].firstMatch
      XCTAssertTrue(command.exists)
      let scroll = app.scrollViews.firstMatch
      func fullyVisible() -> Bool {
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
      XCTAssertGreaterThanOrEqual(command.frame.height, 44, label)
      XCTAssertGreaterThanOrEqual(command.frame.minX, app.frame.minX, label)
      XCTAssertLessThanOrEqual(command.frame.maxX, app.frame.maxX, label)
      if label == "Add item here" {
        XCTAssertGreaterThan(command.frame.width, app.frame.width * 0.5)
      }
      capture("detail-command-" + label.lowercased().replacingOccurrences(of: " ", with: "-"))
    }
    let back = app.navigationBars.buttons.firstMatch
    XCTAssertTrue(back.isHittable)
    back.tap()
    XCTAssertTrue(open.waitForExistence(timeout: 5))
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
    capture("place-search-collapsed")
    searchButton.tap()
    let field = app.searchFields.firstMatch
    XCTAssertTrue(field.waitForExistence(timeout: 5))
    XCTAssertTrue(field.isHittable)
    XCTAssertEqual(field.placeholderValue, "Search this place")
    field.tap()
    waitForKeyboard()
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
      field.isHittable || (!field.exists && searchButton.isHittable)
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [cleared], timeout: 5), .completed)
    capture("place-search-cleared")
    if !field.exists {
      XCTAssertTrue(searchButton.isHittable)
      searchButton.tap()
      XCTAssertTrue(field.waitForExistence(timeout: 5))
    }
    XCTAssertTrue(field.isHittable)
    field.tap()
    waitForKeyboard()
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
    field.tap()
    waitForKeyboard()
    field.typeText("missing")
    XCTAssertEqual(field.value as? String, "missing")
    capture("static-search-before-focused-clear")
    let clear = field.buttons["Clear text"].firstMatch
    XCTAssertTrue(clear.isHittable)
    clear.tap()
    let available = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      field.isHittable || (!field.exists && search.isHittable)
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [available], timeout: 5), .completed)
    capture("static-search-after-focused-clear")
    if !field.exists { search.tap() }
    XCTAssertTrue(field.waitForExistence(timeout: 5))
    XCTAssertTrue(field.isHittable)
    field.tap()
    waitForKeyboard()
    field.typeText("Garage")
    XCTAssertEqual(field.value as? String, "Garage")
    capture("static-search-fresh-query")
  }

  func testAssetRegionRecoveryAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
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
    capture("asset-region-photo-error-accessibility-size")
    photos.tap()
    XCTAssertTrue(photos.waitForNonExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["No photos"].firstMatch.waitForExistence(timeout: 5))
    reveal(app.staticTexts["No photos"].firstMatch)
    capture("asset-region-photo-recovered-accessibility-size")
    reveal(contents)
    capture("asset-region-contents-error-accessibility-size")
    XCTAssertTrue(app.staticTexts["Could not load contents."].firstMatch.exists)
    contents.tap()
    XCTAssertTrue(contents.waitForNonExistence(timeout: 5))
    let empty = app.staticTexts["Nothing here yet"].firstMatch
    XCTAssertTrue(empty.waitForExistence(timeout: 5))
    reveal(empty)
    capture("asset-region-recovered-accessibility-size")
    let back = app.navigationBars.buttons.firstMatch
    XCTAssertTrue(back.isHittable)
    back.tap()
    XCTAssertTrue(open.waitForExistence(timeout: 5))
  }

  func testEditTagDisclosureAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
    let open = app.buttons["Audit Edit tags"]
    for _ in 0..<12 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
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
        let start = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.4 : 0.7))
        let end = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.7 : 0.4))
        start.press(forDuration: 0.05, thenDragTo: end)
      }
      XCTAssertTrue(visible())
      XCTAssertTrue(cancel.isHittable)
      XCTAssertGreaterThanOrEqual(cancel.frame.minY, app.frame.minY)
      XCTAssertLessThanOrEqual(cancel.frame.maxY, app.frame.maxY)
    }
    let retained = app.buttons["Tag 14"].firstMatch
    XCTAssertTrue(retained.waitForExistence(timeout: 5))
    XCTAssertFalse(app.buttons["Tag 13"].exists)
    reveal(retained)
    XCTAssertTrue(retained.isSelected)
    let entry = app.textFields["New tag name"].firstMatch
    reveal(entry)
    entry.tap()
    waitForKeyboard()
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
    capture("edit-unstaged-tag-retained-accessibility-size")
    let add = app.buttons["Add tag"].firstMatch
    reveal(add)
    add.tap()
    XCTAssertTrue(["", "New tag"].contains(entry.value as? String ?? "missing"))
    XCTAssertTrue(app.buttons["Save"].firstMatch.isEnabled)
    let expand = app.buttons["Show all tags"].firstMatch
    reveal(expand)
    expand.tap()
    let extra = app.buttons["Tag 13"].firstMatch
    XCTAssertTrue(extra.waitForExistence(timeout: 5))
    reveal(extra)
    extra.tap()
    XCTAssertTrue(extra.isSelected)
    capture("edit-tags-expanded-accessibility-size")
    let collapse = app.buttons["Show fewer tags"].firstMatch
    reveal(collapse)
    collapse.tap()
    reveal(extra)
    XCTAssertTrue(extra.isSelected)
    XCTAssertTrue(retained.isSelected)
    capture("edit-tags-collapsed-selected-accessibility-size")
    cancel.tap()
    let discard = app.alerts.buttons["Discard"]
    XCTAssertTrue(discard.waitForExistence(timeout: 5))
    discard.tap()
    XCTAssertTrue(open.waitForExistence(timeout: 5))
  }

  func testEditMetadataRecoveryAtAccessibilityTextSize() {
    app.terminate()
    app.launchArguments = ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"]
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
    let open = app.buttons["Audit Edit recovery"]
    for _ in 0..<10 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
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
    waitForKeyboard()
    search.typeText("Tools")
    let completeQuery = XCTNSPredicateExpectation(predicate: NSPredicate(format: "value == %@", "Tools"), object: search)
    XCTAssertEqual(XCTWaiter.wait(for: [completeQuery], timeout: 5), .completed, "Native search must retain the complete query")
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
      func visible() -> Bool {
        let bounds = scroll.frame.intersection(app.frame)
        let top = max(bounds.minY, app.navigationBars["Add item"].frame.maxY)
        return (!requiresHit || element.isHittable) && element.frame.minY >= top && element.frame.maxY <= bounds.maxY
      }
      for _ in 0..<18 where !visible() {
        let above = element.frame.minY < app.navigationBars["Add item"].frame.maxY
        let start = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.4 : 0.7))
        let end = scroll.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: above ? 0.7 : 0.4))
        start.press(forDuration: 0.05, thenDragTo: end)
      }
      XCTAssertTrue(visible())
    }
    func dismissKeyboard() {
      let dismiss = app.buttons["Dismiss keyboard"].firstMatch
      XCTAssertTrue(dismiss.isHittable)
      dismiss.tap()
      XCTAssertTrue(app.keyboards.firstMatch.waitForNonExistence(timeout: 5))
    }
    reveal(name)
    name.tap()
    waitForKeyboard()
    name.typeText("Tent")
    let completeName = XCTNSPredicateExpectation(predicate: NSPredicate(format: "value == %@", "Tent"), object: name)
    XCTAssertEqual(XCTWaiter.wait(for: [completeName], timeout: 5), .completed)
    XCTAssertEqual(name.value as? String, "Tent")
    dismissKeyboard()
    let save = app.buttons["Save item"].firstMatch
    XCTAssertTrue(save.isEnabled)
    let details = app.buttons["More details"].firstMatch
    reveal(details)
    details.tap()
    let entry = app.textFields["New tag name"].firstMatch
    reveal(entry)
    entry.tap()
    waitForKeyboard()
    entry.typeText("Camping")
    let completeTag = XCTNSPredicateExpectation(predicate: NSPredicate(format: "value == %@", "Camping"), object: entry)
    XCTAssertEqual(XCTWaiter.wait(for: [completeTag], timeout: 5), .completed)
    XCTAssertEqual(entry.value as? String, "Camping")
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
      guard entry.exists, staged.exists, save.exists else { return false }
      let value = entry.value
      return (value == nil || value as? String == "" || value as? String == "New tag") && save.isEnabled
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [stageComplete], timeout: 5), .completed,
      "Staging must retain the tag, clear the existing entry and enable Save")
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
    field.tap()
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
    field.tap(); waitForKeyboard(); field.typeText("missing")
    XCTAssertEqual(field.value as? String, "missing")
    XCTAssertTrue(app.staticTexts["No matching locations"].waitForExistence(timeout: 10))
    XCTAssertTrue(bin.waitForNonExistence(timeout: 5))
    capture("voice-location-empty-search")
    let clear = field.buttons["Clear text"]
    XCTAssertTrue(clear.isHittable); clear.tap()
    let clearedField = app.searchFields.firstMatch
    let idleSearch = header.buttons["Search"].firstMatch
    if UIDevice.current.userInterfaceIdiom == .pad {
      let available = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
        clearedField.isHittable || (!clearedField.exists && idleSearch.isHittable)
      }, object: nil)
      XCTAssertEqual(XCTWaiter.wait(for: [available], timeout: 5), .completed,
        "Clearing must leave a field or the native iPad search button available")
      XCTAssertTrue(bin.waitForExistence(timeout: 10), "Clearing must restore unfiltered locations")
      capture("voice-location-after-focused-clear")
      if !clearedField.exists {
        XCTAssertGreaterThanOrEqual(idleSearch.frame.minX, header.frame.minX)
        XCTAssertLessThanOrEqual(idleSearch.frame.maxX, header.frame.maxX)
        XCTAssertGreaterThanOrEqual(idleSearch.frame.minY, header.frame.minY)
        XCTAssertLessThanOrEqual(idleSearch.frame.maxY, header.frame.maxY)
        idleSearch.tap()
      }
    }
    XCTAssertTrue(clearedField.waitForExistence(timeout: 5), "Cleared search must accept a fresh query")
    XCTAssertTrue(clearedField.isHittable); clearedField.tap(); waitForKeyboard(); clearedField.typeText("Garage")
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
    XCTAssertTrue(app.buttons["Add Tag"].waitForExistence(timeout: 10))
    assertCustomizationNotice("Tag archived")
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
