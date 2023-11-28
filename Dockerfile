FROM debian:latest

# RUN echo "deb http://ftp.debian.org/debian experimental main" | tee /etc/apt/sources.list

# add a user, update to latest glibc
RUN useradd -m user
#     && apt-get update \
#     && apt-get -t experimental install -y libc6 libc6-dev libc6-dbg


# add dependencies (staking-deposit-cli, sw operator cli)
# ADD --chown=user:user https://github.com/nodeset-org/staking-deposit-cli/releases/download/v2.7.0-exit-messages/staking-deposit-cli-linux-amd64 /home/user/bin/deposit
# ADD --chown=user:user https://github.com/stakewise/v3-operator/releases/download/v1.0.0/operator-v1.0.0-linux-amd64.tar.gz /home/user/
ADD --chown=user:user operator-v1.0.0-linux-amd64.tar.gz /home/user/

WORKDIR /home/user

# RUN chmod +x deposit

# extract the stakewise operator binary
# RUN tar -xf operator-v1.0.0-linux-amd64.tar.gz \
#     && cp operator-v1.0.0-linux-amd64/operator operator \
#     && rm -dr operator-v1.0.0-linux-amd64* \
#     && chmod +x operator

RUN cp operator-v1.0.0-linux-amd64/operator operator \
    && chown user operator-v1.0.0-linux-amd64/operator \
    && rm -dr operator-v1.0.0-linux-amd64* \
    && chmod +x operator

USER user
ENTRYPOINT ["./operator"]

# eth deposit tool requires these environment vars
# ENV LANG C.UTF-8
# ENV LC_ALL C.UTF-8