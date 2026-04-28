#!/bin/bash

# Redirect output to log for debugging
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1

echo "Provisioning started..."

# 1. Setup 2GB Swap (for 1GB RAM stability)
fallocate -l 2G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab

# 2. Install Docker, Compose and Git
export DEBIAN_FRONTEND=noninteractive
apt-get update && apt-get upgrade -y
apt-get install -y docker.io docker-compose-v2 git ufw iptables-persistent

# 3. Open Ports (80, 443, 81) in iptables
iptables -I INPUT 6 -m state --state NEW -p tcp --dport 80 -j ACCEPT
iptables -I INPUT 6 -m state --state NEW -p tcp --dport 443 -j ACCEPT
iptables -I INPUT 6 -m state --state NEW -p tcp --dport 81 -j ACCEPT
netfilter-persistent save

# 4. Create devops user and grant permissions
useradd -m -s /bin/bash devops
usermod -aG sudo,docker devops
echo "devops ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/90-devops

# 5. Setup SSH access for devops
mkdir -p /home/devops/.ssh
cp /home/ubuntu/.ssh/authorized_keys /home/devops/.ssh/
chown -R devops:devops /home/devops/.ssh
chmod 700 /home/devops/.ssh
chmod 600 /home/devops/.ssh/authorized_keys

# 6. Create deployment directories
mkdir -p /home/devops/deploy/{app,monitoring,npm}
chown -R devops:devops /home/devops/deploy

echo "Provisioning finished."