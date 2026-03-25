# Astiav Backend (FFmpeg 8.0 via libavformat)

This document describes how to build and test discoteca with the **astiav** backend, which uses FFmpeg's libavformat directly via CGO for faster media probing.

## Overview

The astiav backend provides:
- **5-10x faster** metadata extraction compared to ffprobe
- No process spawn overhead for each file
- Direct C library calls to libavformat
- **FFmpeg 8.0.1** from RPM Fusion (Fedora Rawhide)

**Requirements:**
- Podman (for containerized builds)
- FFmpeg 8.0 development libraries (if building natively)
- CGO enabled

## Quick Start

### Option 1: Use pre-built FFmpeg from packages (recommended)

**Alpine Linux (edge)** - Smallest image, musl libc:
```bash
make astiav-image-alpine
make astiav-build-alpine
```

**Arch Linux** - Latest FFmpeg version:
```bash
make astiav-image-arch
make astiav-build-arch
```

**Fedora Rawhide** - FFmpeg 8.0 from RPM Fusion:
```bash
make astiav-image
make astiav-build
```

### Option 2: Build FFmpeg from source (for static linking)

```bash
make astiav-build-static
```

This builds FFmpeg 8.0 from source with `--enable-static --disable-shared`, then builds discoteca with static linking. 

**Note:** 
- Build time: ~15-25 minutes
- Binary size: ~100-150MB
- Result: Fully standalone binary with no external dependencies

## Usage

Once built, use the `--probe-backend=astiav` flag:

```bash
# Run in container (recommended)
podman run --rm -v $(pwd):/src:z -w /src disco-astiav:latest \
    ./disco-astiav add --probe-backend=astiav my.db /path/to/media

# Or copy binary and run with library path
export LD_LIBRARY_PATH=/path/to/ffmpeg8/lib:$LD_LIBRARY_PATH
./disco-astiav add --probe-backend=astiav my.db /path/to/media
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make astiav-image` | Build the container image with FFmpeg 8.0 |
| `make astiav-build` | Build discoteca with astiav backend |
| `make astiav-test` | Run astiav backend tests |
| `make astiav-shell` | Open interactive shell in container |
| `make astiav-quicktest` | Quick single-test validation |
| `make astiav-build-native` | Native build (requires FFmpeg 8.0 on host) |

## Container Options

| Container | Base | FFmpeg Source | Size | Best For |
|-----------|------|---------------|------|----------|
| `disco-astiav:latest` | Fedora Rawhide | RPM Fusion (8.0) | ~2GB | Development, testing |
| `disco-astiav:alpine` | Alpine edge | apk packages | ~900MB | Smallest image |
| `disco-astiav:arch` | Arch Linux | pacman packages | ~1.5GB | Latest features |
| Static build | Fedora Rawhide | Built from source | ~100MB binary | Deployment, portability |

See `Containerfile.astiav*` for full container definitions.

**Note on static linking:** Most Linux distributions don't ship static FFmpeg libraries due to security concerns (static binaries can't be easily patched). For a truly standalone binary, use `make astiav-build-static` which builds FFmpeg from source.

## Performance Comparison

| Backend | Files/sec | Notes |
|---------|-----------|-------|
| ffprobe (default) | ~10-20 | Portable, no CGO |
| astiav | ~50-100+ | Requires FFmpeg 8.0 dev libs |

*Actual performance depends on hardware, I/O, and parallelism settings.*

## Testing

The astiav backend includes comprehensive tests:

```bash
# Run all probe tests
make astiav-test

# Test output:
# === RUN   TestFFProbeBackend_RealMedia
# --- PASS: TestFFProbeBackend_RealMedia (1.65s)
# === RUN   TestFFProbeBackend_RealAudio
# --- PASS: TestFFProbeBackend_RealAudio (0.37s)
# === RUN   TestFFProbeBackend_RealMKVWithChapters
# --- PASS: TestFFProbeBackend_RealMKVWithChapters (1.65s)
```

Tests verify:
- ✅ Format name detection
- ✅ Duration extraction
- ✅ Metadata tags (title, artist, album, genre, etc.)
- ✅ Video stream properties (codec, resolution, fps)
- ✅ Audio stream properties (codec, sample rate, channels)
- ✅ Stream disposition (attached_pic filtering)

## Troubleshooting

### Binary requires shared libraries

The built binary links against FFmpeg 8.0 shared libraries. Run it in the container:

```bash
podman run --rm -v $(pwd):/src:z -w /src disco-astiav:latest \
    ./disco-astiav --help
```

Or copy the libraries to your system (not recommended).

### Container build fails

Check internet connectivity and RPM Fusion availability:

```bash
curl -I https://download1.rpmfusion.org/free/fedora/rpmfusion-free-release-rawhide.noarch.rpm
```

### Astiav backend not available

The astiav backend requires building with the `astiav` tag:

```bash
go build -tags "fts5 astiav" ./cmd/disco
```

## Alternative: Build FFmpeg 8.0 from Source

If you prefer not to use containers:

```bash
# Install build dependencies
sudo dnf install gcc gcc-c++ make yasm nasm \
    libX11-devel zlib-devel openssl-devel gnutls-devel \
    lame-devel x264-devel x265-devel libvpx-devel

# Build FFmpeg 8.0
cd /tmp
git clone https://git.ffmpeg.org/ffmpeg.git
cd ffmpeg
git checkout n8.0
./configure --prefix=/usr/local --enable-shared --enable-gpl
make -j$(nproc)
sudo make install
sudo ldconfig

# Build discoteca
CGO_CFLAGS="-I/usr/local/include" CGO_LDFLAGS="-L/usr/local/lib" \
    go build -tags "fts5 astiav" ./cmd/disco
```

## Implementation Notes

The astiav backend uses:
- **go-astiav v0.40.0** - Go bindings for FFmpeg 8.0
- **Dictionary API** - For metadata tag extraction
- **MediaType API** - For stream type detection
- **DispositionFlags** - For filtering attached pictures

Chapters support is limited in astiav v0.40 (requires direct CGO access to AVFormatContext.chapters).
