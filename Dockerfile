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

RUN true \
    #---------------------------------------------------------------------------
    # Install Godot
    #---------------------------------------------------------------------------
    printf "Installing Godot v%s\n" $GODOT_VERSION \
    && wget -q https://downloads.tuxfamily.org/godotengine/$GODOT_VERSION/Godot_v$GODOT_VERSION-stable_linux.x86_64.zip \
    && unzip Godot_v$GODOT_VERSION-stable_linux.x86_64.zip \
    && mv Godot_v$GODOT_VERSION-stable_linux.x86_64 /usr/local/bin/godot \
    && chmod +x /usr/local/bin/godot \
    && rm Godot_v$GODOT_VERSION-stable_linux.x86_64.zip \
    #---------------------------------------------------------------------------
    # Install Godot export templates
    #---------------------------------------------------------------------------
    && printf "Installing Godot export templates v%s\n" $GODOT_VERSION \
    && wget -q https://downloads.tuxfamily.org/godotengine/$GODOT_VERSION/Godot_v$GODOT_VERSION-stable_export_templates.tpz \
    && mkdir -p ~/.local/share/godot/export_templates/$GODOT_VERSION.stable \
    && unzip -o Godot_v$GODOT_VERSION-stable_export_templates.tpz -d ~/.local/share/godot/export_templates/$GODOT_VERSION.stable \
    && mv ~/.local/share/godot/export_templates/$GODOT_VERSION.stable/templates/* ~/.local/share/godot/export_templates/$GODOT_VERSION.stable/ \
    && rm -r ~/.local/share/godot/export_templates/$GODOT_VERSION.stable/templates \
    && rm Godot_v$GODOT_VERSION-stable_export_templates.tpz

#-------------------------------------------------------------------------------
# Configure the startup environment.
#-------------------------------------------------------------------------------
COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY godotreleaser /usr/local/bin/godotreleaser
RUN chmod +x /usr/local/bin/godotreleaser

ENTRYPOINT [ "/entrypoint.sh" ]
