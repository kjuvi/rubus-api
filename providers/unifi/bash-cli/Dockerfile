FROM ubuntu:latest

COPY . /provider

WORKDIR /provider

RUN apt-get update && apt-get install -y expect netcat openssh-client

# add the Unifi Network Switch to the known hosts
RUN mkdir -p /root/.ssh 
RUN ssh-keyscan 192.168.0.103 >> /root/.ssh/known_hosts
RUN sed -i '1s/^/poe-1.cluster.iiun.unine.ch,/' /root/.ssh/known_hosts

CMD [ "./provider.sh" ]
