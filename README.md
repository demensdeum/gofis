# Gofis - Go File Search CLI Tool

**Gofis** is a lightweight, high-performance command-line utility built in Go for searching files across your system. It leverages Go's native concurrency model to walk directory trees rapidly, making it significantly faster than traditional single-threaded search tools.

---

### 🛠️ Features

* **Concurrent Searching:** Uses a configurable pool of goroutines to scan multiple directories simultaneously.
* **Smart Filtering:** Search by partial filenames, specific extensions, or both.
* **Automatic Ignoring:** Defaults to skip heavy folders like `enode_modules`, `.git`, and `vendor` to save time.
* **Human-Readable Output:** Displays file sizes in a formatted, easy-to-read way (KB, MB, GB, etc.).
* **Flexible Interface:** Supports both traditional flags and shorthand positional arguments.

---

### 🛠️ Installation

Ensure you have [Go](https://go.dev/doc/install) installed on your machine.

```bash
git clone https://github.com/yourusername/gofis.git
cd gofis
go build -o gofis main.go
```

---

### 📖 Usage

Gofis provides two ways to search: using flags or using positional arguments.

### 1. Using Flags (Recommended)
```bash
./gofis -n "config" -e ".yaml" -p ./src
```

### 2. Positional Shorthand
```bash
./gofis "filename" "searchPath" [maxGoroutines]
```
Example: `./gofis "main" "./projects" 50`

---

### ⚙️ How it Works

Gofis uses a **Semaphore Pattern** to manage system resources. While it spawns a new goroutine for every subdirectory it encounters, it limits the number of active operations using a buffered channel (`sem`). This prevents the application from hitting "too many open files" errors or overwhelming the CPU.

### Performance Logic
* **WaitGroup:** Ensures the main process doesn't exit until all sub-directories have been fully scanned.
* **Channels:** Streams results back to the main thread in real-time, allowing you to see matches immediately without waiting for the entire scan to finish.

---

### 📄 License
This project is open-source and available under the [MIT License](LICENSE).
