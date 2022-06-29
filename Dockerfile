FROM node:16-alpine3.14 as build-stage

LABEL name "Scheduled Tasks (Docker Build)"
LABEL maintainer "KagChi"

WORKDIR /tmp/build

RUN apk add --no-cache build-base git python3

COPY package.json .

RUN npm install

COPY . .

RUN npm run build

RUN npm prune --production

FROM node:16-alpine3.14

LABEL name "Scheduled Tasks Production"
LABEL maintainer "KagChi"

WORKDIR /app

RUN apk add --no-cache tzdata

COPY --from=build-stage /tmp/build/package.json .
COPY --from=build-stage /tmp/build/package-lock.json .
COPY --from=build-stage /tmp/build/node_modules ./node_modules
COPY --from=build-stage /tmp/build/dist ./dist

CMD node -r dotenv/config dist/index.js