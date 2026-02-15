# 名称：apinto镜像，携带了部署k8s集群所需要的脚本
# 创建时间：2024-11-19
FROM alpine:latest

RUN sed -i 's|https://dl-cdn.alpinelinux.org/alpine|https://mirrors.aliyun.com/alpine|g' /etc/apk/repositories  \
    && apk update \
    && apk add --no-cache bash tzdata

RUN apk add --no-cache netcat-openbsd curl

ENV TZ=Asia/Shanghai

RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone

#设置环境变量
ARG AppName=apinto

#定义数据卷
VOLUME /var/lib/${AppName}

# Create non-root user and group
RUN addgroup -g 1000 ${AppName} && \
    adduser -D -u 1000 -G ${AppName} ${AppName}

#解压网关程序压缩包
COPY ./apinto.linux.x64.tar.gz /
RUN tar -zxvf apinto.linux.x64.tar.gz && ls /${AppName} / && rm -rf ../apinto.linux.x64.tar.gz

#复制程序默认配置文件以及修改脚本权限
RUN mkdir -p /etc/${AppName}
RUN cp /${AppName}/${AppName}.yml.tpl /etc/${AppName}/${AppName}.yml && cp /${AppName}/config.yml.tpl /etc/${AppName}/config.yml

RUN chmod 755 /${AppName}/*.sh

# Change ownership to non-root user
RUN chown -R ${AppName}:${AppName} /${AppName} /etc/${AppName} /var/lib/${AppName}

# Switch to non-root user
USER ${AppName}

WORKDIR /apinto

CMD ["bash","/apinto/auto-start.sh"]