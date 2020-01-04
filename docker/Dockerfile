FROM node:lts
WORKDIR /work
RUN wget https://github.com/gohugoio/hugo/releases/download/v0.62.1/hugo_extended_0.62.1_Linux-64bit.tar.gz && tar xf *.tar.gz && rm *.tar.gz
RUN wget https://dl.google.com/go/go1.13.5.linux-amd64.tar.gz && tar xf *.tar.gz && rm *.tar.gz
ADD package.json .
RUN npm install