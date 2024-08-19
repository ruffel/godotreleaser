FROM ubuntu:24.04

#-------------------------------------------------------------------------------
# Install system dependencies
#-------------------------------------------------------------------------------
RUN true \
    && export DEBIAN_FRONTEND=noninteractive \
    && apt-get update --no-install-recommends \
    && apt-get install -y \
        ca-certificates \
        dotnet-sdk-8.0 \
        git \
        libfontconfig1 \
        unzip \
        wget \
        gnupg \
        dirmngr \
        apt-transport-https

#-------------------------------------------------------------------------------
# Install Godot
#-------------------------------------------------------------------------------
ENV GODOT_VERSION="4.3"

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
RUN true \
    && /usr/local/bin/godotreleaser dependencies --version $GODOT_VERSION \
    && /usr/local/bin/godotreleaser dependencies --version $GODOT_VERSION --with-mono

ENTRYPOINT [ "/entrypoint.sh" ]
