FROM node:20
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
# Install Playwright browser binaries and OS-level dependencies for e2e tests.
RUN npx playwright install --with-deps chromium
COPY frontend/ ./
EXPOSE 5173
CMD ["npm", "run", "dev"]
