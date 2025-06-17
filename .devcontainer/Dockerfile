# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.231.6/containers/go/.devcontainer/base.Dockerfile

# [Choice] Go version (use -bullseye variants on local arm64/Apple Silicon): 1, 1.16, 1.17, 1-bullseye, 1.16-bullseye, 1.17-bullseye, 1-buster, 1.16-buster, 1.17-buster
ARG VARIANT="1.24-bullseye"
FROM mcr.microsoft.com/vscode/devcontainers/go:1-${VARIANT}

# Install Okta CLI
ARG OKTA_VERSION=0.10.0
RUN curl -fsSL "https://github.com/okta/okta-cli/releases/download/okta-cli-tools-${OKTA_VERSION}/okta-cli-linux-${OKTA_VERSION}-x86_64.zip" | funzip > /usr/local/bin/okta
RUN chmod +x /usr/local/bin/okta

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# [Optional] Add sudo support. Omit if you don't need to install software after connecting.
RUN apt-get update \
    && apt-get install -y sudo \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

RUN groupmod --gid $USER_GID $USERNAME \
    && usermod --uid $USER_UID --gid $USER_GID $USERNAME \
    && chown -R $USER_UID:$USER_GID /home/$USERNAME

# [Choice] Node.js version: none, lts/*, 16, 14, 12, 10
ARG NODE_VERSION="none"
RUN if [ "${NODE_VERSION}" != "none" ]; then su vscode -c "umask 0002 && . /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

RUN yarn global add serverless

COPY awslambdarpc /usr/local/bin

RUN apt-get install -y

RUN mkdir -p /scripts
COPY postCreateCommand.sh /scripts
WORKDIR /scripts
RUN chmod +x postCreateCommand.sh

RUN chmod -R 777 /home/vscode

