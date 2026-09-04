const exactReleaseTagPattern = /^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/;

function resolveMobileReleaseVersion({ tag, production }) {
  const normalizedTag = tag?.trim() ?? '';
  if (normalizedTag) {
    const match = exactReleaseTagPattern.exec(normalizedTag);
    if (!match) {
      throw new Error('STUFF_STASH_MOBILE_RELEASE_TAG must match vMAJOR.MINOR.PATCH.');
    }
    return `${match[1]}.${match[2]}.${match[3]}`;
  }

  if (production) {
    throw new Error('STUFF_STASH_MOBILE_RELEASE_TAG is required for production mobile builds.');
  }
  return '0.0.0';
}

module.exports = { resolveMobileReleaseVersion };
