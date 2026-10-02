import ExpoModulesCore
import Foundation

public final class StuffStashArchiveTransferModule: Module {
  private let lock = NSLock()
  private var closed = false
  private var uploads: [String: ArchiveUpload] = [:]

  public func definition() -> ModuleDefinition {
    Name("StuffStashArchiveTransfer")
    AsyncFunction("upload") { (id: String, address: String, fileURI: String, headers: [String: String], promise: Promise) in
      guard let url = URL(string: address), ["https", "http"].contains(url.scheme ?? ""),
            let file = URL(string: fileURI), file.isFileURL else {
        promise.reject("ERR_ARCHIVE_REQUEST", "Invalid archive transfer request.")
        return
      }
      self.lock.lock()
      guard !self.closed && self.uploads.isEmpty else {
        self.lock.unlock()
        promise.reject("ERR_ARCHIVE_REQUEST", "Archive transfer is already running.")
        return
      }
      let upload = ArchiveUpload(promise: promise) { [weak self] in
        guard let self else { return }
        self.lock.lock()
        self.uploads.removeValue(forKey: id)
        self.lock.unlock()
      }
      self.uploads[id] = upload
      // Install and start while holding the lock, so cancellation cannot miss the task.
      upload.start(url: url, file: file, headers: headers)
      self.lock.unlock()
    }.runOnQueue(.main)
    AsyncFunction("cancel") { (id: String) in
      self.lock.lock()
      let upload = self.uploads[id]
      self.lock.unlock()
      upload?.cancel()
    }.runOnQueue(.main)
    OnDestroy {
      self.lock.lock()
      self.closed = true
      let active = Array(self.uploads.values)
      self.lock.unlock()
      for upload in active { upload.cancel() }
    }
  }
}

private final class ArchiveUpload: NSObject, URLSessionDataDelegate, URLSessionTaskDelegate {
  private let promise: Promise
  private let finished: () -> Void
  private var session: URLSession?
  private var task: URLSessionUploadTask?
  private var response: HTTPURLResponse?
  private var body = Data()
  private var exceededLimit = false
  private let responseLimit = 64 * 1024

  init(promise: Promise, finished: @escaping () -> Void) {
    self.promise = promise
    self.finished = finished
  }
  func start(url: URL, file: URL, headers: [String: String]) {
    let config = URLSessionConfiguration.ephemeral
    config.httpShouldSetCookies = false
    config.urlCache = nil
    config.timeoutIntervalForRequest = 30 * 60
    config.timeoutIntervalForResource = 30 * 60
    let session = URLSession(configuration: config, delegate: self, delegateQueue: nil)
    self.session = session
    var request = URLRequest(url: url)
    request.httpMethod = "POST"
    request.allHTTPHeaderFields = headers
    request.cachePolicy = .reloadIgnoringLocalCacheData
    request.httpShouldHandleCookies = false
    let task = session.uploadTask(with: request, fromFile: file)
    self.task = task
    task.resume()
  }
  func cancel() { task?.cancel() }
  func urlSession(_ session: URLSession, task: URLSessionTask, willPerformHTTPRedirection response: HTTPURLResponse, newRequest request: URLRequest, completionHandler: @escaping (URLRequest?) -> Void) {
    // Do not forward a bearer token or archive to any redirect target.
    completionHandler(nil)
  }
  func urlSession(_ session: URLSession, dataTask: URLSessionDataTask, didReceive response: URLResponse, completionHandler: @escaping (URLSession.ResponseDisposition) -> Void) {
    self.response = response as? HTTPURLResponse
    if response.expectedContentLength > Int64(responseLimit) {
      exceededLimit = true
      completionHandler(.cancel)
    } else { completionHandler(.allow) }
  }
  func urlSession(_ session: URLSession, dataTask: URLSessionDataTask, didReceive data: Data) {
    guard data.count <= responseLimit - body.count else {
      exceededLimit = true
      dataTask.cancel()
      return
    }
    body.append(data)
  }
  func urlSession(_ session: URLSession, task: URLSessionTask, didCompleteWithError error: Error?) {
    session.finishTasksAndInvalidate()
    self.session = nil
    finished()
    if exceededLimit { promise.reject("ERR_ARCHIVE_RESPONSE", "Archive response exceeded its size limit."); return }
    if error != nil { promise.reject("ERR_ARCHIVE_TRANSFER", "Archive upload failed."); return }
    guard let response else { promise.reject("ERR_ARCHIVE_RESPONSE", "Missing archive response."); return }
    promise.resolve(["status": response.statusCode, "body": String(data: body, encoding: .utf8) ?? ""])
  }
}
