# use official node image
FROM node:20

RUN mkdir -p /app

# set working directory
WORKDIR /app

# copy project file
COPY . .

RUN npm install
RUN npm run build

EXPOSE 3000

# start app
CMD ["npm", "start"]
