## Patroni dev image for scion.
## Based on dockerfile from https://github.com/zalando/patroni.
## Instead of etcd we use consul and therefore remove all the etc stuff.
FROM postgres:10

RUN export DEBIAN_FRONTEND=noninteractive \
    && echo 'APT::Install-Recommends "0";\nAPT::Install-Suggests "0";' > /etc/apt/apt.conf.d/01norecommend \
    && apt-get update -y \
    && apt-get upgrade -y \
    # postgres:10 is based on debian, which has patroni package. We will install all required dependencies
    && apt-get install -s patroni | sed -n -e '/^Inst patroni /d' -e 's/^Inst \([^ ]\+\) .*$/\1/p' \
            | xargs apt-get install -y curl jq locales python3-pip \
    ## Make sure we have a en_US.UTF-8 locale available
    && localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8 \
    # Prepare postgres folder
    && mkdir -p /home/postgres \
    && chown postgres:postgres /home/postgres

RUN pip3 install python-consul

# Add patroni code
ENV PATRONIVERSION 1.5.4
RUN mkdir /patronisrc
RUN curl https://codeload.github.com/zalando/patroni/tar.gz/v${PATRONIVERSION} \
    | tar -xz -C /patronisrc
RUN mv /patronisrc/patroni-${PATRONIVERSION}/patroni*.py / \
 && mv /patronisrc/patroni-${PATRONIVERSION}/patroni /patroni/

ADD entrypoint.sh /

# Clean up
RUN apt-get purge -y libpython2.7-stdlib libpython2.7-minimal python3-pip \
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/* /root/.cache \
    && rm -rf /patronisrc

RUN sed -i 's/env python/&3/' patroni*.py && ln -s /patronictl.py /usr/local/bin/patronictl && mkdir /data/ /cfg/ /initsql/ \
    && touch /pgpass /patroni.yml && chown postgres:postgres -R /patroni/ /cfg/ /data/ /initsql/ /pgpass /patroni.yml /var/run/ /var/lib/ /var/log/

EXPOSE 2379 5432 8008

ENV LC_ALL=en_US.UTF-8 LANG=en_US.UTF-8
ENTRYPOINT ["/bin/bash", "/entrypoint.sh"]
USER postgres
