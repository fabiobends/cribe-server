# Cribe Server

A backend server for the Cribe app.

## ⚙️ Setup
- Install [Go](https://go.dev).
- Install [Air](https://github.com/air-verse/air):
  - Use the [install.sh](https://github.com/air-verse/air#via-installsh) script to install Air:
  ```bash
  curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh
  ```
  - Ensure Air is added to your `$PATH`. Follow the instructions [here](https://github.com/air-verse/air#command-not-found-air-or-no-such-file-or-directory) to verify or update your `$PATH`.

## ▶️ Run

### Development (hot-reload)
```bash
air
```

### Production
```bash
go build
./cribe-server
```

## 📜 License

[MIT](LICENSE)