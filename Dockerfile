FROM fedora:34
ENV GOPATH=/go
ENV PATH=$PATH:/go/bin
RUN echo fuk
RUN dnf install -y git make gcc gcc-c++ which iproute iputils procps-ng vim-minimal tmux net-tools htop tar jq npm openssl-devel perl rust cargo golang

RUN PATH="$HOME/.cargo/bin:$PATH" && cargo install ibc-relayer-cli --bin hermes --locked

# Copy cosmos-sdk to debug something
ADD ./cosmos-sdk /cosmos-sdk

# Copy in the repo under test
ADD ./interchain-security /interchain-security

# Build the Go module
RUN pushd /interchain-security/ && PATH=$PATH:/usr/local/go/bin GOPROXY=https://proxy.golang.org make && PATH=$PATH:/usr/local/go/bin make install

# Copy in the shell scripts that run the testnet
ADD ./testnet-scripts /testnet-scripts

# Copy in the hermes config
ADD ./testnet-scripts/hermes-config.toml /root/.hermes/config.toml