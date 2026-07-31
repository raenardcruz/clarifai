# Setup Guide

This guide explains how to set up the environment and dependencies for both the frontend and backend of the **Note-Taker AI** application.

---

## 🛠️ System Prerequisites

Before setting up the project, ensure you have the following installed on your machine:

1. **Go 1.25+**: The backend is built using Go.
2. **Node.js 18+ & npm**: For the Vite/Vue 3 frontend.
3. **FFmpeg**: Required for audio file duration probing.
   - **macOS**: `brew install ffmpeg`
   - **Linux**: `sudo apt install ffmpeg`
   - **Windows**: Install via scoop/chocolatey or download the binaries and add them to your system path.
4. **Ollama**: For local LLM processing (meeting summary and title generation).
   - Download and install from [ollama.com](https://ollama.com/).
5. **PostgreSQL**: Used for user, settings, and recording data storage.

---

## 📥 Backend Setup

### Option A: Local Go Setup (Traditional)

1. **Download Go Dependencies**:
   Navigate to the `backend` directory and fetch the required Go modules:
   ```bash
   cd backend
   go mod download
   ```

2. **Configure Environment Variables**:
   * Create a file named `.env` in the `backend/` directory:
     ```bash
     touch backend/.env
     ```
   * Add the database connection details, Speechmatics API Key, and a JWT signing secret:
     ```env
     POSTGRES_USER="your_postgres_user"
     POSTGRES_PASSWORD="your_postgres_password"
     POSTGRES_HOST="localhost"
     POSTGRES_PORT="5432"
     POSTGRES_DB="clarifi_db"
     SECRET_KEY="your_jwt_signing_secret"
     ```
   * *Note: Speechmatics API key can be configured directly through the Settings UI once the application is running.*

### Option B: Docker Setup (Recommended for containerized deployment)

1. **Build the Docker Image**:
   Whenever backend source code is created or updated, rebuild the image:
   ```bash
   cd backend
   docker build -t note-taker-backend .
   ```

2. **Run the Container**:
   Pass environment variables and mount a Docker volume at `/app/data` to persist audio recordings and temporary chunks.

   - **Connecting to local PostgreSQL and Ollama on Mac/Host**:
     Use `host.docker.internal` so the container can reach services running on your host machine:
     ```bash
     docker run -p 8000:8000 \
       -e POSTGRES_HOST=host.docker.internal \
       -e OLLAMA_URL=http://host.docker.internal:11434 \
       --env-file .env \
       -v note_taker_data:/app/data \
       note-taker-backend
     ```

   - **Custom Host Port (e.g. `8084`)**:
     ```bash
     docker run -p 8084:8000 \
       -e POSTGRES_HOST=host.docker.internal \
       -e OLLAMA_URL=http://host.docker.internal:11434 \
       --env-file .env \
       -v note_taker_data:/app/data \
       note-taker-backend
     ```
     > [!IMPORTANT]
     > If host port `8084` is used, ensure `frontend/vite.config.js` proxy target points to `http://localhost:8084` or set `VITE_API_BASE_URL=http://localhost:8084` in `frontend/.env`.

   *(See [RUNNING_LOCALLY.md](RUNNING_LOCALLY.md) for full detailed guide).*

---

## 🎨 Frontend Setup

1. **Install Dependencies**:
   Navigate to the `frontend` directory and install the Node modules:
   ```bash
   cd frontend
   npm install
   ```

---

## 🧠 Ollama Setup

The backend connects to Ollama for generating summaries and action items.

1. **Start Ollama**:
   Ensure the Ollama application is running.
2. **Pull the Model**:
   By default, the backend expects the model `gemma4:12b-mlx` (defined in settings). If you wish to use a different model, pull it first:
   ```bash
   ollama pull gemma4:12b-mlx
   ```
   > [!NOTE]
   > You can modify the target Ollama model dynamically via the Settings page in the UI or update it in the database.

