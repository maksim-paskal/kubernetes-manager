# Delete feature branch when it was merged into main

This extentions will watch your git repository every 30m to detect stale git branch, for example your git branch was merged into main branch or last commit was more than 20 days ago `-batch.removeBranchDaysInactive=20` - in this case kubernetes namespace and docker registry tags will be deleted

*It will not delete you branch if your git branch is detect as system default is `-system.gitTags=^master$,^release-.*` default scenario can not delete `master` git branch or branch with name `release-*`*
