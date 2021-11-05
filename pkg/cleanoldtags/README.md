# Clear old docker registry tags

This extention can reduce docker registry size by removing old registry tags

to delete old releases you need to name docker tags it with date of release - for example `release-20210101-sometxt` it can be changed with arg `-release.pattern=release-(\\d{4}\\d{2}\\d{2}).*`

Deletion will read all releases in your docker registry and will leave only last 10 day of releases `-release.notDeleteDays=10` for example

```bash
release-20210415-with-master
release-20210415-hotfix
release-20210413
release-20210411
release-20210410
release-20210410-fix
release-20210405
release-20210404 (will be deleted)
release-20210401 (will be deleted)
release-20210325 (will be deleted)
```
