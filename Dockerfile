# Dockerfile for Plonk integration testing
# Provides Ubuntu environment with Homebrew, NPM, and system package managers

FROM ubuntu:22.04

# Avoid interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    sudo \
    nodejs \
    npm \
    ca-certificates \
    build-essential \
    procps \
    && rm -rf /var/lib/apt/lists/*

# Create test user with sudo access
RUN useradd -m -s /bin/bash -G sudo testuser
RUN echo 'testuser ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

# Create workspace directory for mounting source code (as root)
RUN mkdir -p /workspace
RUN chown testuser:testuser /workspace

# Switch to test user for Homebrew installation
USER testuser
WORKDIR /home/testuser

# Install Homebrew (non-interactive)
RUN /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Add Homebrew to PATH
ENV PATH="/home/linuxbrew/.linuxbrew/bin:${PATH}"

# Set up testing environment
ENV HOME=/home/testuser
ENV PLONK_TEST_MODE=true

WORKDIR /workspace

# Default command for interactive debugging
CMD ["/bin/bash"]