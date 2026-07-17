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

### Option B: Docker Setup (Recommended for simple deployment)

You can build a Docker image for the backend. The Docker image automatically bundles the compiled Go binary along with its external tool dependencies (like `ffmpeg`).

1. **Build the Docker Image**:
   Navigate to the `backend` directory and run:
   ```bash
   cd backend
   docker build -t note-taker-backend .
   ```

2. **Prepare Environment Variables**:
   You can pass the same environment variables defined in the `.env` file when running the container (see the [Running Locally Guide](RUNNING_LOCALLY.md) for details).

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

