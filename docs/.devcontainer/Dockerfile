ARG VARIANT=latest
FROM oven/bun:${VARIANT}

RUN DEBIAN_FRONTEND=noninteractive apt-get update \
    && apt-get -y install --no-install-recommends \
    ca-certificates \
    git \
    nano \
    vim-tiny \
    && apt-get autoremove -y \
    && apt-get clean -y 

# Switch to bun user
USER bun

# Add bun to path
ENV PATH=~/.bin/bin:${PATH}

# [Optional] Uncomment this section to install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && apt-get -y install --no-install-recommends <your-package-list-here>

# [Optional] Uncomment if you want to install global node modules
# RUN bun install -g <your-package-list-here>