# Setting Up GCC on Windows for WebP Support

This guide will help you set up GCC on Windows, which is required for the `chai2010/webp` package used in the GreenVue.eu backend.

## Why GCC is Required

The WebP image format support in our application relies on the `chai2010/webp` Go package, which uses CGO (Go's foreign function interface) to call C libraries for optimal performance. This approach provides:

- High-quality WebP conversion
- Better performance than pure Go alternatives
- Mature and well-tested WebP encoding

## Installation Options

You have three main options for setting up GCC on Windows:

### Option 1: Install MSYS2 (Recommended)

MSYS2 provides an easy way to install GCC and other Unix-like tools on Windows.

1. **Download and install MSYS2**:

   - Go to https://www.msys2.org/
   - Download the installer for your system (typically x86_64)
   - Run the installer and follow the instructions

2. **Install GCC and related tools**:

   - Open "MSYS2 MINGW64" from the Start menu
   - Run the following command:
     ```bash
     pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-libwebp
     ```

3. **Add MinGW to your PATH**:

   - Right-click on "This PC" or "My Computer" and select "Properties"
   - Click on "Advanced system settings"
   - Click on "Environment Variables"
   - Find the "Path" variable under "System variables" and click "Edit"
   - Add the following path (may vary depending on your installation):
     ```
     C:\msys64\mingw64\bin
     ```
   - Click "OK" to save changes

4. **Verify the installation**:
   - Open a new Command Prompt or PowerShell
   - Run:
     ```
     gcc --version
     ```
   - This should display the GCC version information

### Option 2: Install TDM-GCC

TDM-GCC is another easy option for Windows that focuses specifically on GCC.

1. **Download TDM-GCC**:

   - Go to https://jmeubank.github.io/tdm-gcc/
   - Download the installer for the latest version

2. **Install TDM-GCC**:

   - Run the installer
   - Select "Create" when asked about the installation type
   - Follow the rest of the installation with default options
   - Make sure "Add to PATH" is checked

3. **Verify the installation**:
   - Open a new Command Prompt or PowerShell
   - Run:
     ```
     gcc --version
     ```
   - This should display the GCC version information

### Option 3: Use Docker (Easiest)

If you're having trouble with GCC setup, you can use Docker instead:

1. **Install Docker Desktop**:

   - Download and install from https://www.docker.com/products/docker-desktop

2. **Run the Application in Docker**:
   - From your project root, run:
     ```
     docker-compose up
     ```
   - This will build and run the application in a container that already has GCC and all dependencies

## Running the Application

### For Local Development (with GCC installed)

After setting up GCC, you can run the application with:

```powershell
$env:CGO_ENABLED=1; go run cmd/main.go
```

### Using Docker

For Docker, simply use:

```powershell
docker-compose up
```

## Troubleshooting

### Common Issues

1. **"gcc: command not found"**:

   - Make sure GCC is in your PATH
   - Restart your terminal or command prompt after updating PATH

2. **Missing libwebp**:

   - If you see errors about missing WebP libraries, install them:
     - For MSYS2: `pacman -S mingw-w64-x86_64-libwebp`
     - For TDM-GCC: You may need to install libwebp separately

3. **CGO_ENABLED issues**:
   - Always set CGO_ENABLED=1 before running:
     ```powershell
     $env:CGO_ENABLED=1; go run cmd/main.go
     ```

### Additional Resources

- WebP project page: https://developers.google.com/speed/webp
- chai2010/webp Go package: https://github.com/chai2010/webp
- MSYS2 package search: https://packages.msys2.org/
