import Foundation
import Capacitor
import AVFoundation

@objc(NativeAudioRecorderPlugin)
public class NativeAudioRecorderPlugin: CAPPlugin, CAPBridgedPlugin {
    public let identifier = "NativeAudioRecorderPlugin"
    public let jsName = "NativeAudioRecorder"
    public let pluginMethods: [CAPPluginMethod] = [
        CAPPluginMethod(name: "checkPermissions", returnType: CAPPluginReturnPromise),
        CAPPluginMethod(name: "requestPermissions", returnType: CAPPluginReturnPromise),
        CAPPluginMethod(name: "startRecording", returnType: CAPPluginReturnPromise),
        CAPPluginMethod(name: "pauseRecording", returnType: CAPPluginReturnPromise),
        CAPPluginMethod(name: "resumeRecording", returnType: CAPPluginReturnPromise),
        CAPPluginMethod(name: "stopRecording", returnType: CAPPluginReturnPromise),
        CAPPluginMethod(name: "getStatus", returnType: CAPPluginReturnPromise),
        CAPPluginMethod(name: "readChunk", returnType: CAPPluginReturnPromise),
        CAPPluginMethod(name: "readEntireFileBase64", returnType: CAPPluginReturnPromise),
        CAPPluginMethod(name: "deleteAudioFile", returnType: CAPPluginReturnPromise),
        CAPPluginMethod(name: "getRecordedFiles", returnType: CAPPluginReturnPromise)
    ]

    private var audioRecorder: AVAudioRecorder?
    private var currentRecordingUrl: URL?
    private var currentRecordingId: String?
    private var recordingStartTime: Date?
    private var accumulatedDuration: TimeInterval = 0
    private var isPaused: Bool = false

