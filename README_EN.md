# FERS - File Encrypt & Remote Storage

<div align="center">

![FERS Logo](Icon.jpg)

**A secure and user-friendly file encryption and cloud storage management tool**

[![Go Version](https://img.shields.io/badge/Go-1.24.6+-blue.svg)](https://golang.org/)
[![Fyne](https://img.shields.io/badge/GUI-Fyne%20v2.6.3-green.svg)](https://fyne.io/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

</div>

## ✨ Project Features

FERS (File Encrypt & Remote Storage) is a file management tool designed specifically for privacy protection, featuring the following unique advantages:

### 🔐 **End-to-End Encryption**

- Uses **AES-256-GCM** military-grade encryption algorithm
- Files are encrypted locally before upload, keys never leave your device
- Provides data integrity verification to prevent file tampering

### 🌐 **Multi-Storage Backend Support**

- **Alibaba Cloud OSS** - Production cloud storage environment
- **Local Simulation** - Development and testing environment
- Modular design, easily extensible to other cloud service providers

### 🖥️ **Intuitive Graphical Interface**

- Cross-platform GUI based on Fyne framework
- File browser-like operation experience
- Real-time log display with clear operation status
- Support for batch file selection and download

### ⚡ **Intelligent Sync Features**

- **Bidirectional Sync** - Automatically identifies file differences between local and remote
- **Incremental Upload** - Only uploads newly added local files
- **Selective Download** - Choose specific remote files to download
- **Operation Cancellation** - Long-running operations can be interrupted at any time

### 🛡️ **Security Design**

- Working directory restrictions prevent accidental access to system files
- Path validation prevents directory traversal attacks
- Structured logging for audit trails and troubleshooting

## 🚀 Quick Start

### System Requirements

- **Operating System**: Windows 10/11, macOS 10.14+, Linux
- **Go Version**: 1.24.6 or higher
- **Network**: Stable internet connection (when using cloud storage)

### Installation

#### Method 1: Download Pre-compiled Version

Download the executable file suitable for your system from the [Releases](https://github.com/mingregister/fers/releases) page.

#### Method 2: Build from Source

```bash
# Clone the project
git clone https://github.com/mingregister/fers.git
cd fers

# Build executable
go build -o fers main.go

# Windows users can use the build script
build.bat
```

### Configuration Setup

Create a `.fers/config.yaml` file in the user's home directory:

```yaml
# Encryption key (please use a strong password)
crypto_key: "your-strong-encryption-password"

# Log file path
log: "app.log"

# Working directory (all operations are restricted within this directory)
target_dir: "/path/to/your/working/directory"

# Storage configuration
storage:
  # Storage type: oss (Alibaba Cloud) or localhost (local testing)
  remote_type: "oss"
  
  # Alibaba Cloud OSS configuration
  oss:
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    region: "cn-hangzhou"
    access_key_id: "your-access-key-id"
    access_key_secret: "your-access-key-secret"
    bucket_name: "your-bucket-name"
    workDir: "your-remote-folder"
  
  # Local testing configuration
  localhost:
    workdir: "/path/to/local/storage"
```

## 📖 User Guide

### Starting the Application

Double-click the executable file or run in terminal:

```bash
./fers
```

### Main Features

#### 🔒 **Encrypt & Upload**

1. Select files or folders to upload in the file list
2. Click the **"Encrypt & Upload"** button
3. Files will be encrypted and uploaded to remote storage

#### 📥 **Sync Download**

- Click **"Sync Download"** - Download files that exist remotely but are missing locally
- Click **"Download Specific"** - Select specific remote files to download

#### 📤 **Sync Upload**

- Click **"Sync Upload"** - Upload files that exist locally but are missing remotely

#### 🗂️ **Directory Navigation**

- Click **"Up"** - Return to parent directory
- Select a folder and click **"Enter"** - Enter subdirectory
- All operations are restricted within the configured working directory

#### 🗑️ **File Management**

- Select a file and click **"Delete Local File"** - Delete local file
- Click **"Refresh"** - Refresh file list

#### ⏹️ **Operation Control**

- Click **"Cancel Operation"** - Cancel ongoing long-running operations

### Interface Description

```
┌─────────────────────────────────────────────────────────────┐
│ Working dir: /your/working/directory                        │
│ Current dir: /your/working/directory/subfolder             │
├─────────────────────────────────────────────────────────────┤
│ [Up] [Enter]                                                │
│ ├─ 📁 Documents/                                           │
│ ├─ 📁 Photos/                                              │
│ ├─ 📄 important.txt                                        │
│ └─ 📄 secret.pdf                                           │
├─────────────────────────────────────────────────────────────┤
│ Application Logs                                            │
│ 2024-01-01 10:00:00 INFO Application started successfully  │
│ 2024-01-01 10:01:00 INFO File encrypted and uploaded       │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Advanced Configuration

### Storage Backend Switching

#### Using Alibaba Cloud OSS

```yaml
storage:
  remote_type: "oss"
  oss:
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    # ... other OSS configurations
```

#### Using Local Storage (Testing)

```yaml
storage:
  remote_type: "localhost"
  localhost:
    workdir: "/path/to/local/storage"
```

### Security Recommendations

1. **Key Management**
   - Use strong passwords containing uppercase, lowercase letters, numbers, and special characters
   - Properly backup keys; files cannot be recovered if keys are lost

2. **Access Control**
   - Regularly rotate cloud storage access keys
   - Configure OSS permissions using the principle of least privilege
   - Monitor cloud storage access logs

3. **Data Backup**
   - Regularly backup important encrypted files
   - Store key backups in multiple locations
   - Test the effectiveness of recovery procedures

## 🛠️ Developer Information

### Technical Architecture

- **Programming Language**: Go 1.24.6
- **GUI Framework**: Fyne v2.6.3
- **Encryption Algorithm**: AES-256-GCM
- **Cloud Storage**: Alibaba Cloud OSS SDK v2
- **Configuration Management**: Viper
- **Logging System**: slog

### Project Structure

```
fers/
├── main.go                 # Application entry point
├── pkg/
│   ├── config/            # Configuration management
│   ├── crypto/            # Encryption/decryption
│   ├── storage/           # Storage abstraction layer
│   ├── dir/               # File management
│   └── appui/             # User interface
├── config.yaml            # Configuration file
└── README.md              # Project documentation
```

### Extension Development

The project uses modular design and supports the following extensions:

- **New Storage Backends**: Implement the `storage.Client` interface
- **New Encryption Algorithms**: Implement the `crypto.Cipher` interface
- **New UI Frameworks**: Re-implement the UI layer

## 🐛 Troubleshooting

### Common Issues

**Q: Application startup fails with configuration file not found error**
A: Ensure the `.fers/config.yaml` file is located in the user's home directory or in the current working directory.

**Q: OSS connection failed**
A: Check network connection, verify OSS configuration information is correct, and ensure access keys are valid.

**Q: File encryption failed**
A: Check file permissions, disk space, and verify encryption key configuration is correct.

**Q: Cannot enter certain directory**
A: Ensure the directory is within the configured working directory scope; the application does not allow access to files outside the working directory.

### Log Analysis

The application logs detailed information in the configured log file:

- **INFO** level: Normal operation records
- **ERROR** level: Error information and stack traces
- **DEBUG** level: Detailed debugging information

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Issues and Pull Requests are welcome!

1. Fork this project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📞 Support

If you encounter problems during use, you can get help through the following methods:

- 📧 Submit an [Issue](https://github.com/mingregister/fers/issues)
- 📖 Check [Project Documentation](PROJECT_OVERVIEW.md)
- 💬 Participate in [Discussions](https://github.com/mingregister/fers/discussions)

---

<div align="center">

**⭐ If this project helps you, please give us a Star!**

Made with ❤️ by [mingregister](https://github.com/mingregister)

</div>
