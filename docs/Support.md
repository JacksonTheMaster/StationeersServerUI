# Get help

SSUI is maintained by people who also run servers, break test installations and occasionally discover that the obvious network problem was, in fact, the network problem. Ask for help before deleting evidence.

<div class="ssui-card-grid">
<a class="ssui-card" href="https://discord.gg/8n3vN92MyJ"><strong>Join the SSUI Discord</strong><span>Best for setup questions, live diagnosis and figuring out which layer is actually unhappy.</span><span class="ssui-card-arrow">↗</span></a>
<a class="ssui-card" href="https://github.com/SteamServerUI/StationeersServerUI/issues"><strong>Open an issue</strong><span>Best for reproducible bugs, documentation errors and changes that need to stay trackable.</span><span class="ssui-card-arrow">↗</span></a>
</div>

## Before you ask

Do the smallest useful check first:

1. Note the SSUI version or branch and the installation method.
2. Write down the operating system or container image.
3. Reproduce the problem once and keep the first relevant error with its timestamp.
4. Say what you expected, what happened instead and what changed immediately before it.
5. If the issue is not obvious, create a [support package](Support-Packages.md) and review it before sharing it.

The [troubleshooting guide](Troubleshooting.md) is organized by symptom. It is a starting point, not an entrance exam. If a step does not fit your setup, say so instead of forcing the setup into the step.

## Where to post what

| Situation | Best place |
|---|---|
| You are stuck during installation or configuration | Discord, with the exact platform and console output |
| Players cannot connect | Discord, with the gameserver port, firewall/router setup and Spacecat result |
| A reproducible crash or regression | GitHub issue, with version, steps and the first error |
| A documentation claim is wrong or unclear | GitHub issue or a focused pull request |
| You need to share logs or configuration | A reviewed support package through a private channel agreed with support |

## Keep secrets out of support requests

Never post passwords, browser cookies, personal API tokens, Discord bot tokens, TLS private keys or a raw `config.json`. Support packages remove known configuration fields, but logs can still contain player names, Steam IDs, paths, addresses or secrets that were printed earlier. Open the archive yourself before sharing it.

If you lose access to an owner account, use the documented [recovery flag](Command-line-flags.md#recovery-user), sign in locally, change the temporary password and remove the recovery account on the next normal start.

## Report the useful version

The version printed by SSUI matters. So does whether you are on a stable release, release candidate or nightly build. Do not report only “latest”; that word has caused more archaeology than the backup manager.

---

[Documentation home](index.md) · [Troubleshooting](Troubleshooting.md) · [Support packages](Support-Packages.md)
