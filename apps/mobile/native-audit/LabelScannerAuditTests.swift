import XCTest
import UIKit

final class LabelScannerAuditTests: XCTestCase {
  private let app = XCUIApplication(bundleIdentifier: "org.stuffstash.mobile")
  private let instance = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
  private func link(_ label: String, instance: String? = nil) -> String {
    "https://old.example/l/v1/\(instance ?? self.instance)/\(label)"
  }
  override func setUpWithError() throws {
    continueAfterFailure = false
    if name.contains("testDeniedCameraPasteRejectsInvalidAndForeignLinks") { app.resetAuthorizationStatus(for: .camera) }
    app.launch()
    XCTAssertTrue(app.staticTexts["Scanner acceptance home"].waitForExistence(timeout: 30))
  }
  override func tearDownWithError() throws {
    capture("scanner-final")
    app.terminate()
  }
  private func capture(_ name: String) {
    let screenshot = XCTAttachment(screenshot: app.screenshot())
    screenshot.name = name; screenshot.lifetime = .keepAlways; add(screenshot)
    let hierarchy = XCTAttachment(string: app.debugDescription)
    hierarchy.name = name + "-hierarchy"; hierarchy.lifetime = .keepAlways; add(hierarchy)
  }
  private func openScanner(requireDenial: Bool = false) {
    app.buttons["Open scanner"].tap()
    let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")
    let permission = springboard.alerts.firstMatch
    let prompted = permission.waitForExistence(timeout: requireDenial ? 10 : 3)
    if requireDenial { XCTAssertTrue(prompted, "Permission denial requires the real OS prompt; camera mount failure is not equivalent") }
    if prompted {
      let deny = permission.buttons.matching(NSPredicate(format: "label CONTAINS[c] %@", "Allow")).allElementsBoundByIndex.first {
        $0.label.lowercased().contains("don")
      }
      XCTAssertNotNil(deny, "Expected the actual OS camera permission denial action")
      deny?.tap()
    }
    XCTAssertTrue(app.staticTexts["Camera unavailable. Allow camera access in Settings, or paste a label link."].waitForExistence(timeout: 10))
    XCTAssertTrue(app.navigationBars.buttons["Cancel"].isHittable)
  }
  private func paste(_ source: String) {
    let field = app.textFields["Paste label link"]
    XCTAssertTrue(field.waitForExistence(timeout: 10))
    UIPasteboard.general.string = source
    field.press(forDuration: 1.2)
    let paste = app.menuItems["Paste"].firstMatch
    let pasteButton = app.buttons["Paste"].firstMatch
    let available = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      paste.exists || pasteButton.exists
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [available], timeout: 5), .completed, "Use the real native Paste action")
    if paste.exists { paste.tap() } else { pasteButton.tap() }
    let entered = NSPredicate(format: "value == %@", source)
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: entered, object: field)], timeout: 5), .completed, "Pasting must retain the complete label link")
    let open = app.buttons["Open label"].firstMatch
    if !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable); open.tap()
  }
  private func cancel() {
    app.navigationBars.buttons["Cancel"].tap()
    XCTAssertTrue(app.staticTexts["Scanner acceptance home"].waitForExistence(timeout: 10))
  }
  func testDeniedCameraPasteRejectsInvalidAndForeignLinks() {
    openScanner(requireDenial: true); capture("scanner-camera-denied-paste")
    paste("https://unrelated.example/not-a-label")
    XCTAssertTrue(app.staticTexts["This is not a supported Stuff Stash label."].waitForExistence(timeout: 10))
    capture("scanner-invalid-link"); cancel()
    openScanner()
    paste(link("01ARZ3NDEKTSV4RRFFQ69G5FAW", instance: "01ARZ3NDEKTSV4RRFFQ69G5FAZ"))
    XCTAssertTrue(app.staticTexts["This label belongs to another Stuff Stash instance. Connect to that server and try again."].waitForExistence(timeout: 10))
    capture("scanner-foreign-instance"); cancel()
  }
  func testOldHostLookupRetryNavigatesToResolvedAsset() {
    openScanner(); paste(link("01ARZ3NDEKTSV4RRFFQ69G5FAX"))
    XCTAssertTrue(app.staticTexts["Label unavailable. Check your connection and access, then try again."].waitForExistence(timeout: 10))
    capture("scanner-recoverable-failure")
    app.buttons["Open label"].firstMatch.tap()
    XCTAssertTrue(app.staticTexts["Cordless drill resolved"].waitForExistence(timeout: 10))
    XCTAssertEqual(app.staticTexts["scanner-selected"].label, "Selected scope count: 1")
    XCTAssertFalse(app.navigationBars.buttons["Cancel"].exists)
    capture("scanner-resolved-asset")
  }
  func testCancelPreventsLateLookupNavigation() {
    openScanner(); paste(link("01ARZ3NDEKTSV4RRFFQ69G5FAY"))
    XCTAssertTrue(app.staticTexts["Opening label…"].waitForExistence(timeout: 10))
    cancel(); app.buttons["Complete delayed lookup"].tap()
    let evidence = app.staticTexts["scanner-evidence"]
    let settled = NSPredicate { _, _ in
      evidence.exists && evidence.label.contains("\"lateReplies\":1") && evidence.label.contains("\"cancelled\":true") && evidence.label.contains("\"selected\":0")
    }
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: settled, object: evidence)], timeout: 10), .completed)
    XCTAssertTrue(app.staticTexts["Scanner acceptance home"].exists)
    XCTAssertTrue(app.staticTexts["No pending label"].exists)
    XCTAssertFalse(app.staticTexts["Cordless drill resolved"].exists)
    capture("scanner-cancelled-late-reply")
  }
  func testPendingLabelSurvivesSimulatedSignInReadiness() {
    app.buttons["Capture label before simulated sign-in"].tap()
    XCTAssertTrue(app.staticTexts["Simulated sign-in required"].waitForExistence(timeout: 10))
    XCTAssertTrue(app.staticTexts["Pending label retained"].exists)
    app.buttons["Complete simulated sign-in"].tap()
    let offer = app.buttons["Open label"].firstMatch
    XCTAssertTrue(offer.waitForExistence(timeout: 10)); offer.tap()
    XCTAssertTrue(app.navigationBars.buttons["Cancel"].waitForExistence(timeout: 10))
    app.buttons["Open label"].firstMatch.tap()
    XCTAssertTrue(app.staticTexts["Cordless drill resolved"].waitForExistence(timeout: 10))
    capture("scanner-simulated-sign-in-return")
  }
}
