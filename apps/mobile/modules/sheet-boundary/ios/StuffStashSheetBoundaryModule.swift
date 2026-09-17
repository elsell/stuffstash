import ExpoModulesCore
import UIKit
import React

public final class StuffStashSheetBoundaryModule: Module {
  public func definition() -> ModuleDefinition {
    Name("StuffStashSheetBoundary")
    View(SheetBoundaryView.self) {
      AsyncFunction("measureInKeyboardWindow") { (view: SheetBoundaryView) -> [String: Double]? in
        // RCTKeyboardObserver publishes coordinates converted into RCTKeyWindow.
        // Only measure when that is also this boundary's own window.
        guard let window = view.window, window === RCTKeyWindow() else { return nil }
        let inWindow = view.convert(view.bounds, to: window)
        return ["x": Double(inWindow.minX), "y": Double(inWindow.maxY), "width": Double(inWindow.width)]
      }
    }
  }
}

final class SheetBoundaryView: ExpoView {}
