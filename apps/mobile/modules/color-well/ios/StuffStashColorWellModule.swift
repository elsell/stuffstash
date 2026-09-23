import ExpoModulesCore
import UIKit

public final class StuffStashColorWellModule: Module {
  public func definition() -> ModuleDefinition {
    Name("StuffStashColorWell")
    View(TagColorWellView.self) {
      Events("onSelectionChange")
      Prop("selection") { (view: TagColorWellView, value: String?) in view.setSelection(value) }
      Prop("enabled") { (view: TagColorWellView, enabled: Bool) in view.setEnabled(enabled) }
    }
  }
}

final class TagColorWellView: ExpoView, UIColorPickerViewControllerDelegate, UIPopoverPresentationControllerDelegate {
  private let button = UIButton(type: .system)
  private var selection: UIColor?
  private var picker: UIColorPickerViewController?
  let onSelectionChange = EventDispatcher()

  required init(appContext: AppContext? = nil) {
    super.init(appContext: appContext)
    isAccessibilityElement = false
    button.setImage(UIImage(systemName: "paintpalette"), for: .normal)
    button.setPreferredSymbolConfiguration(UIImage.SymbolConfiguration(pointSize: 24), forImageIn: .normal)
    button.accessibilityLabel = "Choose any color"
    button.addTarget(self, action: #selector(openPicker), for: .touchUpInside)
    addSubview(button)
  }

  override func layoutSubviews() {
    super.layoutSubviews()
    button.frame = bounds
  }

  override func didMoveToWindow() {
    super.didMoveToWindow()
    if window == nil { retirePicker(animated: false) }
  }

  func setEnabled(_ enabled: Bool) {
    button.isEnabled = enabled
    if !enabled { retirePicker(animated: true) }
  }

  func setSelection(_ value: String?) {
    selection = TagColorRGB.components(value).map {
      UIColor(red: CGFloat($0.red), green: CGFloat($0.green), blue: CGFloat($0.blue), alpha: 1)
    }
    button.accessibilityValue = value
    if selection == nil { retirePicker(animated: true) }
    if let selection, picker?.selectedColor != selection { picker?.selectedColor = selection }
  }

  @objc private func openPicker() {
    guard button.isEnabled, window != nil, picker == nil else { return }
    var responder: UIResponder? = self
    while let current = responder, !(current is UIViewController) { responder = current.next }
    guard let presenter = responder as? UIViewController,
          presenter.viewIfLoaded?.window != nil, presenter.presentedViewController == nil else { return }
    let controller = UIColorPickerViewController()
    controller.supportsAlpha = false
    if let selection { controller.selectedColor = selection }
    controller.delegate = self
    controller.modalPresentationStyle = traitCollection.horizontalSizeClass == .regular ? .popover : .pageSheet
    controller.presentationController?.delegate = self
    controller.popoverPresentationController?.sourceView = button
    controller.popoverPresentationController?.sourceRect = button.bounds
    controller.popoverPresentationController?.delegate = self
    picker = controller
    presenter.present(controller, animated: true)
  }

  func colorPickerViewControllerDidSelectColor(_ viewController: UIColorPickerViewController) {
    guard picker === viewController, button.isEnabled, window != nil else { return }
    var red: CGFloat = 0, green: CGFloat = 0, blue: CGFloat = 0, alpha: CGFloat = 0
    guard viewController.selectedColor.getRed(&red, green: &green, blue: &blue, alpha: &alpha),
          let value = TagColorRGB.hex(red: Double(red), green: Double(green), blue: Double(blue)) else { return }
    onSelectionChange(["value": value])
  }

  func colorPickerViewControllerDidFinish(_ viewController: UIColorPickerViewController) {
    guard picker === viewController else { return }
    picker = nil
  }

  func presentationControllerDidDismiss(_ presentationController: UIPresentationController) {
    guard picker === presentationController.presentedViewController else { return }
    picker = nil
  }

  private func retirePicker(animated: Bool) {
    let controller = picker
    picker = nil
    controller?.delegate = nil
    controller?.dismiss(animated: animated)
  }
}
