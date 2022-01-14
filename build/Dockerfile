FROM ubuntu:18.04 as build

RUN useradd -d /home/hwMindX -u 9000 -m -s  /usr/sbin/nologin hwMindX

COPY ./hccl-controller  /usr/local/bin/

RUN chown -R hwMindX:hwMindX /usr/local/bin/  &&\
    chmod 550 /usr/local/bin/ &&\
    chmod 750 /home/hwMindX &&\
    chmod 500 /usr/local/bin/hccl-controller &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> /home/hwMindX/.bashrc

# 使用hwMindX作为容器默认用户
USER hwMindX

