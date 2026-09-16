import XCTest
import UIKit

final class OnboardingAuditTests: XCTestCase {
  private let app = XCUIApplication(bundleIdentifier: "org.stuffstash.mobile")

  override func setUpWithError() throws {
    continueAfterFailure = false
    app.launch()
  }

  override func tearDownWithError() throws {
    capture("final-state")
    app.terminate()
  }

  private func waitForKeyboard() {
    let keyboard = app.keyboards.firstMatch
    XCTAssertTrue(keyboard.waitForExistence(timeout: 5))
    let ready = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      // This fixture uses the English URL keyboard. Avoid resolving every key,
      // including zero-sized padding nodes, during each readiness observation.
      keyboard.keys["q"].isHittable
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [ready], timeout: 5), .completed, "Typing requires an interactive keyboard")
  }

  private func capture(_ name: String) {
    let attachment = XCTAttachment(screenshot: app.screenshot())
    attachment.name = name
    attachment.lifetime = .keepAlways
    add(attachment)
    if name.hasPrefix("onboarding-landscape-") {
      let screen = XCTAttachment(screenshot: XCUIScreen.main.screenshot())
      screen.name = "\(name)-full-screen"
      screen.lifetime = .keepAlways
      add(screen)
    }
    let hierarchy = XCTAttachment(string: app.debugDescription)
    hierarchy.name = "\(name)-hierarchy"
    hierarchy.lifetime = .keepAlways
    add(hierarchy)
  }

  func testOnboardingAdaptsToLandscape() throws {
    try XCTSkipUnless(UIDevice.current.userInterfaceIdiom == .pad, "The shipped iPhone app supports portrait only")
    XCUIDevice.shared.orientation = .landscapeLeft
    defer { XCUIDevice.shared.orientation = .portrait }
    let address = app.textFields["Server address"]
    XCTAssertTrue(address.waitForExistence(timeout: 30))
    XCTAssertTrue(address.isHittable)
    XCTAssertGreaterThan(app.frame.width, app.frame.height, "The iPad window must actually rotate")
    capture("onboarding-landscape-entry")
    let connect = app.buttons["Connect and sign in"]
    if !connect.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(connect.isHittable, "The form action must remain reachable in landscape")
    capture("onboarding-landscape-action")
  }

  func testIPadKeyboardDismissalFromInsideFormColumn() throws {
    try XCTSkipUnless(UIDevice.current.userInterfaceIdiom == .pad, "Comparison targets the centered iPad form")
    verifyConnectionHelpAndKeyboard(formColumnDrag: true)
  }

  func testConnectionHelpAndKeyboardKeepActionsReachable() throws {
    verifyConnectionHelpAndKeyboard(formColumnDrag: false)
  }

  private func verifyConnectionHelpAndKeyboard(formColumnDrag: Bool) {
    let address = app.textFields["Server address"]
    XCTAssertTrue(address.waitForExistence(timeout: 30))
    XCTAssertTrue(address.isHittable)
    capture("onboarding-entry")

    let help = app.buttons["Need help connecting?"]
    XCTAssertTrue(help.isHittable)
    help.tap()
    let helpText = app.staticTexts.containing(NSPredicate(format: "label CONTAINS %@", "Enter your Stuff Stash server’s full address")).firstMatch
    XCTAssertTrue(helpText.waitForExistence(timeout: 5))
    capture("onboarding-help")
    help.tap()
    XCTAssertTrue(helpText.waitForNonExistence(timeout: 5), "Connection help must close")

    address.tap()
    waitForKeyboard()
    address.typeText("https://example.invalid")
    XCTAssertTrue(app.keyboards.firstMatch.waitForExistence(timeout: 5))
    let completeAddress = XCTNSPredicateExpectation(
      predicate: NSPredicate(format: "value == %@", "https://example.invalid"), object: address)
    XCTAssertEqual(XCTWaiter.wait(for: [completeAddress], timeout: 5), .completed,
      "Address entry must finish with the complete value within five seconds")
    capture("onboarding-keyboard")
    XCTAssertEqual(address.value as? String, "https://example.invalid", "Typing must preserve the complete server address")
    // Interactive keyboard dismissal follows a downward drag from scroll content.
    let origin = app.coordinate(withNormalizedOffset: CGVector(dx: 0, dy: 0))
    let scroll = app.scrollViews.firstMatch
    XCTAssertTrue(scroll.exists)
    // Resolve each frame once: repeated accessibility snapshots can stall on CI.
    let scrollBounds = scroll.frame
    let keyboardBounds = app.keyboards.firstMatch.frame
    let applicationBounds = app.frame
    let startPoint = CGPoint(x: formColumnDrag ? address.frame.midX : scrollBounds.minX + 8, y: min(scrollBounds.maxY, keyboardBounds.minY) - 80)
    XCTAssertTrue(scrollBounds.contains(startPoint), "Dismissal drag must begin inside the scroll surface")
    if formColumnDrag {
      XCTAssertGreaterThan(startPoint.y, app.buttons["Connect and sign in"].frame.maxY, "Comparison drag must begin in blank content, not on a command")
    }
    let start = origin.withOffset(CGVector(dx: startPoint.x, dy: startPoint.y))
    let end = origin.withOffset(CGVector(dx: startPoint.x, dy: min(applicationBounds.maxY - 24, startPoint.y + 300)))
    start.press(forDuration: 0.1, thenDragTo: end)
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: app.keyboards.firstMatch)], timeout: 5), .completed)
    let connect = app.buttons["Connect and sign in"]
    XCTAssertTrue(connect.isHittable, "Connection action must remain reachable after dismissing the keyboard by scrolling")
    help.tap()
    XCTAssertTrue(helpText.waitForExistence(timeout: 5))
    XCTAssertEqual(address.value as? String, "https://example.invalid", "Opening help must preserve the native draft")
    help.tap()
    XCTAssertEqual(address.value as? String, "https://example.invalid", "Closing help must preserve the native draft")
    // Do not contact an external instance or open an identity-provider session.
    capture("onboarding-action-reachable")
  }
}
