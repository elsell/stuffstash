import XCTest

final class CompletePrintingAuditTests: XCTestCase {
  private let app = XCUIApplication(bundleIdentifier: "org.stuffstash.mobile")
  override func setUpWithError() throws { continueAfterFailure = false }
  override func tearDownWithError() throws { capture("complete-print-final"); app.terminate() }
  private func capture(_ name: String) {
    let image = XCTAttachment(screenshot: app.screenshot()); image.name = name; image.lifetime = .keepAlways; add(image)
    let tree = XCTAttachment(string: app.debugDescription); tree.name = name + "-hierarchy"; tree.lifetime = .keepAlways; add(tree)
  }
  private func reveal(_ element: XCUIElement) {
    XCTAssertTrue(element.waitForExistence(timeout: 10))
    for _ in 0..<10 where !element.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(element.isHittable)
    XCTAssertGreaterThanOrEqual(element.frame.minY, app.navigationBars.firstMatch.frame.maxY)
    XCTAssertLessThanOrEqual(element.frame.maxY, app.frame.maxY)
  }
  private func action(_ label: String) { let button = app.buttons[label].firstMatch; reveal(button); button.tap() }
  private func see(_ label: String) { XCTAssertTrue(app.staticTexts[label].firstMatch.waitForExistence(timeout: 10)) }
  func testCompletePrintFlow() { audit(large: false) }
  func testCompletePrintFlowAtAccessibilityTextSize() { audit(large: true) }
  private func audit(large: Bool) {
    app.launchArguments = large ? ["-UIPreferredContentSizeCategoryName", "UICTContentSizeCategoryAccessibilityXXXL"] : []
    app.launch(); let suffix = large ? "large" : "normal"
    action("Open quick print")
    let copies = app.textFields["Copies"].firstMatch
    reveal(copies)
    capture("complete-print-options-\(suffix)")
    copies.tap(); XCTAssertTrue(app.keyboards.firstMatch.waitForExistence(timeout: 5))
    copies.typeText(XCUIKeyboardKey.delete.rawValue + "2")
    let complete = XCTNSPredicateExpectation(predicate: NSPredicate(format: "value == %@", "2"), object: copies)
    XCTAssertEqual(XCTWaiter.wait(for: [complete], timeout: 5), .completed)
    let dismiss = app.buttons["Dismiss keyboard"].firstMatch
    XCTAssertTrue(dismiss.isHittable); dismiss.tap()
    XCTAssertTrue(app.keyboards.firstMatch.waitForNonExistence(timeout: 5))
    action("Preview label")
    see("Printing is unavailable. Check your connection and inventory access, then try again.")
    capture("complete-print-preview-failure-\(suffix)")
    action("Preview label")
    let preview = app.descendants(matching: .any).matching(identifier: "Label preview").firstMatch
    XCTAssertTrue(preview.waitForExistence(timeout: 15), "Success must fetch and decode real PNG bytes through ExpoLabelFiles")
    reveal(preview); XCTAssertGreaterThan(preview.frame.width, 0); XCTAssertGreaterThan(preview.frame.height, 0)
    capture("complete-print-native-png-\(suffix)")
    action("Print label"); see("Queued"); see("0 of 2 copies confirmed")
    capture("complete-print-queued-\(suffix)")
    action("Cancel print job"); see("Canceled")
    action("Reprint label")
    let cancel = app.navigationBars.firstMatch.buttons["Cancel"].firstMatch
    XCTAssertTrue(cancel.waitForExistence(timeout: 10)); XCTAssertTrue(cancel.isHittable); cancel.tap()
    see("Canceled")
    action("Reprint label"); action("Preview label")
    XCTAssertTrue(preview.waitForExistence(timeout: 15)); action("Reprint label")
    see("Queued"); see("0 of 1 copies confirmed")
    capture("complete-print-reprint-\(suffix)")
    for _ in 0..<4 where !app.buttons["Inspect uncertain job"].isHittable { app.navigationBars.firstMatch.buttons.firstMatch.tap() }
    action("Inspect uncertain job"); see("Output uncertain")
    capture("complete-print-uncertain-\(suffix)")
    let outcome = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "What happened at the printer?")).firstMatch
    reveal(outcome); outcome.tap(); app.buttons["No label printed"].firstMatch.tap()
    let acknowledgement = app.switches["I understand the print outcome remains unconfirmed"].firstMatch
    reveal(acknowledgement); acknowledgement.tap()
    action("Resolve job"); see("Uncertainty acknowledged"); see("No label printed")
    let reprint = app.buttons["Reprint label"].firstMatch; reveal(reprint)
    capture("complete-print-resolved-\(suffix)")
  }
}
