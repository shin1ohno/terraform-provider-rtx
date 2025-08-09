FROM alpine:latest

# Install OpenSSH server
RUN apk add --no-cache openssh-server

# Create test user
RUN adduser -D -s /bin/sh testuser && \
    echo "testuser:testpass" | chpasswd

# Generate host keys (for testing only)
RUN ssh-keygen -A

# Configure SSH
RUN echo "PermitRootLogin no" >> /etc/ssh/sshd_config && \
    echo "PasswordAuthentication yes" >> /etc/ssh/sshd_config && \
    echo "PubkeyAuthentication yes" >> /etc/ssh/sshd_config && \
    echo "Port 22" >> /etc/ssh/sshd_config

# Create a simple RTX simulator script
RUN mkdir -p /usr/local/bin
COPY rtx-simulator.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/rtx-simulator.sh

# Set the shell for testuser to our simulator
RUN sed -i 's|/bin/sh|/usr/local/bin/rtx-simulator.sh|' /etc/passwd

EXPOSE 22

# Start SSH daemon
CMD ["/usr/sbin/sshd", "-D", "-e"]