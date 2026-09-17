"""Runner-only M207 experiment; never invoked by production builds."""
from pathlib import Path


def configure_before_attach(source: str) -> str:
    assignment = "          navitem.searchController = searchBar.controller;\n"
    boundary = "#endif /* Check for iOS 26.0 */\n#endif /* !TARGET_OS_TV */"
    if source.count(assignment) != 1 or source.count(boundary) != 1:
        raise ValueError("Pinned search-controller attachment source changed")
    start = source.index(assignment)
    end = source.index(boundary, start)
    section = source[start:end]
    if "preferredSearchBarPlacement" not in section or "searchBarPlacementAllowsToolbarIntegration" not in section:
        raise ValueError("Expected native placement settings before attachment boundary")
    return source.replace(assignment, "", 1).replace(boundary,
        "#endif /* Check for iOS 26.0 */\n" + assignment + "#endif /* !TARGET_OS_TV */", 1)


if __name__ == "__main__":
    root = Path(__file__).resolve().parents[1]
    paths = list((root / "node_modules/.pnpm").glob(
        "react-native-screens@4.23.0*/node_modules/react-native-screens/ios/RNSScreenStackHeaderConfig.mm"))
    if len(paths) != 1:
        raise RuntimeError("Expected exactly one pinned react-native-screens source")
    path = paths[0]
    path.write_text(configure_before_attach(path.read_text()))
