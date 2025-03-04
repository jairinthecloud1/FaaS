# Use the official lightweight Node.js 22.14 image.
# https://hub.docker.com/_/node
FROM node:22.14.0-slim

# Create and change to the app directory.
WORKDIR /usr/src/app

COPY package*.json ./

RUN npm install --only=production
# Copy local code to the container image.
COPY . /usr/src/app

# Run the web service on container startup.
CMD [ "npm", "start" ]
