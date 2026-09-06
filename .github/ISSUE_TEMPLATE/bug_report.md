---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: ''

---

Please add a support package to your bug report:

## How to generate a support package
1: clear config page fields -> `ServerAuthSecret` and `ServerPassword` as those fields currently do not get sanitized (bug, the other sentive fields are)
2: run `sm` in the SSCLI (terminal console that opens when you start SSUI)
3: reproduce the issue
4: type `sp` into SSCLI to generate the support package
5: type `sm` again to disable support mode again
6: drop the support package zip that was created and saved to the server directory either in the bug report or on the [SSUI Discord 
](https://discord.gg/8n3vN92MyJ)

**Describe the bug**



**To Reproduce**
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. etc

**Screenshots**
If applicable, add screenshots to help explain your problem.
