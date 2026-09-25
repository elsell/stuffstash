import ExpoModulesCore
import UIKit

public final class StuffStashCommandButtonModule: Module {
  public func definition() -> ModuleDefinition {
    Name("StuffStashCommandButton")
    View(CommandButtonView.self) {
      Events("onPress", "onSizeChange")
      Prop("label") { (view: CommandButtonView, value: String) in view.label = value }
      Prop("accessibilityLabel") { (view: CommandButtonView, value: String) in
        view.button.accessibilityLabel = value
      }
      Prop("prominence") { (view: CommandButtonView, value: String) in view.prominence = value }
      Prop("fullWidth") { (view: CommandButtonView, value: Bool) in view.fullWidth = value }
      Prop("role") { (view: CommandButtonView, value: String) in view.destructive = value == "destructive" }
      Prop("disabled") { (view: CommandButtonView, value: Bool) in view.button.isEnabled = !value }
    }
  }
}

final class CommandButtonView: ExpoView {
  let button = UIButton(type: .system)
  let onPress = EventDispatcher()
  let onSizeChange = EventDispatcher()
  var label = "" { didSet { configure() } }
  var prominence = "secondary" { didSet { configure() } }
  var fullWidth = false { didSet { setNeedsLayout() } }
  var destructive = false { didSet { configure() } }
  private let minimumTarget: CGFloat = 48
  private let insets = NSDirectionalEdgeInsets(top: 10, leading: 16, bottom: 10, trailing: 16)
  private var reportedHeight: CGFloat = 0

  required init(appContext: AppContext? = nil) {
    super.init(appContext: appContext)
    isAccessibilityElement = false
    addSubview(button)
    button.addTarget(self, action: #selector(activate), for: .touchUpInside)
    configure()
  }

  @objc private func activate() {
    guard button.isEnabled else { return }
    onPress([:])
  }

  private var titleFont: UIFont { UIFont.preferredFont(forTextStyle: .body, compatibleWith: traitCollection) }

  private func configure() {
    var configuration: UIButton.Configuration
    switch prominence {
    case "primary": configuration = .filled()
    case "standard": configuration = .plain()
    default: configuration = .tinted()
    }
    configuration.title = label
    configuration.cornerStyle = .capsule
    configuration.contentInsets = insets
    configuration.baseBackgroundColor = destructive ? .systemRed : .systemBlue
    configuration.baseForegroundColor = prominence == "primary" ? .white : (destructive ? .systemRed : .systemBlue)
    let font = titleFont
    configuration.titleTextAttributesTransformer = UIConfigurationTextAttributesTransformer { attributes in
      var updated = attributes
      updated.font = font
      return updated
    }
    button.configuration = configuration
    button.titleLabel?.numberOfLines = 0
    button.titleLabel?.lineBreakMode = .byWordWrapping
    button.titleLabel?.textAlignment = .center
    setNeedsLayout()
  }

  override func traitCollectionDidChange(_ previousTraitCollection: UITraitCollection?) {
    super.traitCollectionDidChange(previousTraitCollection)
    if previousTraitCollection?.preferredContentSizeCategory != traitCollection.preferredContentSizeCategory {
      configure()
    }
  }

  override func layoutSubviews() {
    super.layoutSubviews()
    guard bounds.width > 0 else { return }
    let font = titleFont
    let horizontalInsets = insets.leading + insets.trailing
    let naturalWidth = ceil((label as NSString).size(withAttributes: [.font: font]).width) + horizontalInsets
    let width = (fullWidth || prominence == "primary") ? bounds.width : min(bounds.width, max(minimumTarget, naturalWidth))
    let textWidth = max(1, width - horizontalInsets)
    let paragraph = NSMutableParagraphStyle()
    paragraph.lineBreakMode = .byWordWrapping
    let textHeight = (label as NSString).boundingRect(
      with: CGSize(width: textWidth, height: .greatestFiniteMagnitude),
      options: [.usesLineFragmentOrigin, .usesFontLeading],
      attributes: [.font: font, .paragraphStyle: paragraph], context: nil
    ).height
    let height = max(minimumTarget, ceil(textHeight) + insets.top + insets.bottom)
    let x = effectiveUserInterfaceLayoutDirection == .rightToLeft ? bounds.width - width : 0
    button.frame = CGRect(x: x, y: 0, width: width, height: height)
    if height != reportedHeight {
      reportedHeight = height
      onSizeChange(["height": height])
    }
  }
}
