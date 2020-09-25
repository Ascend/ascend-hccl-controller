FROM ubuntu:18.04 as build

RUN useradd -d /home/hwMindX -u 9000 -m -s /bin/bash hwMindX && \
    useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser && \
    usermod -a -G HwHiAiUser hwMindX

COPY ./hccl-controller  /usr/local/bin/

RUN chown -R hwMindX:hwMindX /usr/local/bin/hccl-controller  &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> ~/.bashrc

# 使用hwMindX作为容器默认用户
USER hwMindX

