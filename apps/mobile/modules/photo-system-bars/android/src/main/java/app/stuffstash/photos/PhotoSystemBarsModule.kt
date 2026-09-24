package app.stuffstash.photos

import android.content.Context
import android.os.Build
import android.view.View
import android.view.WindowInsetsController
import expo.modules.kotlin.AppContext
import expo.modules.kotlin.modules.Module
import expo.modules.kotlin.modules.ModuleDefinition
import expo.modules.kotlin.views.ExpoView

class PhotoSystemBarsModule : Module() {
  override fun definition() = ModuleDefinition {
    Name("StuffStashPhotoSystemBars")
    View(PhotoSystemBarsView::class) {}
  }
}

class PhotoSystemBarsView(context: Context, appContext: AppContext) : ExpoView(context, appContext) {
  override fun onAttachedToWindow() {
    super.onAttachedToWindow()
    applyDialogAppearance()
  }

  override fun onWindowFocusChanged(hasWindowFocus: Boolean) {
    super.onWindowFocusChanged(hasWindowFocus)
    // React Native copies Activity appearance after Dialog.show(), then permits
    // focus. Own the dialog appearance after that native lifecycle boundary.
    if (hasWindowFocus) applyDialogAppearance()
  }

  @Suppress("DEPRECATION")
  private fun applyDialogAppearance() {
    if (!isAttachedToWindow || rootView === appContext.currentActivity?.window?.decorView) return
    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
      windowInsetsController?.setSystemBarsAppearance(
        0,
        WindowInsetsController.APPEARANCE_LIGHT_STATUS_BARS or
          WindowInsetsController.APPEARANCE_LIGHT_NAVIGATION_BARS
      )
    } else {
      var lightFlags = View.SYSTEM_UI_FLAG_LIGHT_STATUS_BAR
      if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
        lightFlags = lightFlags or View.SYSTEM_UI_FLAG_LIGHT_NAVIGATION_BAR
      }
      rootView.systemUiVisibility = rootView.systemUiVisibility and lightFlags.inv()
    }
  }
}
