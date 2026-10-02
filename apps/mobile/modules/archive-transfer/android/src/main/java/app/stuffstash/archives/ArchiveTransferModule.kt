package app.stuffstash.archives

import android.net.Uri
import expo.modules.kotlin.Promise
import expo.modules.kotlin.functions.Queues
import expo.modules.kotlin.modules.Module
import expo.modules.kotlin.modules.ModuleDefinition
import java.io.IOException
import java.util.concurrent.TimeUnit
import okhttp3.Call
import okhttp3.Callback
import okhttp3.CookieJar
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody
import okhttp3.Response
import okio.BufferedSink

class ArchiveTransferModule : Module() {
  private val client = OkHttpClient.Builder()
    .followRedirects(false)
    .followSslRedirects(false)
    .cookieJar(CookieJar.NO_COOKIES)
    .callTimeout(30, TimeUnit.MINUTES)
    .build()
  private var closed = false
  private val uploads = mutableMapOf<String, Call>()
  private val responseLimit = 64L * 1024L

  override fun definition() = ModuleDefinition {
    Name("StuffStashArchiveTransfer")
    AsyncFunction("upload") { id: String, address: String, fileURI: String, headers: Map<String, String>, promise: Promise ->
      val resolver = appContext.reactContext?.contentResolver
      val uri = Uri.parse(fileURI)
      if (resolver == null || uri.scheme !in listOf("file", "content")) {
        promise.reject("ERR_ARCHIVE_REQUEST", "Invalid archive file.", null)
        return@AsyncFunction
      }
      try {
        val body = object : RequestBody() {
          override fun contentType() = "application/zip".toMediaType()
          override fun writeTo(sink: BufferedSink) {
            val input = resolver.openInputStream(uri) ?: throw IOException("Archive file unavailable.")
            input.use { it.copyTo(sink.outputStream(), 64 * 1024) }
          }
        }
        val builder = Request.Builder().url(address).post(body)
        headers.forEach { (name, value) -> builder.header(name, value) }
        val call = client.newCall(builder.build())
        synchronized(uploads) {
          if (closed || uploads.isNotEmpty()) {
            promise.reject("ERR_ARCHIVE_REQUEST", "Archive transfer is already running.", null)
            return@AsyncFunction
          }
          uploads[id] = call
        }
        call.enqueue(object : Callback {
          override fun onFailure(call: Call, error: IOException) {
            synchronized(uploads) { uploads.remove(id) }
            promise.reject("ERR_ARCHIVE_TRANSFER", "Archive upload failed.", error)
          }
          override fun onResponse(call: Call, response: Response) {
            try {
              val result = response.use {
                val source = it.body?.source()
                if (source != null && source.request(responseLimit + 1)) throw IOException("Archive response exceeded its size limit.")
                val text = source?.buffer?.readUtf8() ?: ""
                mapOf("status" to it.code, "body" to text)
              }
              synchronized(uploads) { uploads.remove(id) }
              promise.resolve(result)
            } catch (error: Exception) {
              synchronized(uploads) { uploads.remove(id) }
              promise.reject("ERR_ARCHIVE_RESPONSE", "Invalid archive response.", error)
            }
          }
        })
      } catch (error: Exception) {
        synchronized(uploads) { uploads.remove(id) }
        promise.reject("ERR_ARCHIVE_REQUEST", "Could not start archive upload.", error)
      }
    }.runOnQueue(Queues.MAIN)
    AsyncFunction("cancel") { id: String -> synchronized(uploads) { uploads[id]?.cancel() } }.runOnQueue(Queues.MAIN)
    OnDestroy {
      synchronized(uploads) { closed = true; uploads.values.forEach { it.cancel() } }
      client.dispatcher.executorService.shutdown()
      client.connectionPool.evictAll()
    }
  }
}
