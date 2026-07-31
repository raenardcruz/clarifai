# Running Locally

This document outlines the step-by-step instructions to run the **Note-Taker AI** application locally for development or testing.

---

## 🏃 Step-by-Step Launch Guide

To run the application, you need three separate terminals running (or run them in the background):

### 1. Start Ollama
Ensure the Ollama application is running and the required model is loaded:
```bash
ollama run gemma4:12b-mlx
```
*(If you want to run Ollama in the background, you can close the prompt after verification, as the API remains active at `http://localhost:11434`)*

---

### 2. Start the Backend Server

You can run the backend server either directly using Go or via the built Docker container:

#### Option A: Running Directly with Go
Navigate to the `backend` directory, set your environment credentials in `.env`, and start the Go server:
```bash
cd backend
go run main.go
```

#### Option B: Running with Docker

> [!IMPORTANT]
> **Always rebuild the Docker image after changing backend code!**
> Docker containers run static compiled binaries. If you modify any backend files (like API routes or services), you must rebuild the image first.

1. **Build (or rebuild) the Docker image**:
   ```bash
   cd backend
   docker build -t note-taker-backend .
   ```

2. **Run the container**:
   - **Default Port (`8000`)**:
     ```bash
     docker run -p 8000:8000 \
       -e POSTGRES_HOST=host.docker.internal \
       -e OLLAMA_URL=http://host.docker.internal:11434 \
       --env-file .env \
       -v note_taker_data:/app/data \
       note-taker-backend
     ```
   - **Custom Host Port (e.g., `8084`)**:
     Map host port `8084` to container port `8000` (`-p 8084:8000`). Make sure `PORT` in `.env` is omitted or set to `8000` inside the container:
     ```bash
     docker run -p 8084:8000 \
       -e POSTGRES_HOST=host.docker.internal \
       -e OLLAMA_URL=http://host.docker.internal:11434 \
       --env-file .env \
       -v note_taker_data:/app/data \
       note-taker-backend
     ```
     *(Note: If you run the backend on port `8084`, update `frontend/vite.config.js` proxy target to `http://localhost:8084` or set `VITE_API_BASE_URL=http://localhost:8084` in `frontend/.env`).*

> [!NOTE]
> **Why `host.docker.internal`?** Inside Docker containers, `localhost` refers to the container itself. Passing `-e POSTGRES_HOST=host.docker.internal` and `-e OLLAMA_URL=http://host.docker.internal:11434` allows the container to connect to PostgreSQL and Ollama running on your Mac/host machine.

The backend server will run on [http://localhost:8000](http://localhost:8000) (or your chosen host port).

---

### 3. Start the Frontend Dev Server
Navigate to the `frontend` directory and spin up the Vite development server:

```bash
cd frontend
npm run dev
```
By default, the frontend will run at [http://localhost:5173/](http://localhost:5173/).

---

## 🔍 Verification Checklist

To verify that the application is operating correctly:

1. **Open the Web Interface**: Navigate to [http://localhost:5173/](http://localhost:5173/) in your browser.
2. **Upload a Meeting**: Drag & drop or browse to select an audio file (e.g., `.mp3`, `.wav`, `.m4a`).
3. **Monitor Progress**:
   - The UI will display a new job card with progress bars indicating the status: `transcribing` ➡️ `summarizing` ➡️ `completed`.
   - Monitor the backend console log for step-by-step updates.
4. **Download Results**:
   - Once the status reaches `completed`, click **Download Transcript** to save the transcription.
   - Click **Download Summary** to retrieve the structured notes, action items, and title.
