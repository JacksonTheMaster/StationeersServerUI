(function () {
  const repository = "https://github.com/SteamServerUI/StationeersServerUI";
  const latestReleaseApi = "https://api.github.com/repos/SteamServerUI/StationeersServerUI/releases/latest";
  const cacheKey = "ssui-docs-latest-release-v1";
  const cacheDuration = 30 * 60 * 1000;

  const assetPatterns = {
    windows: [
      /^StationeersServerUI_v.+_windows_amd64\.exe$/i,
      /^StationeersServerControlv?.+\.exe$/i
    ],
    linux: [
      /^StationeersServerUI_v.+_linux_amd64$/i,
      /^StationeersServerControlv?.+\.x86_64$/i
    ]
  };

  function readCache() {
    try {
      const cached = JSON.parse(localStorage.getItem(cacheKey) || "null");
      if (cached && Date.now() - cached.savedAt < cacheDuration) {
        return cached.release;
      }
    } catch (_) {
      localStorage.removeItem(cacheKey);
    }
    return null;
  }

  function writeCache(release) {
    try {
      localStorage.setItem(cacheKey, JSON.stringify({ savedAt: Date.now(), release }));
    } catch (_) {
      // Private browsing and disabled storage should not break downloads.
    }
  }

  async function fetchRelease() {
    const cached = readCache();
    if (cached) {
      return cached;
    }

    const response = await fetch(latestReleaseApi, {
      headers: { Accept: "application/vnd.github+json" }
    });
    if (!response.ok) {
      const error = new Error(`GitHub returned ${response.status}`);
      error.status = response.status;
      throw error;
    }

    const release = await response.json();
    writeCache(release);
    return release;
  }

  function findAsset(release, platform) {
    const patterns = assetPatterns[platform] || [];
    const assets = release.assets || [];
    const preferred = patterns[0];
    const fallback = patterns.slice(1);
    return assets.find(asset => preferred && preferred.test(asset.name))
      || assets.find(asset => fallback.some(pattern => pattern.test(asset.name)))
      || null;
  }

  function setLink(element, asset, fallback) {
    if (!element) return;
    element.href = asset ? asset.browser_download_url : fallback;
    element.textContent = asset ? `Download for ${element.closest("[data-platform]")?.querySelector("strong")?.textContent || "platform"}` : "Open release page";
  }

  function formatReleaseDate(value) {
    if (!value) return "";
    const date = new Date(value);
    return Number.isNaN(date.valueOf()) ? "" : `Published ${date.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" })}`;
  }

  function showRelease(widget, release) {
    const windows = findAsset(release, "windows");
    const linux = findAsset(release, "linux");
    const releaseUrl = release.html_url || `${repository}/releases/latest`;

    widget.querySelector("#ssui-release-version").textContent = release.tag_name || "Latest stable release";
    widget.querySelector("#ssui-release-description").textContent = [
      release.name && release.name !== release.tag_name ? release.name : "Stationeers Server UI",
      formatReleaseDate(release.published_at)
    ].filter(Boolean).join(" · ");
    widget.querySelector("#ssui-release-page").href = releaseUrl;

    const windowsLink = widget.querySelector("#ssui-download-windows");
    const linuxLink = widget.querySelector("#ssui-download-linux");
    setLink(windowsLink, windows, releaseUrl);
    setLink(linuxLink, linux, releaseUrl);
    widget.querySelector("#ssui-asset-windows").textContent = windows ? windows.name : "No Windows amd64 asset found";
    widget.querySelector("#ssui-asset-linux").textContent = linux ? linux.name : "No Linux amd64 asset found";
    widget.querySelector("#ssui-release-status").textContent = "Release data loaded from GitHub. The links above point to the actual assets.";
    widget.dataset.state = "ready";
  }

  function showError(widget, error) {
    const releaseUrl = `${repository}/releases/latest`;
    widget.querySelector("#ssui-release-version").textContent = "Release lookup unavailable";
    widget.querySelector("#ssui-release-description").textContent = "GitHub did not return release metadata right now.";
    widget.querySelector("#ssui-release-page").href = releaseUrl;
    widget.querySelector("#ssui-download-windows").href = releaseUrl;
    widget.querySelector("#ssui-download-linux").href = releaseUrl;
    widget.querySelector("#ssui-download-windows").textContent = "Open release page";
    widget.querySelector("#ssui-download-linux").textContent = "Open release page";
    widget.querySelector("#ssui-asset-windows").textContent = "Use GitHub releases manually";
    widget.querySelector("#ssui-asset-linux").textContent = "Use GitHub releases manually";
    widget.querySelector("#ssui-release-status").textContent = error.status === 403
      ? "GitHub's API rate limit was reached. The manual release link still works."
      : "The GitHub API could not be reached. The manual release link still works.";
    widget.dataset.state = "error";
  }

  function initializeDownloads() {
    const widget = document.querySelector("[data-ssui-download-widget]");
    if (!widget || widget.dataset.initialized === "true") return;
    widget.dataset.initialized = "true";

    fetchRelease().then(release => showRelease(widget, release)).catch(error => showError(widget, error));
  }

  document$.subscribe(initializeDownloads);
})();