    private var recordingsDirectory: URL {
        let paths = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask)
        let dir = paths[0].appendingPathComponent("OfflineRecordings", isDirectory: true)
        if !FileManager.default.fileExists(atPath: dir.path) {
            try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true, attributes: nil)
        }
        return dir
    }

    // MARK: - Permissions
    @objc override public func checkPermissions(_ call: CAPPluginCall) {
        let status = AVAudioSession.sharedInstance().recordPermission
        switch status {
        case .granted:
            call.resolve(["record": "granted"])
        case .denied:
            call.resolve(["record": "denied"])
        case .undetermined:
            call.resolve(["record": "prompt"])
        @unknown default:
            call.resolve(["record": "prompt"])
        }
    }

    @objc override public func requestPermissions(_ call: CAPPluginCall) {
        let session = AVAudioSession.sharedInstance()
        session.requestRecordPermission { granted in
            call.resolve(["record": granted ? "granted" : "denied"])
        }
    }

    // MARK: - Recording Operations
    @objc public func startRecording(_ call: CAPPluginCall) {
        let session = AVAudioSession.sharedInstance()
        do {
            // Configure audio session for background recording & speaker playback
            try session.setCategory(
                .playAndRecord,
                mode: .default,
                options: [.defaultToSpeaker, .allowBluetooth, .allowBluetoothA2DP]
            )
            try session.setActive(true, options: .notifyOthersOnDeactivation)
        } catch {
            call.reject("Failed to configure audio session: \(error.localizedDescription)")
            return
        }

        let recordingId = call.getString("id") ?? UUID().uuidString
        let filename = "rec_\(recordingId).m4a"
        let fileUrl = recordingsDirectory.appendingPathComponent(filename)

        let settings: [String: Any] = [
            AVFormatIDKey: Int(kAudioFormatMPEG4AAC),
            AVSampleRateKey: 44100.0,
            AVNumberOfChannelsKey: 1,
            AVEncoderAudioQualityKey: AVAudioQuality.high.rawValue,
            AVEncoderBitRateKey: 64000
        ]

        do {
            audioRecorder = try AVAudioRecorder(url: fileUrl, settings: settings)
            audioRecorder?.isMeteringEnabled = true
            guard let recorder = audioRecorder, recorder.record() else {
                call.reject("Failed to start audio recording engine")
                return
            }

            self.currentRecordingUrl = fileUrl
            self.currentRecordingId = recordingId
            self.recordingStartTime = Date()
            self.accumulatedDuration = 0
            self.isPaused = false

            call.resolve([
                "recordingId": recordingId,
                "filePath": fileUrl.path,
                "status": "recording"
            ])
        } catch {
            call.reject("Could not create audio recorder: \(error.localizedDescription)")
        }
    }

    @objc public func pauseRecording(_ call: CAPPluginCall) {
        guard let recorder = audioRecorder, recorder.isRecording else {
            call.reject("Not currently recording")
            return
        }

        recorder.pause()
        if let start = recordingStartTime {
            accumulatedDuration += Date().timeIntervalSince(start)
            recordingStartTime = nil
        }
        isPaused = true

        call.resolve([
            "status": "paused",
            "duration": accumulatedDuration
        ])
    }

    @objc public func resumeRecording(_ call: CAPPluginCall) {
        guard let recorder = audioRecorder, isPaused else {
            call.reject("Not currently paused")
            return
        }

        recorder.record()
        recordingStartTime = Date()
        isPaused = false

        call.resolve([
            "status": "recording",
            "duration": accumulatedDuration
        ])
    }

    @objc public func stopRecording(_ call: CAPPluginCall) {
        guard let recorder = audioRecorder else {
            call.reject("No active recording found to stop")
            return
        }

        if let start = recordingStartTime {
            accumulatedDuration += Date().timeIntervalSince(start)
        }

        let finalDuration = accumulatedDuration
        let recordingId = currentRecordingId ?? ""
        let fileUrl = currentRecordingUrl ?? recorder.url

        recorder.stop()
        self.audioRecorder = nil
        self.recordingStartTime = nil
        self.isPaused = false

        // Deactivate audio session to allow other system sounds to resume
        let session = AVAudioSession.sharedInstance()
        try? session.setActive(false, options: .notifyOthersOnDeactivation)

        var fileSize: Int64 = 0
        if let attrs = try? FileManager.default.attributesOfItem(atPath: fileUrl.path) {
            fileSize = (attrs[.size] as? Int64) ?? 0
        }

        call.resolve([
            "recordingId": recordingId,
            "filePath": fileUrl.path,
            "filename": fileUrl.lastPathComponent,
            "duration": finalDuration,
            "fileSize": fileSize,
            "mimeType": "audio/m4a"
        ])
    }

    @objc public func getStatus(_ call: CAPPluginCall) {
        guard let recorder = audioRecorder else {
            call.resolve([
                "isRecording": false,
                "isPaused": false,
                "duration": 0,
                "meterLevel": 0
            ])
            return
        }

        var currentDuration = accumulatedDuration
        if let start = recordingStartTime {
            currentDuration += Date().timeIntervalSince(start)
        }

        recorder.updateMeters()
        let power = recorder.averagePower(forChannel: 0)
        // Normalize -60dB .. 0dB to 0.0 .. 1.0
        let normalizedLevel = max(0.0, min(1.0, (power + 60.0) / 60.0))

        call.resolve([
            "isRecording": recorder.isRecording,
            "isPaused": isPaused,
            "duration": currentDuration,
            "meterLevel": normalizedLevel
        ])
    }

    // MARK: - Chunked File Reading for Uploads
    @objc public func readChunk(_ call: CAPPluginCall) {
        guard let filePath = call.getString("filePath"),
              let offset = call.getInt("offset"),
              let length = call.getInt("length") else {
            call.reject("Missing required parameters: filePath, offset, length")
            return
        }

        let url = URL(fileURLWithPath: filePath)
        do {
            let fileHandle = try FileHandle(forReadingFrom: url)
            defer { try? fileHandle.close() }

            try fileHandle.seek(toOffset: UInt64(offset))
            let data = fileHandle.readData(ofLength: length)
            let base64 = data.base64EncodedString()

            call.resolve([
                "base64": base64,
                "bytesRead": data.count
            ])
        } catch {
            call.reject("Failed reading chunk: \(error.localizedDescription)")
        }
    }

    @objc public func readEntireFileBase64(_ call: CAPPluginCall) {
        guard let filePath = call.getString("filePath") else {
            call.reject("filePath is required")
            return
        }

        let url = URL(fileURLWithPath: filePath)
        do {
            let data = try Data(contentsOf: url)
            call.resolve([
                "base64": data.base64EncodedString(),
                "fileSize": data.count
            ])
        } catch {
            call.reject("Failed reading file: \(error.localizedDescription)")
        }
    }

    @objc public func deleteAudioFile(_ call: CAPPluginCall) {
        guard let filePath = call.getString("filePath") else {
            call.reject("filePath is required")
            return
        }

        let url = URL(fileURLWithPath: filePath)
        do {
            if FileManager.default.fileExists(atPath: url.path) {
                try FileManager.default.removeItem(at: url)
            }
            call.resolve(["deleted": true])
        } catch {
            call.reject("Failed to delete file: \(error.localizedDescription)")
        }
    }

    @objc public func getRecordedFiles(_ call: CAPPluginCall) {
        let dir = recordingsDirectory
        do {
            let fileUrls = try FileManager.default.contentsOfDirectory(at: dir, includingPropertiesForKeys: [.fileSizeKey, .creationDateKey], options: .skipsHiddenFiles)
            let files = fileUrls.map { url -> [String: Any] in
                var size: Int64 = 0
                var createdAt: TimeInterval = 0
                if let resourceValues = try? url.resourceValues(forKeys: [.fileSizeKey, .creationDateKey]) {
                    size = Int64(resourceValues.fileSize ?? 0)
                    createdAt = resourceValues.creationDate?.timeIntervalSince1970 ?? 0
                }
                return [
                    "name": url.lastPathComponent,
                    "path": url.path,
                    "size": size,
                    "createdAt": createdAt
                ]
            }
            call.resolve(["files": files])
        } catch {
            call.reject("Failed to list files: \(error.localizedDescription)")
        }
    }
}
