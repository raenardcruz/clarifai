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
4. **AI Summarization Engine (Ollama or Google Gemini)**:
   - **Ollama**: For local offline LLM processing (meeting summary and title generation). Download and install from [ollama.com](https://ollama.com/).
   - **Google Gemini**: Alternatively, use Google Gemini cloud models (`gemini-4-argon`, `gemini-3.8-flash`, `gemini-3.5-flash-lite`) with a Google Gemini API Key.
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
   * Environment variables are kept to a minimum (only PostgreSQL connection and optional JWT secret). All operational settings (Speechmatics API Key, Gemini API Key, Ollama URL, Ollama/Gemini models, and prompts) are configured in the **Settings** page and saved in the database:
     ```env
     POSTGRES_USER="your_postgres_user"
     POSTGRES_PASSWORD="your_postgres_password"
     POSTGRES_HOST="localhost"
     POSTGRES_PORT="5432"
     POSTGRES_DB="clarifi_db"
     SECRET_KEY="your_jwt_signing_secret"
     ```
   * *(Optional fallback: You can also specify `SPEECHMATICS_API_KEY`, `GEMINI_API_KEY`, or `OLLAMA_URL` in `.env` if you prefer environment-based overrides, but the Settings UI and database take precedence).*

### Option B: Docker Setup (Recommended for containerized deployment)

1. **Build the Docker Image**:
   Whenever backend source code is created or updated, rebuild the image:
   ```bash
   cd backend
   docker build -t note-taker-backend .
   ```

2. **Run the Container**:
   Pass environment variables and mount a Docker volume at `/app/data` to persist audio recordings and temporary chunks.

   - **Connecting to local PostgreSQL on Mac/Host**:
     Use `host.docker.internal` so the container can reach services running on your host machine:
     ```bash
     docker run -p 8000:8000 \
       -e POSTGRES_HOST=host.docker.internal \
       --env-file .env \
       -v note_taker_data:/app/data \
       note-taker-backend
     ```
     *(Note: If using Ollama with Docker on Mac, configure the Ollama Server URL as `http://host.docker.internal:11434` directly in the web Settings page!)*

   - **Custom Host Port (e.g. `8084`)**:
     ```bash
     docker run -p 8084:8000 \
       -e POSTGRES_HOST=host.docker.internal \
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

## 🧠 AI Provider Setup (Ollama & Google Gemini)

The application supports both local Ollama models and Google Gemini cloud models for generating titles, meeting summaries, SOAP clinical notes, and speaker diarization names. You can choose the active provider in the **Settings** page.

### Option 1: Ollama (Local)
1. **Start Ollama**: Ensure the Ollama application is running.
2. **Pull the Model**: By default, the local model is set to `gemma4:12b-mlx`. Pull it using:
   ```bash
   ollama pull gemma4:12b-mlx
   ```
   > [!NOTE]
   > You can select any installed Ollama model dynamically via the Settings page dropdown in the UI.

### Option 2: Google Gemini (Cloud)
1. **Get an API Key**: Obtain a Gemini API key from [Google AI Studio](https://aistudio.google.com/).
2. **Configure in Settings**:
   - Go to **System Settings** in the web interface.
   - Set **AI Provider** to **Google Gemini**.
   - Choose your preferred Gemini model (`gemini-4-argon`, `gemini-3.8-flash`, or `gemini-3.5-flash-lite`).
   - Enter your **Gemini API Key** and click **Save Settings** (or supply `GEMINI_API_KEY` in `backend/.env`).

