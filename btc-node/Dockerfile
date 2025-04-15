FROM debian:bullseye-slim

# Download btc code
RUN apt-get update && apt-get install -y wget gpg && rm -rf /var/lib/apt/lists/*

RUN wget https://bitcoincore.org/bin/bitcoin-core-28.1/bitcoin-28.1-aarch64-linux-gnu.tar.gz \
    && wget https://bitcoincore.org/bin/bitcoin-core-28.1/SHA256SUMS \
    && wget https://bitcoincore.org/bin/bitcoin-core-28.1/SHA256SUMS.asc

# Install Bitcoin Core
RUN tar -xzf bitcoin-28.1-aarch64-linux-gnu.tar.gz \
    && mv bitcoin-28.1/bin/* /usr/local/bin/ \
    && rm -rf bitcoin-28.1* SHA256SUMS*

VOLUME ["/root/.bitcoin"]
EXPOSE 8332 8333

CMD ["bitcoind"]
VOLUME ["/root/.bitcoin"]
EXPOSE 8332 8333
