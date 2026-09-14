import XCTest

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

  private func capture(_ name: String) {
    let attachment = XCTAttachment(screenshot: app.screenshot())
    attachment.name = name
    attachment.lifetime = .keepAlways
    add(attachment)
  }

  func testConnectionHelpAndKeyboardKeepActionsReachable() throws {
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
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: helpText)], timeout: 5), .completed)

    address.tap()
    address.typeText("https://example.invalid")
    XCTAssertTrue(app.keyboards.firstMatch.waitForExistence(timeout: 5))
    capture("onboarding-keyboard")
    XCTAssertEqual(address.value as? String, "https://example.invalid", "Typing must preserve the complete server address")
    // Interactive keyboard dismissal follows a downward drag from scroll content.
    let origin = app.coordinate(withNormalizedOffset: CGVector(dx: 0, dy: 0))
    let start = origin.withOffset(CGVector(dx: app.frame.midX, dy: app.keyboards.firstMatch.frame.minY - 24))
    let end = origin.withOffset(CGVector(dx: app.frame.midX, dy: app.frame.maxY - 24))
    start.press(forDuration: 0.1, thenDragTo: end)
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: app.keyboards.firstMatch)], timeout: 5), .completed)
    let connect = app.buttons["Connect and sign in"]
    XCTAssertTrue(connect.isHittable, "Connection action must remain reachable after dismissing the keyboard by scrolling")
    // Do not contact an external instance or open an identity-provider session.
    capture("onboarding-action-reachable")
  }
}
