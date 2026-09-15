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
      keyboard.keys.allElementsBoundByIndex.contains { $0.isHittable }
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
    let diagnostics = app.staticTexts.matching(NSPredicate(format: "label == %@", "Audit query readiness"))
    for (index, element) in diagnostics.allElementsBoundByIndex.enumerated() {
      guard let value = element.value as? String, value.hasPrefix("{") else { continue }
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

  func testCheckoutHistoryRemainsReadableAndDismissibleAfterExpansion() {
    let open = app.buttons["Audit Checkout history"]
    for _ in 0..<7 where !open.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(open.isHittable)
    open.tap()
    let bar = app.navigationBars["Checkout history"]
    XCTAssertTrue(bar.waitForExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["Audit ladder"].waitForExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["Audit checkout 1: borrowed for cleaning the gutters."].waitForExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["Audit checkout 1: borrowed for cleaning the gutters."].isHittable)
    XCTAssertTrue(app.buttons["Close"].isHittable)
    capture("checkout-history-medium")
    let initialTop = bar.frame.minY
    bar.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
      .press(forDuration: 0.1, thenDragTo: app.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.12)))
    if UIDevice.current.userInterfaceIdiom == .phone {
      let expanded = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in bar.frame.minY < initialTop - 40 }, object: nil)
      XCTAssertEqual(XCTWaiter.wait(for: [expanded], timeout: 5), .completed)
    }
    XCTAssertTrue(app.staticTexts["Audit checkout 1: borrowed for cleaning the gutters."].isHittable)
    capture("checkout-history-expanded")
    let older = app.buttons["Load older checkouts"]
    for _ in 0..<6 where !older.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(older.isHittable)
    older.tap()
    let loaded = app.staticTexts["Older audit checkout"]
    XCTAssertTrue(loaded.waitForExistence(timeout: 5))
    for _ in 0..<4 where !loaded.isHittable { app.scrollViews.firstMatch.swipeUp() }
    XCTAssertTrue(loaded.isHittable)
    capture("checkout-history-older-page")
    app.buttons["Close"].tap()
    XCTAssertTrue(open.waitForExistence(timeout: 5))
    XCTAssertFalse(bar.exists)
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
    choice.tap()
    XCTAssertTrue(app.buttons["Available"].waitForExistence(timeout: 5))
    capture("choice-menu-accessibility-size")
  }

  func testExpirationOverviewAccessibility() throws {
    app.buttons["Audit Expiration filters"].tap()
    XCTAssertTrue(app.buttons["Choose tags"].waitForExistence(timeout: 5))
    XCTAssertTrue(app.buttons["Apply expiration filters"].isHittable)
    capture("expiration-overview-accessibility")
    if #available(iOS 17.0, *) {
      try app.performAccessibilityAudit(for: [.hitRegion, .sufficientElementDescription, .trait, .dynamicType, .textClipped])
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
    XCTAssertTrue(app.staticTexts["Observed \(mode) input: https://example.invalid"].waitForExistence(timeout: 5))
  }

  func testControlledAddressEntry() { verifyAddressEntry("controlled") }
  func testUncontrolledAddressEntry() { verifyAddressEntry("uncontrolled") }
  func testSystemAddressEntry() { verifyAddressEntry("system") }

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
    capture("home-return-save-error-retained")
    XCTAssertTrue(save.isHittable)
    save.tap()
    XCTAssertEqual(XCTWaiter.wait(for: [XCTNSPredicateExpectation(predicate: NSPredicate(format: "exists == false"), object: app.navigationBars["Return details"])], timeout: 5), .completed)
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
      try app.performAccessibilityAudit(for: [.hitRegion, .sufficientElementDescription, .trait, .contrast, .dynamicType, .textClipped])
    } else {
      throw XCTSkip("XCTest accessibility audit requires iOS 17")
    }
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
    picker.tap()
    let sliders = app.buttons["Sliders"]
    XCTAssertTrue(sliders.waitForExistence(timeout: 5), "The system color picker should open directly")
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

  func testColorWellTargetOpensSystemPicker() {
    openSettingsControls()
    let picker = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", "Choose any color")).firstMatch
    XCTAssertTrue(picker.waitForExistence(timeout: 5))
    XCTAssertTrue(picker.isHittable)
    XCTAssertGreaterThan(picker.frame.width, picker.frame.height)
    // The captured LTR system control places its circular well at the row's trailing edge.
    let well = picker.coordinate(withNormalizedOffset: CGVector(dx: 1, dy: 0.5))
      .withOffset(CGVector(dx: -picker.frame.height / 2, dy: 0))
    well.tap()
    XCTAssertTrue(app.buttons["Sliders"].waitForExistence(timeout: 5))
    capture("color-visible-well-target")
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
