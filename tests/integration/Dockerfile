FROM ubuntu:22.04

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive
ENV PLONK_DIR=/home/testuser/dotfiles

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    software-properties-common \
    python3 \
    python3-pip \
    ruby \
    ruby-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js and npm
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs

# Install Rust and Cargo
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

# Install Go (architecture-aware)
RUN ARCH=$(dpkg --print-architecture) && \
    if [ "$ARCH" = "arm64" ]; then \
        wget -O go.tar.gz https://go.dev/dl/go1.22.0.linux-arm64.tar.gz; \
    else \
        wget -O go.tar.gz https://go.dev/dl/go1.22.0.linux-amd64.tar.gz; \
    fi && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

# Install Homebrew
RUN /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
ENV PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:${PATH}"

# Create test user to avoid running as root
RUN useradd -m -s /bin/bash testuser
USER testuser
WORKDIR /home/testuser

# Set up environment for test user
ENV PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:/usr/local/go/bin:/home/testuser/.cargo/bin:${PATH}"
ENV PLONK_DIR=/home/testuser/dotfiles

# Create directories that tests will use
RUN mkdir -p $PLONK_DIR

# Copy plonk binary (will be mounted from host)
# The binary will be available at /usr/local/bin/plonk

CMD ["/bin/bash"]
