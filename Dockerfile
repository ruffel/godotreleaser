FROM ubuntu:24.04

#-------------------------------------------------------------------------------
# Install system dependencies
#-------------------------------------------------------------------------------
RUN true \
    && export DEBIAN_FRONTEND=noninteractive \
    && apt-get update --no-install-recommends \
    && apt-get install -y \
        ca-certificates \
        git \
        libfontconfig1 \
        unzip \
        wget

#-------------------------------------------------------------------------------
# Install Godot
#-------------------------------------------------------------------------------
ENV GODOT_VERSION="4.2.2"

#-------------------------------------------------------------------------------
# Configure the startup environment.
#-------------------------------------------------------------------------------
COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY godotreleaser /usr/local/bin/godotreleaser
RUN chmod +x /usr/local/bin/godotreleaser

#-------------------------------------------------------------------------------
# Preload Godot dependencies to speed up the build process.
#-------------------------------------------------------------------------------
RUN /usr/local/bin/godotreleaser dependencies --version $GODOT_VERSION

ENTRYPOINT [ "/entrypoint.sh" ]
