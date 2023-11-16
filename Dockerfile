FROM debian:latest

# RUN echo "deb http://ftp.debian.org/debian experimental main" | tee /etc/apt/sources.list

# add a user, update to latest glibc
# RUN useradd -m user \
#     && apt-get update \
#     && apt-get -t experimental install -y libc6 libc6-dev libc6-dbg

USER user

# add dependencies (staking-deposit-cli, sw operator cli, and nimbus)
# ADD --chown=user:user https://github.com/nodeset-org/staking-deposit-cli/releases/download/v2.7.0-exit-messages/staking-deposit-cli-linux-amd64 /home/user/bin/deposit
ADD --chown=user:user https://github.com/stakewise/v3-operator/releases/download/v0.3.4/operator-v0.3.4-linux-amd64.tar.gz /home/user/
# ADD --chown=user:user https://github.com/status-im/nimbus-eth2/releases/download/v23.10.0/nimbus-eth2_Linux_amd64_23.10.0_8b07f4fd.tar.gz /home/user/bin/nimbus.tar.gz

WORKDIR /home/user

# RUN chmod +x deposit

# extract the stakewise operator binary
RUN tar -xf operator-v0.3.4-linux-amd64.tar.gz \
    && cp operator-v0.3.4-linux-amd64/operator operator \
    && rm -dr operator-v0.3.4-linux-amd64* \
    && chmod +x operator

# extract nimbus
# RUN tar -xf nimbus.tar.gz \
#     && mv nimbus-eth2_Linux_amd64_23.10.0_8b07f4fd nimbus \
#     && rm nimbus.tar.gz

# eth deposit tool requires these environment vars
# ENV LANG C.UTF-8
# ENV LC_ALL C.UTF-8