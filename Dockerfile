# Plonk BATS Test Container
# This container includes all supported package managers for comprehensive testing
#
# Supported package managers:
#   - brew (Homebrew)
#   - npm (Node.js)
#   - pnpm (Fast npm alternative)
#   - cargo (Rust)
#   - pipx (Python applications)
#   - conda (Scientific computing)
#   - gem (Ruby)
#   - uv (Fast Python tool manager)
#
# Build: docker build -t plonk-test .
# Run:   docker run --rm plonk-test

FROM ubuntu:24.04

# Avoid interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# Install base dependencies
RUN apt-get update && apt-get install -y \
    # Build essentials
    build-essential \
    curl \
    wget \
    git \
    # Required for Homebrew
    procps \
    file \
    # Required for BATS
    bats \
    # Ruby
    ruby-full \
    # Python
    python3 \
    python3-pip \
    python3-venv \
    # Misc utilities
    sudo \
    locales \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set up locale
RUN locale-gen en_US.UTF-8
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8

# Create non-root user for Homebrew and testing
# Homebrew refuses to run as root
RUN useradd -m -s /bin/bash plonk && \
    echo "plonk ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Switch to plonk user for remaining setup
USER plonk
WORKDIR /home/plonk

# Set up environment variables
ENV HOME=/home/plonk
ENV PATH="${HOME}/.local/bin:${HOME}/go/bin:${HOME}/.cargo/bin:${HOME}/.linuxbrew/bin:${HOME}/.linuxbrew/sbin:/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:${PATH}"

# Install Homebrew
RUN /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" && \
    echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> ~/.bashrc && \
    eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)" && \
    brew --version

# Update PATH for Homebrew (ensuring it's available in subsequent RUN commands)
ENV PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:${PATH}"
ENV HOMEBREW_NO_AUTO_UPDATE=1
ENV HOMEBREW_NO_INSTALL_CLEANUP=1

# Install Go via Homebrew
RUN brew install go && go version

# Set Go environment
ENV GOPATH="${HOME}/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# Install Node.js and npm via Homebrew
RUN brew install node && \
    node --version && \
    npm --version

# Install pnpm and set up global bin directory
ENV PNPM_HOME="${HOME}/.local/share/pnpm"
ENV PATH="${PNPM_HOME}:${PATH}"
RUN npm install -g pnpm && \
    mkdir -p "${PNPM_HOME}" && \
    pnpm config set global-bin-dir "${PNPM_HOME}" && \
    pnpm --version

# Install Rust and Cargo
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y && \
    . "${HOME}/.cargo/env" && \
    cargo --version

# Ensure cargo is in PATH
ENV PATH="${HOME}/.cargo/bin:${PATH}"

# Install uv (fast Python package manager)
RUN curl -LsSf https://astral.sh/uv/install.sh | sh && \
    ~/.local/bin/uv --version

# Ensure uv is in PATH
ENV PATH="${HOME}/.local/bin:${PATH}"

# Install pipx
RUN python3 -m pip install --user --break-system-packages pipx && \
    python3 -m pipx ensurepath && \
    ~/.local/bin/pipx --version

# Install Miniconda for conda support (architecture-aware)
ARG TARGETARCH
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "aarch64" || echo "x86_64") && \
    curl -fsSL "https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-${ARCH}.sh" -o miniconda.sh && \
    bash miniconda.sh -b -p "${HOME}/miniconda3" && \
    rm miniconda.sh && \
    ~/miniconda3/bin/conda init bash && \
    ~/miniconda3/bin/conda --version

# Add conda to PATH
ENV PATH="${HOME}/miniconda3/bin:${PATH}"

# Accept conda Terms of Service for default channels (required for non-interactive use)
RUN conda tos accept --override-channels --channel https://repo.anaconda.com/pkgs/main && \
    conda tos accept --override-channels --channel https://repo.anaconda.com/pkgs/r

# Install BATS support libraries via Homebrew
RUN brew install bats-core && \
    brew tap bats-core/bats-core && \
    brew install bats-support bats-assert

# Create the bats library directories that test_helper.bash expects
RUN sudo mkdir -p /usr/local/lib && \
    sudo ln -sf /home/linuxbrew/.linuxbrew/lib/bats-support /usr/local/lib/bats-support && \
    sudo ln -sf /home/linuxbrew/.linuxbrew/lib/bats-assert /usr/local/lib/bats-assert

# Create working directory for plonk
WORKDIR /home/plonk/plonk

# Copy go.mod and go.sum first for better layer caching
COPY --chown=plonk:plonk go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy the rest of the source code
COPY --chown=plonk:plonk . .

# Build plonk binary and install to a location in PATH
RUN go build -o /tmp/plonk-bats/plonk ./cmd/plonk && \
    chmod +x /tmp/plonk-bats/plonk

# Add plonk to PATH (at the beginning so it's found first)
ENV PATH="/tmp/plonk-bats:${PATH}"

# Create plonk config directory
RUN mkdir -p ~/.config/plonk

# Set test environment variables
ENV PLONK_TEST_CLEANUP_PACKAGES=1
ENV PLONK_TEST_CLEANUP_DOTFILES=1

# Make entrypoint script executable
RUN chmod +x scripts/docker-entrypoint.sh

# Set entrypoint
ENTRYPOINT ["scripts/docker-entrypoint.sh"]

# Default command runs all behavioral tests
CMD ["all"]
