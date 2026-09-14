import XCTest
import UIKit

final class FixtureAuditTests: XCTestCase {
  private let app = XCUIApplication(bundleIdentifier: "org.stuffstash.mobile")
  override func setUpWithError() throws {
    continueAfterFailure = false
    app.launch()
    XCTAssertTrue(app.buttons["Audit Browse filters"].waitForExistence(timeout: 30))
  }
  override func tearDownWithError() throws {
    capture("final-state")
    app.terminate()
  }
  private func waitForKeyboard() {
    let keyboard = app.keyboards.firstMatch
    XCTAssertTrue(keyboard.waitForExistence(timeout: 5))
    let ready = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      keyboard.keys.firstMatch.exists && keyboard.keys.firstMatch.isHittable
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
    dismiss.tap()
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: dismiss)], timeout: 5), .completed)
    capture("expiration-calendar-dismissed")
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
    let search = app.searchFields.firstMatch
    XCTAssertTrue(search.waitForExistence(timeout: 5))
    search.tap()
    waitForKeyboard()
    search.typeText("Tools")
    XCTAssertTrue(app.keyboards.firstMatch.waitForExistence(timeout: 5))
    capture("expiration-search-keyboard")
    XCTAssertTrue(app.buttons["Apply expiration filters"].isHittable)
    let back = app.buttons["Cancel or return to filters"]
    XCTAssertTrue(back.isHittable)
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

  private func verifyAddressEntry(_ mode: String) {
    app.buttons["Audit \(mode) input"].tap()
    let input = app.textFields["Audit \(mode) address"]
    XCTAssertTrue(input.waitForExistence(timeout: 5))
    if !input.isHittable { app.scrollViews.firstMatch.swipeUp() }
    input.tap()
    waitForKeyboard()
    input.typeText("https://example.invalid")
    capture("\(mode)-address-entry")
    XCTAssertEqual(input.value as? String, "https://example.invalid")
  }

  func testControlledAddressEntry() { verifyAddressEntry("controlled") }
  func testUncontrolledAddressEntry() { verifyAddressEntry("uncontrolled") }

  func testAddDraftRetainsTextAndRecoversAfterRejectedSave() {
    let open = app.buttons["Audit Add draft"]
    for _ in 0..<4 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let name = app.textFields["Asset name"]
    XCTAssertTrue(name.waitForExistence(timeout: 10))
    XCTAssertTrue(app.navigationBars["Add item"].exists)
    capture("add-native-header")
    name.tap()
    waitForKeyboard()
    name.typeText("Native draft name")
    XCTAssertEqual(name.value as? String, "Native draft name")
    let save = app.buttons["Save item"]
    let close = app.buttons["Close Add"]
    XCTAssertTrue(save.isHittable)
    save.tap()
    XCTAssertFalse(save.isEnabled)
    XCTAssertFalse(close.isEnabled)
    let rejected = app.staticTexts["Rejected draft: Native draft name"]
    XCTAssertTrue(rejected.waitForExistence(timeout: 10))
    XCTAssertEqual(name.value as? String, "Native draft name")
    XCTAssertTrue(save.isEnabled)
    XCTAssertTrue(close.isEnabled)
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
    let connect = app.buttons["Connect and sign in"]
    for _ in 0..<3 where !connect.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(connect.isHittable)
    connect.tap()
    XCTAssertTrue(app.staticTexts["Submitted address: https://example.invalid"].waitForExistence(timeout: 5))
    capture("onboarding-complete-address-submission")
  }

  private func openSettingsControls() {
    let button = app.buttons["Audit settings controls"]
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

  func testColorPickerOpensDirectlyAndClearPreservesParentDraft() {
    openSettingsControls()
    XCTAssertTrue(app.staticTexts["Color value: none"].exists)
    XCTAssertFalse(app.buttons["Choose a custom tag color"].exists)
    let picker = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose any color")).firstMatch
    XCTAssertTrue(picker.waitForExistence(timeout: 5))
    picker.tap()
    XCTAssertTrue(app.buttons["Close"].waitForExistence(timeout: 5), "The system color picker should open directly")
    capture("native-color-picker")
    app.buttons["Close"].tap()
    XCTAssertTrue(app.staticTexts["Color value: none"].exists, "Opening and closing must not invent a color")
    app.buttons["Choose Green tag color"].tap()
    XCTAssertTrue(app.staticTexts["Color value: #2E7D32"].exists)
    app.buttons["No tag color"].tap()
    XCTAssertTrue(app.staticTexts["Color value: none"].exists)
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
}
