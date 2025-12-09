# J-Switch ☕

**J-Switch** is a lightweight, cross-platform command-line tool written in Go designed to make managing and switching between different Java (JDK) versions effortless.

Whether you are a developer juggling multiple projects with different Java requirements or just need a quick way to update your environment, J-Switch has you covered.

## 🚀 Features

- **🔍 Auto-Discovery**: Automatically scans your system (`Program Files`, `/usr/lib/jvm`, etc.) to find installed JDKs.
- **⚡ Fast Switching**: Switch your active Java version instantly. Updates `JAVA_HOME` and `PATH` system environment variables.
- **⬇️ Built-in Downloader**: Fetch and install the latest Java versions directly from the [Eclipse Adoptium](https://adoptium.net/) API.
- **🖥️ Beautiful TUI**: Interactive terminal user interface for easy selection.
- **🪟 Cross-Platform**: Works on Windows, macOS, and Linux.

## 📦 Installation

*(Binaries coming soon)*

To build from source:

```bash
git clone https://github.com/mattbaconz/J-Switch.git
cd J-Switch
go build -o jswitch ./cmd/jswitch
```

## 🛠️ Usage

Running `jswitch` without arguments opens the interactive UI.

```bash
# Open the interactive UI
jswitch

# Scan your system for Java installations
jswitch scan

# List known installations
jswitch list

# Install a specific Java version (e.g., Java 17)
jswitch install 17

# Switch to a specific version via CLI
jswitch use 17
```

## 🔗 Connect & Support

If you find this tool useful, consider supporting the development or joining the community!

- 💬 **Discord**: [Join the Community](https://discord.com/invite/VQjTVKjs46)
- ☕ **Ko-fi**: [Buy me a coffee](https://ko-fi.com/mbczishim/tip)
- 💸 **PayPal**: [Donate](https://www.paypal.com/paypalme/MatthewWatuna)
- 🐙 **GitHub**: [mattbaconz](https://github.com/mattbaconz)

---

Built with ❤️ by [Matt Bacon](https://github.com/mattbaconz).
