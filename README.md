# Antigravity Usage Checker

🚀 Check your Antigravity AI usage quota from terminal

![Version](https://img.shields.io/badge/version-0.3.0-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Platform](https://img.shields.io/badge/platform-Windows-lightgrey)

## 🇬🇧 English

### Installation

1. **Download** [`antigravity-usage-checker-windows.zip`](https://github.com/TungCorn/antigravity-usage-checker/releases/latest)

2. **Extract** the zip file

3. **Run install script** (PowerShell):
```powershell
powershell -ExecutionPolicy Bypass -File install.ps1
```

4. **Restart terminal** and run:
```bash
agusage
```

> 💡 **Tip**: If `agusage` is not found, run this to refresh PATH:
> ```powershell
> $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
> ```

> ⚠️ Antigravity must be running

### Output

![Screenshot](assets/image.png)

### Commands

```bash
agusage          # Show quota
agusage --json   # JSON output
agusage --help   # Help
```

---

## 🇻🇳 Tiếng Việt

### Cài đặt

1. **Tải** [`antigravity-usage-checker-windows.zip`](https://github.com/TungCorn/antigravity-usage-checker/releases/latest)

2. **Giải nén** file zip

3. **Chạy script cài đặt** (PowerShell):
```powershell
powershell -ExecutionPolicy Bypass -File install.ps1
```

4. **Khởi động lại terminal** và chạy:
```bash
agusage
```

> 💡 **Mẹo**: Nếu lệnh `agusage` không tìm thấy, chạy lệnh này để refresh PATH:
> ```powershell
> $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
> ```

> ⚠️ Antigravity phải đang chạy

### Kết quả

![Screenshot](assets/image.png)

### Các lệnh

```bash
agusage          # Xem quota
agusage --json   # Xuất JSON
agusage --help   # Trợ giúp
```

---

## License

MIT © 2024

---

<p align="center">
  <b>If you find this useful, give it a ⭐!</b>
</p>
