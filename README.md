# Cribe Server

A backend server for the Cribe app.

## 💡 Setup

1. Make sure [Go](https://go.dev) and your path are set up.

2. Make sure [Docker](https://www.docker.com/) is installed and running.

3. Copy the `.env.example` file to `.env.dev` and make any changes you need.
```bash
cp .env.example .env.dev
```

## ⚙️ Installing dependencies

Install the following dependencies:

  - [Air](https://github.com/air-verse/air)
  - [golangci-lint](https://golangci-lint.run/)

## 🚀 Running the server

Run the following commands to start the server:

**Development (hot-reload)**
```bash
make dev
```

**Development (no hot-reload)**
```bash
make run
```

**Production**
```bash
make build
./cribe-server
```

## 🧪 Testing
Run the following command to run the tests:

```bash
make test
```

## 📜 License

[MIT](LICENSE)
