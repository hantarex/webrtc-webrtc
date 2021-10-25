FROM hantarex/gstreamer:2.1p
#FROM restreamio/gstreamer:2021-07-15T13-58-52Z-dev-with-source
RUN apt-get update && apt-get -y install libjson-glib-1.0-0 libjson-glib-dev
RUN cd /tmp && wget https://golang.org/dl/go1.16.6.linux-amd64.tar.gz && rm -rf /usr/local/go && tar -C /usr/local -xzf go1.16.6.linux-amd64.tar.gz
ENV PATH $PATH:/usr/local/go/bin

ENTRYPOINT ["tail", "-f", "/dev/null"]
