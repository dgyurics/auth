# Builder stage
FROM node:20 AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Final stage
FROM node:20

WORKDIR /app

COPY --from=builder /app/package*.json ./
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
ENV NODE_ENV=production

EXPOSE 3000

CMD ["npm", "start"]
