# Note-Taker AI Frontend

Modern web and mobile interface for the **Note-Taker AI** application, built with **Vue 3**, **Vite**, **PrimeVue**, and **Capacitor iOS**.

---

## 🚀 Features

- **Apple-Inspired Design**: Clean macOS/iOS design system with grouped inset settings, cards, and smooth micro-interactions.
- **AI Model & Provider Configuration**:
  - Switch seamlessly between **Ollama (Local)** and **Google Gemini (Cloud)** in the System Settings page.
  - **Database-Driven Settings**: Configure Ollama server URL (`http://localhost:11434`), Gemini endpoint proxy, and API keys directly in the UI without maintaining extensive backend `.env` variables.
  - Dynamic model selection:
    - **Ollama**: Dynamically lists and refreshes all installed local models from the configured Ollama URL.
    - **Google Gemini**: Choose between `gemini-4-argon`, `gemini-3.8-flash`, and `gemini-3.5-flash-lite`.
  - Secure **Gemini API Key** and **Speechmatics API Key** management with show/hide password toggles.
- **Meeting Transcription & Playback**: Real-time progress, audio player, interactive speaker diarization segments, and markdown summary viewer.
- **Offline & iOS Support**: Native audio recording and offline synchronization queue via Capacitor.

---

## 🛠️ Development Setup

1. **Install Dependencies**:
   ```bash
   cd frontend
   npm install
   ```

2. **Start Dev Server**:
   ```bash
   npm run dev
   ```
   The application runs by default at [http://localhost:5173](http://localhost:5173).

3. **Build for Production**:
   ```bash
   npm run build
   ```

4. **Sync with iOS (Capacitor)**:
   ```bash
   npx cap sync ios
   npx cap open ios
   ```
