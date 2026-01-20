My config have todoist and nextcloud setup, yet I've only been able to use SQLite.

Step:
```
# Launch todoat to get config sample, add todoist and nextcloud config, set any of these that has lists of task as default backend, sync disabled
todoat list
No lists found. Create one with: todoat list create "MyList"
```

It should have shown list in one of the 2 backend config, but it clearly just use sqlite db instead
