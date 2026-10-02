import XCTest

/// Production routes and repositories; no fixture installation or injected session.
final class ConnectedAuditTests: XCTestCase {
  private let app = XCUIApplication(bundleIdentifier: "org.stuffstash.mobile")
  private var stage = "launch"

  override func setUpWithError() throws {
    continueAfterFailure = false
    app.launch()
  }

  override func tearDownWithError() throws {
    let evidence = XCTAttachment(string: stage)
    evidence.name = "connected-stage"
    evidence.lifetime = .keepAlways
    add(evidence)
    app.terminate()
  }

  private func button(_ name: String) -> XCUIElement { app.buttons[name].firstMatch }
  private func tap(_ name: String) {
    let control = button(name)
    XCTAssertTrue(control.waitForExistence(timeout: 20), "Expected workflow control")
    XCTAssertTrue(control.isHittable, "Workflow control must be reachable")
    control.tap()
  }
  private func enter(_ field: XCUIElement, _ value: String) {
    XCTAssertTrue(field.waitForExistence(timeout: 20), "Expected input")
    field.tap()
    let keyboard = app.keyboards.firstMatch
    XCTAssertTrue(keyboard.waitForExistence(timeout: 5), "Expected keyboard")
    field.typeText(value)
    let entered = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      field.value as? String == value
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [entered], timeout: 10), .completed, "Full text must settle before submission")
  }
  private func capture(_ name: String) {
    let evidence = XCTAttachment(screenshot: app.screenshot())
    evidence.name = "connected-safe-" + name
    evidence.lifetime = .keepAlways
    add(evidence)
  }

  private func signIn(_ email: String, first: Bool) {
    stage = first ? "owner-sign-in" : "other-sign-in"
    if first {
      enter(app.textFields["Server address"].firstMatch, "http://localhost:8080")
      tap("Connect and sign in")
    } else {
      tap("Connect and sign in")
    }
    // ASWebAuthenticationSession may ask to continue before showing the provider.
    let springboard = XCUIApplication(bundleIdentifier: "com.apple.springboard")
    let permission = springboard.alerts.buttons["Continue"].firstMatch
    if permission.waitForExistence(timeout: 3) { permission.tap() }
    let localPermission = app.alerts.buttons["Continue"].firstMatch
    if localPermission.exists { localPermission.tap() }
    let safari = XCUIApplication(bundleIdentifier: "com.apple.SafariViewService")
    let inAppWeb = app.webViews.firstMatch
    let web = inAppWeb.waitForExistence(timeout: 5) ? inAppWeb : safari.webViews.firstMatch
    XCTAssertTrue(web.waitForExistence(timeout: 20), "System authentication browser must open")
    let localLogin = web.links.matching(NSPredicate(format: "label CONTAINS[c] 'email'")).firstMatch
    if localLogin.waitForExistence(timeout: 2) { localLogin.tap() }
    let login = web.textFields.firstMatch
    XCTAssertTrue(login.waitForExistence(timeout: 15), "Provider login form")
    login.tap(); login.typeText(email)
    let completeEmail = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
      login.value as? String == email
    }, object: nil)
    XCTAssertEqual(XCTWaiter.wait(for: [completeEmail], timeout: 10), .completed, "Provider email must settle")
    let password = web.secureTextFields.firstMatch
    XCTAssertTrue(password.waitForExistence(timeout: 5))
    password.tap(); password.typeText("password")
    let submit = web.buttons.matching(NSPredicate(format: "label MATCHES[c] '(sign in|log in)'")).firstMatch
    XCTAssertTrue(submit.waitForExistence(timeout: 5)); submit.tap()
    XCTAssertTrue(button("Create household").waitForExistence(timeout: 30), "Real OIDC callback must reach setup")
  }

  private func response(_ path: String, token: String? = nil, form: String? = nil) throws -> (Int, Any) {
    let base = form == nil ? "http://localhost:8080" : "http://localhost:5556"
    var request = URLRequest(url: URL(string: base + path)!)
    request.timeoutInterval = 15
    if let token { request.setValue("Bearer " + token, forHTTPHeaderField: "Authorization") }
    if let form {
      request.httpMethod = "POST"
      request.httpBody = form.data(using: .utf8)
      request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")
    }
    let done = expectation(description: "API response")
    var status = 0
    var body: Any = NSNull()
    URLSession.shared.dataTask(with: request) { data, response, _ in
      status = (response as? HTTPURLResponse)?.statusCode ?? 0
      if let data { body = (try? JSONSerialization.jsonObject(with: data)) ?? NSNull() }
      done.fulfill()
    }.resume()
    wait(for: [done], timeout: 20)
    XCTAssertNotEqual(status, 0, "API request must complete")
    return (status, body)
  }
  private func token(_ email: String) throws -> String {
    let (status, body) = try response("/dex/token", form: "grant_type=password&client_id=stuff-stash-local&client_secret=stuff-stash-local-secret&scope=openid%20email%20profile&username=\(email)&password=password")
    XCTAssertEqual(status, 200, "Independent audit principal authentication")
    guard let value = (body as? [String: Any])?["id_token"] as? String else {
      XCTFail("Identity token missing"); return ""
    }
    return value
  }
  private func list(_ path: String, token: String) throws -> [[String: Any]] {
    let (status, body) = try response(path, token: token)
    XCTAssertEqual(status, 200, "Authorized collection must load")
    guard let items = (body as? [String: Any])?["data"] as? [[String: Any]] else {
      XCTFail("Collection envelope missing"); return []
    }
    return items
  }

  func testRealSignInPersistenceAndPrincipalIsolation() throws {
    signIn("owner@example.com", first: true)
    stage = "owner-household"
    enter(app.textFields["Household name"].firstMatch, "Native Audit Home")
    tap("Create household")
    XCTAssertTrue(button("Add an asset").waitForExistence(timeout: 30))
    stage = "create-item"
    tap("Add an asset")
    enter(app.textFields["Asset name"].firstMatch, "Connected native lamp")
    tap("Save item")
    XCTAssertTrue(app.staticTexts["Asset saved"].firstMatch.waitForExistence(timeout: 20))
    tap("Close Add")
    stage = "relaunch"
    app.terminate(); app.launch()
    let item = app.staticTexts["Connected native lamp"].firstMatch
    XCTAssertTrue(item.waitForExistence(timeout: 30), "Saved item must survive relaunch")
    item.tap()
    XCTAssertTrue(app.navigationBars["Details"].waitForExistence(timeout: 15))
    capture("persisted-detail")

    stage = "server-persistence"
    let owner = try token("owner@example.com")
    let tenants = try list("/me/tenants", token: owner)
    let tenant = try XCTUnwrap(tenants.first?["id"] as? String)
    let inventories = try list("/tenants/\(tenant)/inventories", token: owner)
    let inventory = try XCTUnwrap(inventories.first?["id"] as? String)
    let path = "/tenants/\(tenant)/inventories/\(inventory)"
    let items = try list(path + "/assets", token: owner)
    XCTAssertTrue(items.contains { $0["title"] as? String == "Connected native lamp" }, "API must retain the UI-created item")

    stage = "sign-out"
    tap("Home")
    tap("Open account and settings")
    tap("Open Account settings")
    tap("Sign Out")
    let confirm = app.alerts.buttons["Sign Out"].firstMatch
    XCTAssertTrue(confirm.waitForExistence(timeout: 5)); confirm.tap()
    signIn("viewer@example.com", first: false)
    stage = "other-account-isolation"
    XCTAssertFalse(app.staticTexts["Native Audit Home"].exists)
    XCTAssertFalse(app.staticTexts["Connected native lamp"].exists)
    let other = try token("viewer@example.com")
    XCTAssertTrue(try list("/me/tenants", token: other).isEmpty, "Other principal must have no owner membership")
    let (denied, _) = try response(path, token: other)
    XCTAssertTrue([403, 404].contains(denied), "Other principal must not read the owner inventory")
    let (anonymous, _) = try response(path)
    XCTAssertEqual(anonymous, 401)
    capture("other-account-setup")
    stage = "passed"
  }
}
