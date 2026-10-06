import { registerPlugin, Capacitor } from '@capacitor/core'

export const NativeAudioRecorder = registerPlugin('NativeAudioRecorder', {
  web: () => import('./web').then(m => new m.NativeAudioRecorderWeb()),
})

export const isNativePlatform = () => Capacitor.isNativePlatform()
