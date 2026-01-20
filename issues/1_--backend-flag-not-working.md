The user should be able to select the backend with `--backend` flag

```bash
todoat --backend nextcloud-test mytestlist a test-1

```

This would select the `nextcloud-test` config in user config `config.yml`.

Current behaviour:
```
todoat --backend todoist list
Error: unknown flag: --backend
```
