import ExpoModulesCore
import UIKit

public final class StuffStashColorWellModule: Module {
  public func definition() -> ModuleDefinition {
    Name("StuffStashColorWell")
    View(TagColorWellView.self) {
      Events("onSelectionChange")
      Prop("selection") { (view: TagColorWellView, value: String?) in view.setSelection(value) }
      Prop("enabled") { (view: TagColorWellView, enabled: Bool) in view.well.isEnabled = enabled }
    }
  }
}

final class TagColorWellView: ExpoView {
  let well = UIColorWell()
  let onSelectionChange = EventDispatcher()

  required init(appContext: AppContext? = nil) {
    super.init(appContext: appContext)
    isAccessibilityElement = false
    well.supportsAlpha = false
    well.title = "Choose any color"
    well.accessibilityLabel = "Choose any color"
    well.selectedColor = nil
    well.addTarget(self, action: #selector(selectionChanged), for: .valueChanged)
    addSubview(well)
  }

  override func layoutSubviews() {
    super.layoutSubviews()
    well.frame = bounds
  }

  func setSelection(_ value: String?) {
    guard let rgb = TagColorRGB.components(value) else {
      well.selectedColor = nil
      return
    }
    let color = UIColor(red: CGFloat(rgb.red), green: CGFloat(rgb.green), blue: CGFloat(rgb.blue), alpha: 1)
    if well.selectedColor != color { well.selectedColor = color }
  }

  @objc private func selectionChanged() {
    guard well.isEnabled, let color = well.selectedColor else { return }
    var red: CGFloat = 0, green: CGFloat = 0, blue: CGFloat = 0, alpha: CGFloat = 0
    guard color.getRed(&red, green: &green, blue: &blue, alpha: &alpha),
          let value = TagColorRGB.hex(red: Double(red), green: Double(green), blue: Double(blue)) else { return }
    onSelectionChange(["value": value])
  }
}
