FROM node:20-alpine as build-stage

WORKDIR /src

COPY package.json .
COPY package-lock.json .

RUN npm install

COPY . .

RUN npm run build

FROM node:20-alpine

WORKDIR /app

COPY --from=build-stage /src/dist /app/dist

COPY package.json .
COPY package-lock.json .

RUN npm install --production

EXPOSE 3000

CMD [ "node", "dist/analog/server/index.mjs" ]
