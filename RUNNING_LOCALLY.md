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

### 4. Installing and Running on an iOS Phone (Physical iPhone or Simulator)

The iOS application provides the exact same Vue 3 UI and Apple design system, backed by native iOS background audio capabilities (`AVAudioRecorder` + `AVAudioSession` + `UIBackgroundModes: ["audio"]`) and an offline-first resilient sync queue.

#### 📋 Prerequisites
- **Mac** with **Xcode** installed (available free from the Mac App Store).
- A physical **iPhone** (iOS 15.0 or newer) and a USB-C or Lightning cable.
- A standard **personal Apple ID** (a paid Apple Developer account is **not** required).

---

#### 📱 Step-by-Step Installation on a Physical iPhone

##### 1. Connect Your iPhone to Your Mac
1. Plug your iPhone into your Mac using a USB cable.
2. If prompted on your iPhone, tap **Trust This Computer** and enter your passcode.
3. Open **Finder** on your Mac, select your iPhone in the sidebar, and confirm it is connected.

##### 2. Enable Developer Mode on Your iPhone (iOS 16, 17, & 18+)
Apple requires Developer Mode to run sideloaded apps from Xcode:
1. On your iPhone, open **Settings**.
2. Tap **Privacy & Security**.
3. Scroll all the way to the bottom and tap **Developer Mode**.
4. Toggle the switch to **ON**, then tap **Restart** when prompted.
5. Once your iPhone reboots and you unlock it, tap **Turn On** on the alert and enter your passcode.

##### 3. Add Your Free Apple ID to Xcode
1. Launch **Xcode**.
2. In the top macOS menu bar, click **Xcode** ➡️ **Settings...** (or **Preferences...** on older versions).
3. Select the **Accounts** tab.
4. Click the **`+`** button in the lower-left corner ➡️ select **Apple ID** ➡️ click **Continue**.
5. Enter your personal Apple ID and password. A "Personal Team" will automatically appear.

##### 4. Build Web Assets and Open the Xcode Project
In your Mac terminal:
```bash
cd /Users/raenard/Documents/note-taker/frontend
npm run build
npx cap sync ios
npx cap open ios
```
This opens `App.xcworkspace` / `App.xcodeproj` in Xcode.

##### 5. Configure Signing in Xcode
1. In the left navigation pane of Xcode, click the top-level **App** project (blue icon).
2. In the center editor, select the **App** target under *Targets*.
3. Click the **Signing & Capabilities** tab.
4. Check **Automatically manage signing**.
5. In the **Team** dropdown, select your **Personal Team** (your name).
6. In **Bundle Identifier**, if you see a red error saying the identifier is unavailable:
   - Change `com.notetaker.ai` to a unique name, e.g., `com.<yourname>.notetaker` (e.g., `com.raenard.notetaker`).

##### 6. Select Your iPhone and Install
1. In Xcode's top toolbar, click the device selector (next to the Play ▶ and Stop ■ buttons).
2. Select your physical **iPhone** (listed under *iOS Device*), not a simulator.
3. Click the **Play / Run ▶** button (or press `⌘R`).
4. Xcode will compile the native app, transfer it, and install it on your iPhone.

##### 7. Trust Developer Certificate on Your iPhone (First Time Only)
When you first attempt to open the app on your phone, iOS may display an *"Untrusted Developer"* dialog. To authorize it:
1. On your iPhone, open **Settings**.
2. Tap **General** ➡️ **VPN & Device Management**.
3. Under **Developer App**, tap your Apple ID email.
4. Tap **Trust "[your Apple ID email]"** and confirm by tapping **Trust**.
5. Open the **Note-Taker AI** app from your iPhone's home screen!

---

#### 🌐 Connecting Your iPhone to Your Local Mac Backend

To let the iPhone communicate with the Go backend running on your Mac:

1. **Connect both devices to the same Wi-Fi network**.
2. **Find your Mac's Local IP Address**:
   Run in your terminal:
   ```bash
   ipconfig getifaddr en0
   ```
   *(e.g., `192.168.3.111`)*
3. **Configure the App**:
   - Open **Note-Taker AI** on your iPhone.
   - Go to **System Settings** in the app.
   - Under **Server & Offline Sync**, enter your backend URL:
     ```text
     http://<YOUR_MAC_IP>:8000
     ```
     *(Example: `http://192.168.3.111:8000`)*
   - Tap **Test** to verify connection, then tap **Save Settings**.

---

#### 🧪 Testing Offline & Background Recording

- **Lock Screen / Background Mode**:
  Start recording a live meeting in the app, then lock your iPhone screen or switch to another app (e.g. Safari or Messages). Notice the audio recording continues smoothly without interruptions.
- **Offline First**:
  Turn on **Airplane Mode** on your iPhone. Record an audio memo and press Stop. The audio is instantly saved locally in the **Offline Queue** on your device.
- **Automatic Sync**:
  Turn off Airplane Mode. As soon as your connection returns, the app automatically streams the chunks to your backend and begins Speechmatics transcription.

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
