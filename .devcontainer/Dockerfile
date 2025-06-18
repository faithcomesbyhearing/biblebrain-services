# Use the dev-prefixed tag so the FROM actually exists
ARG VARIANT="1.24-bullseye"
FROM mcr.microsoft.com/vscode/devcontainers/go:dev-${VARIANT}

ARG OKTA_VERSION=0.10.0
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID
ARG NODE_VERSION="lts/*"

# 1) Install sudo, curl, unzip; clean apt cache
RUN apt-get update \
  && apt-get install -y --no-install-recommends sudo curl unzip \
  && rm -rf /var/lib/apt/lists/* \
  # add sudoers entry for vscode
  && echo "$USERNAME ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME \
  && chmod 0440 /etc/sudoers.d/$USERNAME

# 2) Sync UID/GID for non-root user
RUN groupmod --gid $USER_GID $USERNAME \
  && usermod --uid $USER_UID --gid $USER_GID $USERNAME \
  && chown -R $USER_UID:$USER_GID /home/$USERNAME

# 3) Install Okta CLI
RUN curl -fsSL \
    "https://github.com/okta/okta-cli/releases/download/okta-cli-tools-${OKTA_VERSION}/okta-cli-linux-${OKTA_VERSION}-x86_64.zip" \
  | funzip > /usr/local/bin/okta \
  && chmod +x /usr/local/bin/okta

# 4) Install Node via NVM and global packages
RUN su $USERNAME -c ". /usr/local/share/nvm/nvm.sh && \
    nvm install ${NODE_VERSION} && \
    npm install -g yarn serverless"

# 5) Copy any local binaries/scripts
COPY awslambdarpc /usr/local/bin/
COPY postCreateCommand.sh /scripts/
RUN chmod +x /scripts/postCreateCommand.sh

# 6) Set working directory
WORKDIR /workspace
USER $USERNAME

# (Optional) Run your post-create hook
CMD ["/scripts/postCreateCommand.sh"]
