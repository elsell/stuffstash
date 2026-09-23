import Foundation

enum TagColorRGB {
  static func components(_ hex: String?) -> (red: Double, green: Double, blue: Double)? {
    guard let hex, hex.range(of: "^#[0-9A-Fa-f]{6}$", options: .regularExpression) != nil,
          let value = UInt32(hex.dropFirst(), radix: 16) else { return nil }
    return (Double((value >> 16) & 255) / 255, Double((value >> 8) & 255) / 255, Double(value & 255) / 255)
  }
  static func hex(red: Double, green: Double, blue: Double) -> String? {
    let channels = [red, green, blue]
    guard channels.allSatisfy({ $0.isFinite }) else { return nil }
    let bytes = channels.map { Int(max(0, min(255, ($0 * 255).rounded()))) }
    return String(format: "#%02X%02X%02X", bytes[0], bytes[1], bytes[2])
  }
}
