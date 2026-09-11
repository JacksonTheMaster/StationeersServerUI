# Downloads

Pick a platform and the page will query the latest stable SSUI release directly from GitHub. The filename shown below is the actual release asset you are about to download, not a filename guessed from an old install.

<div class="ssui-download-widget" data-ssui-download-widget aria-live="polite">
<div class="ssui-download-summary">
<div>
<p class="ssui-download-kicker">Latest stable release</p>
<h2 id="ssui-release-version">Checking GitHub...</h2>
<p id="ssui-release-description">Loading the current release metadata.</p>
</div>
<a id="ssui-release-page" class="md-button" href="https://github.com/SteamServerUI/StationeersServerUI/releases/latest" target="_blank" rel="noopener">View release</a>
</div>

<div class="ssui-download-grid">
<article class="ssui-download-card" data-platform="windows">
<div class="ssui-download-platform"><span aria-hidden="true">▣</span><strong>Windows 64-bit</strong></div>
<p>Run the `.exe` from a writable installation folder.</p>
<a id="ssui-download-windows" class="md-button md-button--primary" href="https://github.com/SteamServerUI/StationeersServerUI/releases/latest" target="_blank" rel="noopener">Open release page</a>
<code id="ssui-asset-windows">Waiting for release data...</code>
</article>

<article class="ssui-download-card" data-platform="linux">
<div class="ssui-download-platform"><span aria-hidden="true">▰</span><strong>Linux amd64</strong></div>
<p>Download the executable, then make it runnable with `chmod +x`.</p>
<a id="ssui-download-linux" class="md-button md-button--primary" href="https://github.com/SteamServerUI/StationeersServerUI/releases/latest" target="_blank" rel="noopener">Open release page</a>
<code id="ssui-asset-linux">Waiting for release data...</code>
</article>

<article class="ssui-download-card" data-platform="docker">
<div class="ssui-download-platform"><span aria-hidden="true">▤</span><strong>Docker</strong></div>
<p>Pull the image and keep the complete installation in one `/app` bind mount.</p>
<a class="md-button md-button--primary" href="Docker-Guide/">Open Docker guide</a>
<code>ghcr.io/steamserverui/stationeersserverui:latest</code>
</article>
</div>

<p class="ssui-download-status" id="ssui-release-status">GitHub release lookup is loading.</p>
</div>

<noscript>
<div class="admonition warning">
<p class="admonition-title">JavaScript is disabled</p>
<p>This page cannot query release assets without JavaScript. Open the <a href="https://github.com/SteamServerUI/StationeersServerUI/releases/latest">latest GitHub release</a> manually instead.</p>
</div>
</noscript>

If GitHub's API is temporarily unavailable, use [GitHub releases](https://github.com/SteamServerUI/StationeersServerUI/releases) and choose the asset for your platform. The page caches a successful lookup in your browser for a short time so opening the documentation does not hit the API on every visit.

## Which file do I need?

| Platform | Asset to choose | Notes |
|---|---|---|
| Windows 64-bit | Windows amd64 executable | It has the `.exe` suffix. Install the current Microsoft Visual C++ runtime if Stationeers asks for it. |
| Linux 64-bit | Linux amd64 executable | It has no `.exe` suffix. Make it executable with `chmod +x` before the first run. |
| Docker | Published container image | Do not download a native executable into the container. Pull the image and keep the complete installation in the `/app` bind mount. |

The release asset name includes the SSUI version and platform. Use the asset from the release you selected instead of guessing from an older filename. If you are on v5 and moving to v6, read the [upgrade notes](Installation.md#upgrading-from-ssui-v5) before replacing anything.

For a complete migration checklist, use [Upgrade from v5](Upgrade-Guide.md). It explains the API, backup identity, authentication and Docker changes that are easy to miss when the first browser login looks normal.

## Stable, nightly and release candidates

- **Stable** is the normal choice for a server you want to forget about for a while.
- **Release candidates** are intended for real-world testing before a stable release. They can contain rough edges and require deliberate update confirmation.
- **Nightly** builds are development builds. Use them when you are testing a fix or feature and are prepared to report what broke.

Do not mix the executable from one release with the runtime data of an unrelated installation while troubleshooting. Keep the binary and its installation folder together, and keep a copy of the complete data directory before a major upgrade.

## After downloading

Continue with the platform-specific [installation guide](Installation.md), or jump directly to [first-time setup](First-Time-Setup.md) if SSUI is already running.

---

[Documentation home](index.md) · [Installation](Installation.md) · [First-time setup](First-Time-Setup.md)
