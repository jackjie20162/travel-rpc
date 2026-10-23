FROM alpine:3.22

# Define the project name | 定义项目名称
ARG PROJECT=travel
# Define the config file name | 定义配置文件名
ARG CONFIG_FILE=travel-rpc.yaml
# Define the author | 定义作者
ARG AUTHOR="610841147@qq.com"

LABEL org.opencontainers.image.authors=${AUTHOR}

WORKDIR /app
ENV PROJECT=${PROJECT}
ENV CONFIG_FILE=${CONFIG_FILE}

COPY ./${PROJECT}_rpc ./
COPY ./etc/${CONFIG_FILE} ./etc/

EXPOSE 9205

ENTRYPOINT ["./travel_rpc", "-f", "etc/travel-rpc.yaml"]