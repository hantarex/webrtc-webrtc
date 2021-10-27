FROM restreamio/gstreamer:1.19.2.0-prod
ENV TZ=Europe/Moscow
ENV GIT_SSL_NO_VERIFY=1
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
#FROM restreamio/gstreamer:2021-07-15T13-58-52Z-dev-with-source
RUN apt-get update && apt-get -y install libjson-glib-1.0-0 libjson-glib-dev wget libunwind-dev libdw-dev libsoup2.4-dev build-essential liborc-0.4-dev
RUN cd /tmp && wget https://golang.org/dl/go1.16.6.linux-amd64.tar.gz && rm -rf /usr/local/go && tar -C /usr/local -xzf go1.16.6.linux-amd64.tar.gz
ENV PATH $PATH:/usr/local/go/bin

ENTRYPOINT ["tail", "-f", "/dev/null"]
