#!/usr/bin/env python3
"""Compile the installed picker's numeric conversion, rather than a copied formula.

Run on the pinned macOS CI toolchain. UIKit presentation is verified separately.
"""

from pathlib import Path
import re
import subprocess
import tempfile


def main():
    root = Path(__file__).resolve().parent.parent
    source = (root / "apps/mobile/node_modules/@expo/ui/ios/ColorPickerView.swift").read_text()
    conversion = source.split("private static func colorToHex(", 1)
    if len(conversion) != 2:
        raise SystemExit("ColorPicker converter changed; review the native contract harness")
    expressions = re.findall(r"\]\.map\s*\{\s*([^\n{}]+)\s*\}", conversion[1])
    if len(expressions) != 1:
        raise SystemExit("Expected exactly one native channel conversion expression")
    harness = r'''
import Foundation

let quantize: (CGFloat) -> Int = { CONVERSION }
var checks = 0
func verify(_ channel: CGFloat, _ expected: Int, _ context: String) {
    checks += 1
    let actual = quantize(channel)
    guard actual == expected else {
        fatalError("\(context): expected \(expected), got \(actual) for \(channel)")
    }
}

for byte in 0...255 {
    let channel = CGFloat(byte) / 255
    verify(channel, byte, "exact eight-bit round trip")
    verify(channel.nextDown, byte, "adjacent lower representation")
    verify(channel.nextUp, byte, "adjacent upper representation")
    var feedback = channel
    for _ in 0..<32 {
        verify(feedback, byte, "untouched channel in controlled feedback")
        feedback = CGFloat(quantize(feedback)) / 255
    }
}
for byte in 0..<255 {
    verify((CGFloat(byte) + 0.49) / 255, byte, "below nearest-byte boundary")
    verify((CGFloat(byte) + 0.51) / 255, byte + 1, "above nearest-byte boundary")
}
verify(-0.1, 0, "lower clamp")
verify(1.1, 255, "upper clamp")
print("Passed \(checks) installed native color conversion checks")
'''.replace("CONVERSION", expressions[0].strip())
    with tempfile.TemporaryDirectory(prefix="stuffstash-color-contract-") as temp:
        script = Path(temp) / "ColorConversionContract.swift"
        script.write_text(harness)
        subprocess.run(["swift", str(script)], check=True, cwd=root)


if __name__ == "__main__":
    main()
